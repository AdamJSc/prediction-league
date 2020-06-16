package handlers

import (
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"strings"
	"time"
)

type createEntryResponse struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	ShortCode string `json:"short_code"`
}

type retrieveSeasonResponse struct {
	Name  string        `json:"name"`
	Teams []models.Team `json:"teams"`
}

type retrieveLatestEntrySelectionResponse struct {
	Teams       []models.Team `json:"teams"`
	LastUpdated time.Time     `json:"last_updated"`
}

// responseFromError returns a rest package-level error from a domain-level error
func responseFromError(err error) *rest.Response {
	switch err.(type) {
	case domain.NotFoundError:
		return rest.NotFoundError(err)
	case domain.UnauthorizedError:
		return rest.UnauthorizedError()
	case domain.ConflictError:
		if cErr, ok := err.(domain.ConflictError); ok {
			return &rest.Response{
				Code:    http.StatusConflict,
				Message: "conflict error",
				Data: &rest.Data{
					Type:    "error",
					Content: strings.Title(cErr.Error()),
				},
			}
		}
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
