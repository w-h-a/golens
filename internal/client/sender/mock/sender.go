package mock

import (
	"context"
	"io"
	"strings"
	"sync"

	v1 "github.com/w-h-a/golens/api/dto/v1"
	"github.com/w-h-a/golens/internal/client/sender"
)

type mockV1Sender struct {
	options  sender.Options
	rspBody  string
	captured *v1.Request
	mtx      sync.RWMutex
}

func (s *mockV1Sender) Send(ctx context.Context, req *v1.Request, opts ...sender.SendOption) (*v1.Response, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.captured = req

	return &v1.Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(s.rspBody)),
	}, nil
}

func (s *mockV1Sender) Captured() *v1.Request {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.captured
}

func NewSender(opts ...sender.Option) *mockV1Sender {
	options := sender.NewOptions(opts...)

	s := &mockV1Sender{
		options: options,
		mtx:     sync.RWMutex{},
	}

	if rsp, ok := RspBodyFrom(options.Context); ok {
		s.rspBody = rsp
	}

	return s
}
