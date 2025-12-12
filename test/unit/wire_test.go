package unit

import (
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
	"github.com/w-h-a/golens/internal/client/sender/mock"
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
	mockStream := `data: {"model":"gpt-4","choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":" World"}}]}

data: [DONE]
`
	sender := mock.NewSender(
		mock.WithRspBody(mockStream),
	)

	wire := wire.New(sender)

	req := &v1dto.Request{
		Path: "/v1/chat",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Act
	rsp, err := wire.Tap(context.Background(), req, func() { wg.Done() })
	require.NoError(t, err)

	_, err = io.ReadAll(rsp.Body)
	require.NoError(t, err)

	err = rsp.Body.Close()
	require.NoError(t, err)

	// Assert
	wg.Wait()
}
