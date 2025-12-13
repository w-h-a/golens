package mock

import (
	"context"
	"sync"

	v1 "github.com/w-h-a/golens/api/event/v1"
	"github.com/w-h-a/golens/internal/client/saver"
)

type mockV1Saver struct {
	options  saver.Options
	captured *v1.Event
	count    int
	mtx      sync.RWMutex
}

func (s *mockV1Saver) Save(ctx context.Context, event *v1.Event, opts ...saver.SaveOption) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.captured = event
	s.count++

	return nil
}

func (s *mockV1Saver) Captured() *v1.Event {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.captured
}

func (s *mockV1Saver) Count() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.count
}

func NewSaver(opts ...saver.Option) *mockV1Saver {
	options := saver.NewOptions(opts...)

	s := &mockV1Saver{
		options: options,
		mtx:     sync.RWMutex{},
	}

	return s
}
