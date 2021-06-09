package app

import (
	"errors"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewEmailQueueRunner(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		if _, gotErr := NewEmailQueueRunner(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		mg := &mailgun.Client{}
		q := domain.NewInMemEmailQueue()
		l := &mockLogger{}

		tt := []struct {
			emlCl domain.EmailClient
			emlQ  domain.EmailQueue
			l     domain.Logger
			wantErr error
		}{
			{nil, q, l, domain.ErrIsNil},
			{mg, nil, l, domain.ErrIsNil},
			{mg, q, nil, domain.ErrIsNil},
			{mg, q, l, nil},
		}

		for idx, tc := range tt {
			cnt := &container{emailClient: tc.emlCl, emailQueue: tc.emlQ, logger: tc.l}
			eqr, gotErr := NewEmailQueueRunner(cnt)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && eqr == nil {
				t.Fatalf("tc #%d: want non-empty email queue runner, got nil", idx)
			}
		}
	})
}
