package domain

import (
	"errors"
	"testing"
)

func TestNewNoopFootballDataSource(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		// TODO - tests: replace with tt and wantErr
		l := &mockLogger{}

		if _, gotErr := NewNoopFootballDataSource(nil); !errors.Is(gotErr, ErrIsNil) {
			t.Fatalf("waNewNoopFootballDataSourcent ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		fds, err := NewNoopFootballDataSource(l)
		if err != nil {
			t.Fatal(err)
		}
		if fds == nil {
			t.Fatal("want non-empty logger football data source, got nil")
		}
	})
}
