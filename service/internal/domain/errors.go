package domain

import (
	"fmt"
	"prediction-league/service/internal/repositories"
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
	case repositories.DuplicateDBRecordError:
		return ConflictError{err}
	case repositories.MissingDBRecordError:
		return NotFoundError{err}
	}

	return InternalError{err}
}
