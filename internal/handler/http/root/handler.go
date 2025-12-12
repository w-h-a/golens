package root

import (
	"io"
	"log"
	"net/http"

	v1 "github.com/w-h-a/golens/api/dto/v1"
	httphandler "github.com/w-h-a/golens/internal/handler/http"
	"github.com/w-h-a/golens/internal/service/wire"
	"github.com/w-h-a/golens/internal/util"
)

type rootHandler struct {
	wire *wire.Wire
}

func (h *rootHandler) Handle(w http.ResponseWriter, r *http.Request) {
	traceId := httphandler.GetTraceId(r)

	ctx := util.WithTraceID(r.Context(), traceId)

	req := &v1.Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header,
		Body:    r.Body,
	}

	rsp, err := h.wire.Tap(ctx, req, nil)
	if err != nil {
		log.Printf("Proxy Error: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer rsp.Body.Close()

	for k, vv := range rsp.Headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(rsp.StatusCode)

	if _, err := io.Copy(w, rsp.Body); err != nil {
		log.Printf("Streaming error: %v", err)
	}
}

func New(w *wire.Wire) *rootHandler {
	return &rootHandler{
		wire: w,
	}
}
