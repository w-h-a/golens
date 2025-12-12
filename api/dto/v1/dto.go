package v1

import "io"

type Request struct {
	Method  string
	Path    string
	Headers map[string][]string
	Body    io.ReadCloser
}

type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       io.ReadCloser
}
