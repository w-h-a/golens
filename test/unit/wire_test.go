package unit

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1dto "github.com/w-h-a/golens/api/dto/v1"
	v1event "github.com/w-h-a/golens/api/event/v1"
	mocksaver "github.com/w-h-a/golens/internal/client/saver/mock"
	mocksender "github.com/w-h-a/golens/internal/client/sender/mock"
	"github.com/w-h-a/golens/internal/service/wire"
)

func TestProcessStream(t *testing.T) {
	// Arrange
	wire := &wire.Wire{}
	event := &v1event.Event{StartTime: time.Now(), Model: "unknown"}

	input := `data: {"model":"gpt-4","choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":" World"}}]}

data: [DONE]`

	reader := strings.NewReader(input)

	// Act
	wire.ProcessStream(context.Background(), reader, event)

	// Assert
	assert.Equal(t, "gpt-4", event.Model)
	assert.Equal(t, "Hello World", event.Response)
	assert.Equal(t, 2, event.TokenCount)
}

func TestTap(t *testing.T) {
	// Arrange
	dirty := map[string][]string{}

	dirty["Authorization"] = []string{"Bearer fake-token"}
	dirty["Golens-Attribute-User-Id"] = []string{"user-123"}
	dirty["Content-Type"] = []string{"application/json"}

	inputBody := `{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}`

	mockStream := `data: {"model":"gpt-4","choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":" World"}}]}

data: [DONE]
`
	sender := mocksender.NewSender(
		mocksender.WithRspBody(mockStream),
	)

	saver := mocksaver.NewSaver()

	wire := wire.New(sender, saver)

	req := &v1dto.Request{
		Path:    "/v1/chat",
		Headers: dirty,
		Body:    io.NopCloser(bytes.NewBufferString(inputBody)),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Act
	rsp, err := wire.Tap(context.Background(), req, func() { wg.Done() })
	require.NoError(t, err)

	bs, err := io.ReadAll(rsp.Body)
	require.NoError(t, err)

	err = rsp.Body.Close()
	require.NoError(t, err)

	wg.Wait()

	// Assert
	assert.Contains(t, string(bs), "Hello")
	assert.Contains(t, string(bs), "World")
	assert.Equal(t, 1, saver.Count())
	assert.NotNil(t, saver.Captured())
	assert.Equal(t, "gpt-4", saver.Captured().Model)
	assert.Equal(t, "Hello World", saver.Captured().Response)
	assert.Equal(t, 2, saver.Captured().TokenCount)

	assert.NotNil(t, saver.Captured().Request)
	assert.JSONEq(t, inputBody, string(saver.Captured().Request))

	assert.NotNil(t, saver.Captured().Attributes)
	assert.Equal(t, "user-123", saver.Captured().Attributes["User-Id"])
	assert.NotContains(t, saver.Captured().Attributes, "Authorization")
	assert.NotContains(t, saver.Captured().Attributes, "Content-Type")

	capturedBody, _ := io.ReadAll(sender.Captured().Body)
	assert.JSONEq(t, inputBody, string(capturedBody))

	assert.NotNil(t, sender.Captured().Headers)
	assert.NotContains(t, sender.Captured().Headers, "Golens-Attribute-User-Id")
	assert.Equal(t, "Bearer fake-token", sender.Captured().Headers["Authorization"][0])
	assert.Equal(t, "application/json", sender.Captured().Headers["Content-Type"][0])
}
