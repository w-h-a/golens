package noop

import (
	"context"
	"log"

	v1 "github.com/w-h-a/golens/api/event/v1"
	"github.com/w-h-a/golens/internal/client/saver"
)

type noopV1Saver struct {
	options saver.Options
}

func (s *noopV1Saver) Save(ctx context.Context, event *v1.Event, opts ...saver.SaveOption) error {
	log.Printf("--- [DB PERSISTENCE] ---\nTrace: %s\nModel: %s\nTokens: %d\nResponse: %s...\n-----------------------------",
		event.TraceId, event.Model, event.TokenCount, saver.Truncate(event.Response, 50))
	return nil
}

func NewSaver(opts ...saver.Option) saver.V1Saver {
	options := saver.NewOptions(opts...)

	s := &noopV1Saver{
		options: options,
	}

	return s
}
