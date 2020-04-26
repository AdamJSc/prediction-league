package domain

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
)

// BadRequestError translates to a 400 Bad Request response status code
type BadRequestError struct{ Err error }

func (e BadRequestError) Error() string {
	return e.Err.Error()
}

// UnauthorizedError translates to a 401 Unauthorized response status code
type UnauthorizedError struct{ error }

// NotFoundError translates to a 404 Not Found response status code
type NotFoundError struct{ error }

// ConflictError translates to a 409 Conflict response status code
type ConflictError struct{ error }

// ValidationError translates to a 422 Unprocessable Entity response status code
type ValidationError struct {
	Reasons []string `json:"reasons"`
	Fields  []string `json:"fields"`
}

func (e ValidationError) Error() string {
	reasons := strings.Join(e.Reasons, " | ")
	fields := strings.Join(e.Fields, " | ")
	return fmt.Sprintf("reasons: %s, with fields: %v", strings.ToLower(reasons), strings.ToLower(fields))
}

// InternalError translates to a 500 Internal Server Error response status code
type InternalError struct{ error }

// vPackageErrorToValidationError extracts field names and messages from an error returned by the `v` validation package
// and transforms them into a ValidationError
func vPackageErrorToValidationError(err error, structure interface{}) ValidationError {
	var reasons, fields []string

	for _, validationPart := range strings.Split(err.Error(), "[validation]") {
		// reason is the full validation text
		reason := strings.Trim(validationPart, " ")
		if reason == "" {
			continue
		}

		reasons = append(reasons, reason)

		// structFieldName is the first part of the validation text, up to the colon delimiter
		structFieldName := strings.Trim(strings.Split(reason, ":")[0], " ")
		if structFieldName == "" {
			continue
		}

		// fallback to the original struct field name
		field := structFieldName
		if structField, ok := reflect.TypeOf(structure).FieldByName(structFieldName); ok {
			// try to get a db annotation for the struct field
			if tag := structField.Tag.Get("db"); tag != "" {
				field = tag
			}
			// try to get a json annotation
			if tag := structField.Tag.Get("json"); tag != "" {
				field = tag
			}
		}
		fields = append(fields, field)
	}

	return ValidationError{
		Reasons: reasons,
		Fields:  fields,
	}
}

// dbMissingRecordError represents an error from an SQL agent that pertains to a missing record
type dbMissingRecordError struct {
	error
}

// dbDuplicateRecordError represents an error from an SQL agent that pertains to a unique constraint violation
type dbDuplicateRecordError struct {
	error
}

// wrapDBError wraps an error from an SQL agent according to its nature as per the representations above
func wrapDBError(err error) error {
	if e, ok := err.(*mysql.MySQLError); ok {
		switch e.Number {
		case 1060:
		case 1061:
		case 1062:
			return dbDuplicateRecordError{err}
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return dbMissingRecordError{err}
	}

	return err
}

// domainErrorFromDBError returns the appropriate domain-level error from a database-specific error
func domainErrorFromDBError(err error) error {
	switch err.(type) {
	case dbDuplicateRecordError:
		return ConflictError{err}
	case dbMissingRecordError:
		return NotFoundError{err}
	}

	return InternalError{err}
}
