package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"prediction-league/service/internal/domain"
	"strings"
	"time"
)

var errInvalidKeyName = func(key string) error { return fmt.Errorf("invalid key name: %q", key) }

// response defines a JSON response body over HTTP.
type response struct {
	Code    int    `json:"code"`           // Any valid HTTP response code
	Message string `json:"message"`        // Any relevant message (optional)
	Data    *data  `json:"data,omitempty"` // Data to pass along to the response (optional)
}

// writeTo writes a JSON response to a HTTP writer.
func (r response) writeTo(w http.ResponseWriter) error {
	// don't attempt to write a body for 204s.
	if r.Code == http.StatusNoContent {
		w.WriteHeader(r.Code)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	return json.NewEncoder(w).Encode(r)
}

// okResponse returns a prepared 200 OK response.
func okResponse(data *data) *response {
	return &response{
		Code:    http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
	}
}

// createdResponse returns a prepared 201 Created response.
func createdResponse(data *data) *response {
	return &response{
		Code:    http.StatusCreated,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
	}
}

// unauthorizedError returns a prepared 401 Unauthorized error.
func unauthorizedError() *response {
	return &response{Code: http.StatusUnauthorized, Message: "unauthorized"}
}

// notFoundError returns a prepared 404 Not Found response, including the message passed by the user in the message field of the response object.
func notFoundError(msg interface{}) *response {
	return &response{Code: http.StatusNotFound, Message: fmt.Sprintf("resource not found: %v", msg)}
}

// internalError returns a prepared 500 Internal Server Error, including the error message in the message field of the response object.
func internalError(msg interface{}) *response {
	return &response{Code: http.StatusInternalServerError, Message: fmt.Sprintf("internal server error: %v", msg)}
}

type createEntryResponse struct {
	ID           string `json:"id"`
	Nickname     string `json:"nickname"`
	ShortCode    string `json:"short_code"`
	NeedsPayment bool   `json:"needs_payment"`
}

type retrieveSeasonResponse struct {
	Name  string        `json:"name"`
	Teams []domain.Team `json:"teams"`
}

type retrieveLatestEntryPredictionResponse struct {
	Teams       []domain.Team `json:"teams"`
	LastUpdated time.Time     `json:"last_updated"`
}

type retrieveLatestScoredEntryPredictionResponse struct {
	LastUpdated time.Time                              `json:"last_updated"`
	RoundScore  int                                    `json:"round_score"`
	Rankings    []scoredEntryPredictionResponseRanking `json:"rankings"`
}

type scoredEntryPredictionResponseRanking struct {
	domain.RankingWithScore
	MetaPosition int `json:"meta_position"`
}

// responseFromError returns a rest package-level error from a domain-level error
func responseFromError(err error) *response {
	switch err.(type) {
	case domain.NotFoundError:
		return notFoundError(err)
	case domain.UnauthorizedError:
		return unauthorizedError()
	case domain.ConflictError:
		if cErr, ok := err.(domain.ConflictError); ok {
			return &response{
				Code:    http.StatusConflict,
				Message: "conflict error",
				Data: &data{
					Type:    "error",
					Content: strings.Title(cErr.Error()),
				},
			}
		}
	case domain.ValidationError:
		if vErr, ok := err.(domain.ValidationError); ok {
			return &response{
				Code:    http.StatusUnprocessableEntity,
				Message: "validation error",
				Data: &data{
					Type:    "error",
					Content: vErr,
				},
			}
		}
	case domain.BadRequestError:
		return &response{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	return internalError(err)
}

// data represents the collection data the the response will return to the consumer.
// Type ends up being the name of the key containing the collection of Content
type data struct {
	Type    string
	Content interface{}
}

// UnmarshalJSON implements the Unmarshaler interface
// this implementation will fill the type in the case we're been provided a valid single collection
// and set the content to the contents of said collection.
// for every other options, it behaves like normal.
// Despite the fact that we are not supposed to marshal without a type set,
// this is purposefully left open to unmarshal without a collection name set, in case you may want to set it later,
// and for interop with other systems which may not send the collection properly.
func (d *data) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &d.Content); err != nil {
		log.Printf("cannot unmarshal data: %v", err)
	}

	data, ok := d.Content.(map[string]interface{})
	if !ok {
		return nil
	}
	// count how many collections were provided
	var count int
	for _, value := range data {
		switch value.(type) {
		case map[string]interface{}, []interface{}:
			count++
		}
	}
	if count > 1 {
		// we can stop there since this is not a single collection
		return nil
	}
	for key, value := range data {
		switch value.(type) {
		case map[string]interface{}, []interface{}:
			d.Type = key
			d.Content = data[key]
		}
	}

	return nil
}

// MarshalJSON implements the Marshaler interface and is there to ensure the output
// is correct when we return data to the consumer
func (d *data) MarshalJSON() ([]byte, error) {
	if d.Type == "" || strings.Contains(d.Type, " ") {
		return nil, errInvalidKeyName(d.Type)
	}
	return json.Marshal(map[string]interface{}{
		d.Type: d.Content,
	})
}
