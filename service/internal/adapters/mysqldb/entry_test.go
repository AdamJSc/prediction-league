package mysqldb_test

import (
	"database/sql"
	"errors"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewEntryRepo(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		db := &sql.DB{}

		tt := []struct {
			db     *sql.DB
			wantErr error
		}{
			{nil, domain.ErrIsNil},
			{db, nil},
		}
		for idx, tc := range tt {
			repo, gotErr := mysqldb.NewEntryRepo(tc.db)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && repo == nil {
				t.Fatalf("tc #%d: want non-empty repo, got nil", idx)
			}
		}
	})
}
