package app

import (
	"context"
	"log"
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

	for message := range e.EmailQueue() {
		if !e.Config().InProduction {
			// if we're not in production mode, don't send via the client
			// just print to the logs instead
			log.Printf("email on queue: %+v", message)
			continue
		}

		// TODO - add retry mechanism
		if err := e.EmailClient().SendEmail(message); err != nil {
			log.Printf("failed to send email: %+v", err)
		}
	}

	return nil
}

// Halt stops the queue runner
func (e EmailQueueRunner) Halt(context.Context) error {
	log.Println("stopping email queue runner...")
	close(e.EmailQueue())
	return nil
}
