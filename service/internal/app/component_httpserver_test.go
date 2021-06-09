package app

import (
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewHTTPServer(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		if _, gotErr := NewHTTPServer(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		c := &Config{}
		l := &mockLogger{}

		tt := []struct {
			c       *Config
			l       domain.Logger
			wantErr error
		}{
			{nil, l, domain.ErrIsNil},
			{c, nil, domain.ErrIsNil},
			{c, l, nil},
		}

		for idx, tc := range tt {
			cnt := &container{config: tc.c, logger: tc.l}
			srv, gotErr := NewHTTPServer(cnt)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && srv == nil {
				t.Fatalf("tc #%d: want non-empty server, got nil", idx)
			}
		}
	})
}
