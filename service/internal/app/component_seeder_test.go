package app

import (
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewSeeder(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		if _, gotErr := NewSeeder(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		ea := &domain.EntryAgent{}
		l := &mockLogger{}

		tt := []struct {
			ea      *domain.EntryAgent
			l       domain.Logger
			wantErr error
		}{
			{nil, l, domain.ErrIsNil},
			{ea, nil, domain.ErrIsNil},
			{ea, l, nil},
		}
		for idx, tc := range tt {
			s, gotErr := NewSeeder(&container{entryAgent: tc.ea, logger: tc.l})
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && s == nil {
				t.Fatalf("tc #%d: want non-empty seeder, got nil", idx)
			}
		}
	})
}

type mockLogger struct{ domain.Logger }
