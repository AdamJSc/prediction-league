package handlers

import (
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/domain"
)

// responseFromError returns a rest package-level error from a domain-level error
func responseFromError(err error) *rest.Response {
	switch err.(type) {
	case domain.NotFoundError:
		return rest.NotFoundError(err)
	case domain.UnauthorizedError:
		return rest.UnauthorizedError()
	case domain.ConflictError:
		return rest.ConflictError(err)
	case domain.ValidationError:
		if vErr, ok := err.(domain.ValidationError); ok {
			return &rest.Response{
				Code:    http.StatusUnprocessableEntity,
				Message: "validation error",
				Data: &rest.Data{
					Type:    "error",
					Content: vErr,
				},
			}
		}
	case domain.BadRequestError:
		return &rest.Response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	return rest.InternalError(err)
}
