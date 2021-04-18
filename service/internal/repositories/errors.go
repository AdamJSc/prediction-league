package repositories

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"prediction-league/service/internal/domain"
)

// wrapDBError wraps an error from an SQL agent according to its nature as per the representations above
func wrapDBError(err error) error {
	if e, ok := err.(*mysql.MySQLError); ok {
		switch e.Number {
		case 1060:
		case 1061:
		case 1062:
			return domain.DuplicateDBRecordError{Err: err}
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.MissingDBRecordError{Err: err}
	}

	return err
}
