package domain

import (
	"errors"
	"testing"
)

func TestNewNoopFootballDataSource(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		l := &mockLogger{}

		tt := []struct {
			l Logger
			wantErr error
		}{
			{nil, ErrIsNil},
			{l, nil},
		}
		for idx, tc := range tt {
			fds, gotErr := NewNoopFootballDataSource(tc.l)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && fds == nil {
				t.Fatalf("tc #%d: want non-empty football data srouce, got nil", idx)
			}
		}
	})
}
