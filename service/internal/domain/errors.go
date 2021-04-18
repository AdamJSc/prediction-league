package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
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

// UnauthorizedError translates to a 401 Unauthorized response status code
type UnauthorizedError struct{ error }

// NotFoundError translates to a 404 Not Found response status code
type NotFoundError struct{ error }

// ConflictError translates to a 409 Conflict response status code
type ConflictError struct{ error }

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
