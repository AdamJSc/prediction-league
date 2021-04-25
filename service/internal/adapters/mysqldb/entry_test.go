package mysqldb_test

import (
	"errors"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewEntryRepo(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		_, gotErr := mysqldb.NewEntryRepo(nil)
		if !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}
	})
}
