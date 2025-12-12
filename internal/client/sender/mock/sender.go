package mock

import (
	"context"
	"io"
	"strings"

	v1 "github.com/w-h-a/golens/api/dto/v1"
	"github.com/w-h-a/golens/internal/client/sender"
)

type mockV1Sender struct {
	options sender.Options
	rspBody string
}

func (s *mockV1Sender) Send(ctx context.Context, req *v1.Request, opts ...sender.SendOption) (*v1.Response, error) {
	return &v1.Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(s.rspBody)),
	}, nil
}

func NewSender(opts ...sender.Option) sender.V1Sender {
	options := sender.NewOptions(opts...)

	s := &mockV1Sender{
		options: options,
	}

	if rsp, ok := getRspBodyFromCtx(options.Context); ok {
		s.rspBody = rsp
	}

	return s
}
