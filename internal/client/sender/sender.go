package sender

import (
	"context"

	v1 "github.com/w-h-a/golens/api/dto/v1"
)

type V1Sender interface {
	Send(ctx context.Context, req *v1.Request, opts ...SendOption) (*v1.Response, error)
}
