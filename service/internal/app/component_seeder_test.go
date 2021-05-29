package app

import (
	"errors"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewSeeder(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		ea := &domain.EntryAgent{}
		l := &logger.Logger{}

		if _, gotErr := NewSeeder(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		tt := []struct {
			ea *domain.EntryAgent
			l  domain.Logger
		}{
			{nil, l},
			{ea, nil},
		}
		for idx, tc := range tt {
			cnt := &container{entryAgent: tc.ea, logger: tc.l}
			if _, gotErr := NewSeeder(cnt); !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("tc #%d: want ErrIsNil, got %s (%T)", idx, gotErr, gotErr)
			}
		}

		s, err := NewSeeder(&container{entryAgent: ea, logger: l})
		if err != nil {
			t.Fatal(err)
		}
		if s == nil {
			t.Fatal("want non-empty seeder, got nil")
		}
	})
}
