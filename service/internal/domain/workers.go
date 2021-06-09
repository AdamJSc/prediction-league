package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Worker defines the behaviour required for a worker
type Worker interface {
	DoWork(ctx context.Context) error
}

// HandleWorker returns a function that wraps and handles the provided Worker
func HandleWorker(name string, timeout int, w Worker, l Logger) (func(), error) {
	if w == nil {
		return nil, fmt.Errorf("worker: %w", ErrIsNil)
	}
	if l == nil {
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		var errs []error

		if err := w.DoWork(ctx); err != nil {
			mErr := MultiError{}
			switch {
			case errors.As(err, &mErr):
				for _, e := range mErr.Errs {
					errs = append(errs, e)
				}
			default:
				errs = []error{err}
			}
			for _, e := range errs {
				l.Errorf("%s: %s", name, e.Error())
			}
		}
	}, nil
}
