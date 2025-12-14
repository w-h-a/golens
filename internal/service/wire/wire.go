package wire

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	v1dto "github.com/w-h-a/golens/api/dto/v1"
	v1event "github.com/w-h-a/golens/api/event/v1"
	"github.com/w-h-a/golens/internal/client/saver"
	"github.com/w-h-a/golens/internal/client/sender"
	"github.com/w-h-a/golens/internal/util"
)

// TODO: make configurable
const (
	maxSize = 10 * 1024
)

type Wire struct {
	sender    sender.V1Sender
	saver     saver.V1Saver
	isRunning bool
	mtx       sync.RWMutex
}

func (w *Wire) Run(stop chan struct{}) error {
	w.mtx.RLock()
	if w.isRunning {
		w.mtx.RUnlock()
		return errors.New("proxy already running")
	}
	w.mtx.RUnlock()

	if err := w.Start(); err != nil {
		return fmt.Errorf("failed to start user service: %w", err)
	}

	<-stop

	return w.Stop()
}

func (w *Wire) Start() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if w.isRunning {
		return errors.New("proxy already started")
	}

	w.isRunning = true

	return nil
}

func (w *Wire) Stop() error {
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	return w.stop(stopCtx)
}

func (w *Wire) stop(ctx context.Context) error {
	w.mtx.Lock()

	if !w.isRunning {
		w.mtx.Unlock()
		return errors.New("user service not running")
	}

	w.isRunning = false

	w.mtx.Unlock()

	gracefulStopDone := make(chan struct{})
	go func() {
		// TODO: close clients gracefully
		close(gracefulStopDone)
	}()

	var stopErr error

	select {
	case <-gracefulStopDone:
	case <-ctx.Done():
		stopErr = ctx.Err()
	}

	return stopErr
}

func (w *Wire) Tap(ctx context.Context, req *v1dto.Request, onDone func()) (*v1dto.Response, error) {
	traceId, _ := util.TraceIdFrom(ctx)

	bs := []byte{}
	if req.Body != nil {
		originalBody := req.Body
		defer originalBody.Close()

		body, err := io.ReadAll(originalBody)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}

		bs = body
		req.Body = io.NopCloser(bytes.NewBuffer(bs))
	}

	attributes, clean := extractAndCleanHeaders(req.Headers)

	req.Headers = clean

	event := &v1event.Event{
		TraceId:    traceId,
		StartTime:  time.Now(),
		Model:      "unknown",
		Request:    json.RawMessage(bs),
		Attributes: attributes,
	}

	rsp, err := w.sender.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	event.StatusCode = rsp.StatusCode

	pr, pw := io.Pipe()
	tee := io.TeeReader(rsp.Body, pw)

	go func() {
		if onDone != nil {
			defer onDone()
		}

		defer pr.Close()

		w.ProcessStream(ctx, pr, event)

		// create a detached context so if the user cancels, the db save still happens.
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		event.EndTime = time.Now()
		event.DurationMs = event.EndTime.Sub(event.StartTime).Milliseconds()

		if err := w.saver.Save(saveCtx, event); err != nil {
			log.Printf("[Wire] failed to save event: %v", err)
		} else {
			log.Printf("[Wire] saved log trace=%s model=%s tokens=%d", event.TraceId, event.Model, event.TokenCount)
		}
	}()

	wrappedBody := &pipeBody{
		Reader:       tee,
		originalBody: rsp.Body,
		pw:           pw,
	}

	return &v1dto.Response{
		StatusCode: rsp.StatusCode,
		Headers:    rsp.Headers,
		Body:       wrappedBody,
	}, nil
}

func (w *Wire) ProcessStream(ctx context.Context, r io.Reader, event *v1event.Event) {
	scanner := bufio.NewScanner(r)
	stringsBuilder := strings.Builder{}
	currentSize := 0

	for scanner.Scan() {
		line := scanner.Bytes()

		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		payload := bytes.TrimPrefix(line, []byte("data: "))
		if string(payload) == "[DONE]" {
			break
		}

		var chunk struct {
			Model   string `json:"model"`
			Choices []struct {
				Delta *struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(payload, &chunk); err == nil {
			if len(chunk.Model) > 0 && event.Model == "unknown" {
				event.Model = chunk.Model
			}

			if len(chunk.Choices) == 0 || chunk.Choices[0].Delta == nil {
				continue
			}

			content := chunk.Choices[0].Delta.Content

			if currentSize < maxSize {
				if currentSize+len(content) > maxSize {
					stringsBuilder.WriteString("... [TRUNCATED]")
					currentSize = maxSize
				} else {
					stringsBuilder.WriteString(content)
					currentSize += len(content)
				}
			}

			// TODO: figure this out for real
			if len(content) > 0 {
				event.TokenCount++
			}
		}
	}

	event.Response = stringsBuilder.String()
}

func New(sender sender.V1Sender, saver saver.V1Saver) *Wire {
	return &Wire{
		sender:    sender,
		saver:     saver,
		isRunning: false,
		mtx:       sync.RWMutex{},
	}
}
