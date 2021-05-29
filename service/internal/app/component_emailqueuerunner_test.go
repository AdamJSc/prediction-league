package app

import (
	"errors"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewEmailQueueRunner(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		if _, gotErr := NewEmailQueueRunner(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		mg := &mailgun.Client{}
		q := domain.NewInMemEmailQueue()
		l := &logger.Logger{}

		tt := []struct {
			emlQ domain.EmailQueue
			l    domain.Logger
		}{
			{nil, l},
			{q, nil},
		}

		for _, tc := range tt {
			cnt := &container{emailClient: mg, emailQueue: tc.emlQ, logger: tc.l}
			if _, gotErr := NewEmailQueueRunner(cnt); !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
			}
		}

		// accept nil email client
		cnt := &container{emailClient: nil, emailQueue: q, logger: l}
		if _, gotErr := NewEmailQueueRunner(cnt); gotErr != nil {
			t.Fatalf("want nil err, got %s (%T)", gotErr, gotErr)
		}

		// want no error
		cnt = &container{emailClient: mg, emailQueue: q, logger: l}
		emlQ, err := NewEmailQueueRunner(cnt)
		if err != nil {
			t.Fatalf("want nil err, got %s (%T)", err, err)
		}
		if emlQ == nil {
			t.Fatal("want non-empty email queue, got nil")
		}
	})
}
