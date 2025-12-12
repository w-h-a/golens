package v1

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/url"

	v1 "github.com/w-h-a/golens/api/dto/v1"
	"github.com/w-h-a/golens/internal/client/sender"
)

type v1Sender struct {
	options sender.Options
	client  *http.Client
}

func (s *v1Sender) Send(ctx context.Context, req *v1.Request, opts ...sender.SendOption) (*v1.Response, error) {
	targetURL, err := url.JoinPath(s.options.BaseURL, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create target URL: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, req.Body)
	if err != nil {
		return nil, err
	}

	for k, vv := range req.Headers {
		for _, v := range vv {
			httpReq.Header.Add(k, v)
		}
	}

	httpRsp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	rspHeaders := map[string][]string{}
	maps.Copy(rspHeaders, httpRsp.Header)

	return &v1.Response{
		StatusCode: httpRsp.StatusCode,
		Headers:    rspHeaders,
		Body:       httpRsp.Body,
	}, nil
}

func NewSender(opts ...sender.Option) sender.V1Sender {
	options := sender.NewOptions(opts...)

	// TODO: validate options

	s := &v1Sender{
		options: options,
		client:  &http.Client{},
	}

	return s
}
