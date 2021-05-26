package app

import (
	"errors"
	"os"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewEmailQueueRunner(t *testing.T) {
	if _, gotErr := NewEmailQueueRunner(nil); !errors.Is(gotErr, domain.ErrIsNil) {
		t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
	}

	mg := &mailgun.Client{}
	q := make(chan domain.Email)
	l, _ := logger.NewLogger(os.Stdout, &domain.RealClock{})

	tt := []struct {
		emlQ chan domain.Email
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
}
