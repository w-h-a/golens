package wire

import "io"

type pipeBody struct {
	io.Reader
	originalBody io.Closer
	pw           *io.PipeWriter
}

func (b *pipeBody) Read(p []byte) (n int, err error) {
	n, err = b.Reader.Read(p)
	if err == io.EOF {
		b.pw.Close()
	}
	return n, err
}

func (b *pipeBody) Close() error {
	b.pw.Close()
	return b.originalBody.Close()
}
