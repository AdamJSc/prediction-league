package domain

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
)

type ValidationError struct {
	Reason string
	Fields []string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("reason: %s, with fields: %v", e.Reason, strings.Join(e.Fields, ","))
}

type ConflictError struct{ error }

type NotFoundError struct{ error }

type InternalError struct{ error }

func fieldsFromValidationPackageError(err error, structure interface{}) []string {
	var fields []string

	for _, validation := range strings.Split(err.Error(), "[validation]") {
		fieldName := strings.Trim(strings.Split(validation, ":")[0], " ")
		if fieldName == "" {
			continue
		}

		field := fieldName
		if structField, ok := reflect.TypeOf(structure).FieldByName(fieldName); ok {
			// try and get an annotation for the struct field
			if tag := structField.Tag.Get("json"); tag != "" {
				field = tag
			}
			if tag := structField.Tag.Get("db"); tag != "" {
				field = tag
			}
		}
		fields = append(fields, field)
	}

	return fields
}

type dbMissingRecordError struct {
	error
}

type dbDuplicateRecordError struct {
	error
}

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

func domainErrorFromDBError(err error) error {
	switch err.(type) {
	case dbDuplicateRecordError:
		return ConflictError{err}
	case dbMissingRecordError:
		return NotFoundError{err}
	}

	return InternalError{err}
}
