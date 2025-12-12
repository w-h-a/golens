package mock

import (
	"context"

	"github.com/w-h-a/golens/internal/client/sender"
)

type rspBodyKey struct{}

func WithRspBody(rsp string) sender.Option {
	return func(o *sender.Options) {
		o.Context = context.WithValue(o.Context, rspBodyKey{}, rsp)
	}
}

func getRspBodyFromCtx(ctx context.Context) (string, bool) {
	rsp, ok := ctx.Value(rspBodyKey{}).(string)
	return rsp, ok
}
