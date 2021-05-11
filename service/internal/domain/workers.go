package domain

import (
	"context"
	"fmt"
	"time"
)

// Worker defines the behaviour required for a worker
type Worker interface {
	DoWork(ctx context.Context) error
}

// HandleWorker returns a function that wraps and handles the provided Worker
func HandleWorker(ctx context.Context, name string, timeout int, w Worker, l Logger) (func(), error) {
	if w == nil {
		return nil, fmt.Errorf("worker: %w", ErrIsNil)
	}
	if l == nil {
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	return func() {
		ctxWithTO, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		if err := w.DoWork(ctxWithTO); err != nil {
			l.Errorf("%s: %w", name, err)
		}
	}, nil
}
