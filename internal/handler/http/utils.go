package http

import (
	"context"
	"net/http"
	"strings"
)

func ReqToCtx(r *http.Request) context.Context {
	ctx := r.Context()

	for k, v := range r.Header {
		ctx = context.WithValue(ctx, strings.ToLower(k), v[0])
	}

	return ctx
}
