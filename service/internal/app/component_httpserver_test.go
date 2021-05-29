package app

import (
	"errors"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewHTTPServer(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		if _, gotErr := NewHTTPServer(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		c := &config{}
		l := &logger.Logger{}

		tt := []struct {
			c *config
			l domain.Logger
		}{
			{nil, l},
			{c, nil},
		}

		for _, tc := range tt {
			cnt := &container{config: tc.c, logger: tc.l}
			if _, gotErr := NewHTTPServer(cnt); !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
			}
		}

		// want no error
		cnt := &container{config: c, logger: l}
		srv, err := NewHTTPServer(cnt)
		if err != nil {
			t.Fatalf("want nil err, got %s (%T)", err, err)
		}
		if srv == nil {
			t.Fatal("want non-empty http server, got nil")
		}
	})
}
