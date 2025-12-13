package saver

import (
	"context"

	v1 "github.com/w-h-a/golens/api/event/v1"
)

type V1Saver interface {
	Save(ctx context.Context, event *v1.Event, opts ...SaveOption) error
}
