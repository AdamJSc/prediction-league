package domain

import (
	"errors"
	"testing"
)

func TestHandleWorker(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		w := &mockWorker{}
		l := &mockLogger{}

		tt := []struct {
			w       Worker
			l       Logger
			wantErr error
		}{
			{nil, l, ErrIsNil},
			{w, nil, ErrIsNil},
			{w, l, nil},
		}
		for idx, tc := range tt {
			fn, gotErr := HandleWorker("aaa", 123, tc.w, tc.l)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && fn == nil {
				t.Fatalf("tc #%d: want non-empty worker function, got nil", idx)
			}
		}
	})
}

type mockWorker struct{ Worker }
type mockLogger struct{ Logger }
