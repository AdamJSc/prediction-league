package handlers

import (
	"net/http"
	"prediction-league/service/internal/app/httph"

	"github.com/LUSHDigital/core/rest"
	"github.com/gorilla/mux"
)

func showDatabasesEndpoint(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVal, ok := mux.Vars(r)["anything"]
		if !ok {
			rest.JSONError("vars went wrong").WriteTo(w)
			return
		}

		rows, err := c.MySQL().Query("SHOW DATABASES")
		if err != nil {
			rest.JSONError(err).WriteTo(w)
			return
		}

		var results []string
		var result string

		for rows.Next() {
			if err := rows.Scan(&result); err != nil {
				rest.JSONError(err).WriteTo(w)
				return
			}
			results = append(results, result)
		}

		rest.OKResponse(&rest.Data{
			Type:    pathVal,
			Content: results,
		}, nil).WriteTo(w)
	}
}

// RegisterRoutes attaches all routes to the router
func RegisterRoutes(c *httph.HTTPAppContainer) {
	// unauthenticated endpoints
	c.Router().HandleFunc("/{anything}", showDatabasesEndpoint(c)).Methods(http.MethodGet)

	// api endpoints
	api := c.Router().PathPrefix("/api").Subrouter()
	api.HandleFunc("/season", createSeasonHandler(c)).Methods(http.MethodPost)
}
