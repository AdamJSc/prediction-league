package domain_test

import (
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestHandleWorker(t *testing.T) {
	t.Run("passing nil must return the expected error", func(t *testing.T) {
		tt := []struct {
			w       domain.Worker
			l       domain.Logger
			wantErr bool
		}{
			{nil, &mockLogger{}, true},
			{&mockWorker{}, nil, true},
			{&mockWorker{}, &mockLogger{}, false},
		}
		for idx, tc := range tt {
			fn, gotErr := domain.HandleWorker("aaa", 123, tc.w, tc.l)
			if tc.wantErr && !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("tc #%d: want ErrIsNil, got %s (%T)", idx, gotErr, gotErr)
			}
			if !tc.wantErr && fn == nil {
				t.Fatalf("tc #%d: want non-empty worker function, got nil", idx)
			}
		}
	})
}

type mockWorker struct{ domain.Worker }
type mockLogger struct{ domain.Logger }
