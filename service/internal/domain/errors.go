package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrIsNil defines an error that represents a missing argument passed to a constructor function
	ErrIsNil = errors.New("is nil")
	// ErrIsEmpty defines an error that represents an empty argument passed to a constructor function
	ErrIsEmpty = errors.New("is empty")
	// ErrIsInvalid defines an error that represents an argument passed with an invalid value
	ErrIsInvalid = errors.New("is invalid")
	// ErrCurrentTimeFrameIsMissing defines an error representing a missing timeframe
	ErrCurrentTimeFrameIsMissing = errors.New("current timeframe is missing")
	// ErrNoMatchingPredictionWindow defines an error representing a no matching prediction windows
	ErrNoMatchingPredictionWindow = errors.New("no matching prediction window")
)

// BadRequestError translates to a 400 Bad Request response status code
type BadRequestError struct{ Err error }

func (e BadRequestError) Error() string {
	return e.Err.Error()
}

func (e BadRequestError) Unwrap() error {
	return e.Err
}

// UnauthorizedError translates to a 401 Unauthorized response status code
type UnauthorizedError struct{ error }

func (e UnauthorizedError) Unwrap() error {
	return e.error
}

// NotFoundError translates to a 404 Not Found response status code
type NotFoundError struct{ error }

func (e NotFoundError) Unwrap() error {
	return e.error
}

// ConflictError translates to a 409 Conflict response status code
type ConflictError struct{ error }

func (e ConflictError) Unwrap() error {
	return e.error
}

// ValidationError translates to a 422 Unprocessable Entity response status code
type ValidationError struct {
	Reasons []string `json:"reasons"`
}

func (e ValidationError) Error() string {
	reasons := strings.Join(e.Reasons, " | ")
	return fmt.Sprintf("reasons: %s", strings.ToLower(reasons))
}

// InternalError translates to a 500 Internal Server Error response status code
type InternalError struct{ error }

func (e InternalError) Unwrap() error {
	return e.error
}

// domainErrorFromRepositoryError returns the appropriate domain-level error from a repository-specific error
func domainErrorFromRepositoryError(err error) error {
	switch err.(type) {
	case DuplicateDBRecordError:
		return ConflictError{err}
	case MissingDBRecordError:
		return NotFoundError{err}
	}

	return InternalError{err}
}

// MissingDBRecordError represents an error from an SQL agent that pertains to a missing record
type MissingDBRecordError struct {
	Err error
}

func (m MissingDBRecordError) Error() string {
	return m.Err.Error()
}

// DuplicateDBRecordError represents an error from an SQL agent that pertains to a unique constraint violation
type DuplicateDBRecordError struct {
	Err error
}

func (d DuplicateDBRecordError) Error() string {
	return d.Err.Error()
}

// MultiError encapsulates multiple errors
type MultiError struct {
	Errs []error
}

func (m MultiError) Error() string {
	return "multiple errors occurred"
}
