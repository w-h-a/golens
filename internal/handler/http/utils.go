package http

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func GetTraceId(r *http.Request) string {
	if id := r.Header.Get("traceparent"); len(id) > 0 {
		return id
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	return fmt.Sprintf("00-%016x%016x-%016x-01", rnd.Uint64(), rnd.Uint64(), rnd.Uint64())
}
