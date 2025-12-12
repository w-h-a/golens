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
	"github.com/w-h-a/golens/internal/client/sender"
)

// TODO: make configurable
const (
	maxSize = 10 * 1024
)

type Wire struct {
	sender    sender.V1Sender
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
	// TODO: try to extract via W3C trace standard
	event := &v1event.Event{
		TraceID:   "trace-" + fmt.Sprint(time.Now().UnixNano()),
		StartTime: time.Now(),
		Model:     "unknown",
	}

	rsp, err := w.sender.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	event.Status = rsp.StatusCode

	// (see below for main highway)
	// off-ramp whenever the tee reader pulls from LLM
	// it will write to pw
	// pr will read from pw
	// build the off-ramp first
	pr, pw := io.Pipe()

	// main highway LLM -> tee -> client/agent
	// but tee will divert to the off-ramp pw -> pr
	tee := io.TeeReader(rsp.Body, pw)

	go func() {
		defer pr.Close()
		w.ProcessStream(ctx, pr, event)

		// temporary
		log.Printf("[Event] Trace: %s | Status: %d | Tokens: %d | Model: %s | Dur: %dms",
			event.TraceID, event.Status, event.TokenCount, event.Model, event.DurationMs)
	}()

	wrappedBody := &pipeBody{
		Reader:       tee,
		originalBody: rsp.Body,
		pw:           pw,
		onFinish:     onDone,
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
	event.DurationMs = time.Since(event.StartTime).Milliseconds()
}

func New(s sender.V1Sender) *Wire {
	return &Wire{
		sender:    s,
		isRunning: false,
		mtx:       sync.RWMutex{},
	}
}
