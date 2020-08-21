package app

import (
	"context"
	"log"
	"prediction-league/service/internal/messages"
	"time"
)

// EmailQueueRunnerInjector defines the dependencies required by our EmailQueueRunner
type EmailQueueRunnerInjector interface {
	ConfigInjector
	EmailClientInjector
	EmailQueueInjector
}

// EmailQueueRunner handles the sending of emails added to the email queue
type EmailQueueRunner struct {
	EmailQueueRunnerInjector
}

// Run starts the queue runner
func (e EmailQueueRunner) Run(_ context.Context) error {
	log.Println("starting email queue runner")

	if e.Config().MailgunAPIKey == "" {
		log.Println("missing config: mailgun... transactional email content will be printed to log...")
	}

	for message := range e.EmailQueue() {
		if e.Config().MailgunAPIKey == "" {
			// email client config is missing so just print to the logs instead
			log.Printf("email on queue: %+v", message)
			continue
		}

		go func(m messages.Email) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := e.EmailClient().SendEmail(ctx, m); err != nil {
				log.Printf("failed to send email: %+v", err)
			}
		}(message)

	}

	return nil
}

// Halt stops the queue runner
func (e EmailQueueRunner) Halt(context.Context) error {
	log.Println("stopping email queue runner...")
	close(e.EmailQueue())
	return nil
}
