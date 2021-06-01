package domain

import (
	"errors"
	"testing"
)

func TestHandleWorker(t *testing.T) {
	t.Run("passing nil must return the expected error", func(t *testing.T) {
		w := &mockWorker{}
		l := &mockLogger{}
		tt := []struct {
			w       Worker
			l       Logger
			wantErr bool
		}{
			{nil, l, true},
			{w, nil, true},
			{w, l, false},
		}
		for idx, tc := range tt {
			fn, gotErr := HandleWorker("aaa", 123, tc.w, tc.l)
			if tc.wantErr && !errors.Is(gotErr, ErrIsNil) {
				t.Fatalf("tc #%d: want ErrIsNil, got %s (%T)", idx, gotErr, gotErr)
			}
			if !tc.wantErr && fn == nil {
				t.Fatalf("tc #%d: want non-empty worker function, got nil", idx)
			}
		}
	})
}

type mockWorker struct{ Worker }
type mockLogger struct{ Logger }
