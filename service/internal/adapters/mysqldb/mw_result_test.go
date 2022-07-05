package mysqldb_test

import (
	"errors"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewMatchWeekResultRepo(t *testing.T) {
	t.Run("passing non-nil db must succeed", func(t *testing.T) {
		if _, err := mysqldb.NewMatchWeekResultRepo(db, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("passing nil db must produce the expected error", func(t *testing.T) {
		if _, err := mysqldb.NewMatchWeekResultRepo(nil, nil); !errors.Is(err, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekResultRepo_GetBySubmissionID(t *testing.T) {
	t.Skip()
	// TODO: feat - write repo method tests
}

func TestMatchWeekResultRepo_Insert(t *testing.T) {
	t.Skip()
	// TODO: feat - write repo method tests
}

func TestMatchWeekResultRepo_Update(t *testing.T) {
	t.Skip()
	// TODO: feat - write repo method tests
}
