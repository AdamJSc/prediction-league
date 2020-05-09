package repositories

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
)

// wrapDBError wraps an error from an SQL agent according to its nature as per the representations above
func wrapDBError(err error) error {
	if e, ok := err.(*mysql.MySQLError); ok {
		switch e.Number {
		case 1060:
		case 1061:
		case 1062:
			return DuplicateDBRecordError{err}
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return MissingDBRecordError{err}
	}

	return err
}

// MissingDBRecordError represents an error from an SQL agent that pertains to a missing record
type MissingDBRecordError struct {
	error
}

// DuplicateDBRecordError represents an error from an SQL agent that pertains to a unique constraint violation
type DuplicateDBRecordError struct {
	error
}
