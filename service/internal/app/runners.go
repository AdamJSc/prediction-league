package app

import (
	"context"
	"fmt"
	"prediction-league/service/internal/domain"
	"time"
)

// EmailQueueRunner handles the sending of emails added to the email queue
type EmailQueueRunner struct {
	emlCl domain.EmailClient
	emlQ  chan domain.Email
	l     domain.Logger
}

// Run starts the queue runner
func (e *EmailQueueRunner) Run(_ context.Context) error {
	e.l.Info("starting email queue runner...")

	if e.emlCl == nil {
		e.l.Info("missing email client: transactional email content will be printed to log...")
	}

	for msg := range e.emlQ {
		if e.emlCl == nil {
			// email client config is missing so just print to the logs instead
			e.l.Infof("email on queue: %+v", msg)
			continue
		}

		go func(em domain.Email) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := e.emlCl.SendEmail(ctx, em); err != nil {
				e.l.Errorf("failed to send email: %s", err.Error())
			}
		}(msg)
	}

	return nil
}

// Halt stops the queue runner
func (e *EmailQueueRunner) Halt(context.Context) error {
	e.l.Info("stopping email queue runner...")
	close(e.emlQ)
	return nil
}

func NewEmailQueueRunner(cnt *container) (*EmailQueueRunner, error) {
	if cnt == nil {
		return nil, fmt.Errorf("container: %w", domain.ErrIsNil)
	}
	if cnt.emailQueue == nil {
		return nil, fmt.Errorf("email queue: %w", domain.ErrIsNil)
	}
	if cnt.logger == nil {
		return nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}
	return &EmailQueueRunner{
		emlCl: cnt.emailClient,
		emlQ:  cnt.emailQueue,
		l:     cnt.logger,
	}, nil
}
