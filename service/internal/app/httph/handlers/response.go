package handlers

import (
	"github.com/LUSHDigital/core/rest"
	"log"
	"net/http"
	"prediction-league/service/internal/domain"
)

func responseFromError(err error) *rest.Response {
	log.Println(err)

	switch err.(type) {
	case domain.NotFoundError:
		return rest.NotFoundError(err)
	case domain.ConflictError:
		return rest.ConflictError(err)
	case domain.ValidationError:
		if vErr, ok := err.(domain.ValidationError); ok {
			return &rest.Response{
				Code:    http.StatusUnprocessableEntity,
				Message: vErr.Reason,
				Data: &rest.Data{
					Type:    "fields",
					Content: vErr.Fields,
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
