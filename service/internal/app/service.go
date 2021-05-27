package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"prediction-league/service/internal/domain"
	"sync"
	"syscall"
	"time"
)

// service represents the minimal information required to define a working service.
type service struct {
	name  string
	grace time.Duration
	l     domain.Logger
}

// MustRun will start the given service workers and block indefinitely, until interrupted.
// The process exits with an appropriate status code
func (s *service) MustRun(ctx context.Context, cmps ...Component) {
	os.Exit(s.Run(ctx, cmps...))
}

// Run will start the given service workers and block indefinitely, until interrupted.
func (s *service) Run(ctx context.Context, cmps ...Component) int {
	const fail int = 1
	nCmps := len(cmps)
	if nCmps < 1 {
		s.l.Error("need at least 1 service component")
		return fail
	}
	var (
		cancelled    <-chan int
		completed    <-chan int
		done, cancel func()
	)
	ctx, cancelled, cancel = contextWithSignals(ctx, s.l)
	completed, cancelled, done = waitWithTimeout(nCmps, cancelled, s.grace)

	var run = func(ctx context.Context, cmp Component, done, cancel func()) {
		if err := cmp.Run(ctx); err != nil {
			s.l.Errorf("failed to run component: %s", err.Error())
			go cancel()
		}
		done()
	}
	var halt = func(ctx context.Context, cmp Component) {
		if err := cmp.Halt(ctx); err != nil {
			s.l.Errorf("failed to halt component: %s", err.Error())
		}
	}

	s.l.Infof("starting service: %s", s.name)

	for _, cmp := range cmps {
		go run(ctx, cmp, done, cancel)
	}
	for {
		select {
		case <-cancelled:
			for _, cmp := range cmps {
				go halt(ctx, cmp)
			}
		case code := <-completed:
			message := "shutdown gracefully..."
			if code > 0 {
				message = "failed to shutdown gracefully: killing!"
			}
			s.l.Info(message)
			return code
		}
	}
}

func NewService(name string, toSecs int, l domain.Logger) (*service, error) {
	if name == "" {
		return nil, fmt.Errorf("name: %w", domain.ErrIsEmpty)
	}
	if toSecs == 0 {
		return nil, fmt.Errorf("timeout: %w", domain.ErrIsEmpty)
	}
	if l == nil {
		return nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}

	return &service{
		name:  name,
		grace: time.Duration(toSecs) * time.Second,
		l:     l,
	}, nil
}

// Component represents the behaviour for a runnable service component
type Component interface {
	// Run should run start processing the component and be a blocking operation
	Run(context.Context) error
	// Halt should tell the worker to stop doing work
	Halt(context.Context) error
}

// contextWithSignals creates a new instance of signal context.
func contextWithSignals(ctx context.Context, l domain.Logger) (context.Context, <-chan int, context.CancelFunc) {
	var cancelCtx context.CancelFunc
	ctx, cancelCtx = context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	cancelled := make(chan int, 1)

	signal.Notify(sigs,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	var cancel = func() {
		cancelCtx()
		cancelled <- 1
	}

	go func() {
		sig := <-sigs
		l.Infof("received signal: %s", sig)
		cancel()
	}()

	return ctx, cancelled, cancel
}

// waitWithTimeout will wait for a number of pieces of work has finished and send a message on the completed channel.
func waitWithTimeout(delta int, cancelled <-chan int, timeout time.Duration) (<-chan int, <-chan int, func()) {
	completedC := make(chan int, 1)
	cancelledC := make(chan int, 1)
	wg := &sync.WaitGroup{}
	wg.Add(delta)
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		completedC <- 0
	}(wg)
	go func() {
		select {
		case code := <-cancelled:
			cancelledC <- code
			time.Sleep(timeout)
			completedC <- code
		}
	}()
	return completedC, cancelledC, wg.Done
}
