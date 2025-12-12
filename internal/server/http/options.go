package http

import (
	"context"

	"github.com/w-h-a/golens/internal/server"
)

type middlewareKey struct{}

func WithMiddleware(ms ...Middleware) server.Option {
	return func(o *server.Options) {
		o.Context = context.WithValue(o.Context, middlewareKey{}, ms)
	}
}

func getMiddlewareFromCtx(ctx context.Context) ([]Middleware, bool) {
	ms, ok := ctx.Value(middlewareKey{}).([]Middleware)
	return ms, ok
}
