package handlers

import (
	"net/http"
	"prediction-league/service/internal/app/httph"
)

// RegisterRoutes attaches all routes to the router
func RegisterRoutes(c *httph.HTTPAppContainer) {
	// placeholder endpoint
	c.Router().HandleFunc("/{anything}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}).Methods(http.MethodGet)

	// api endpoints
	api := c.Router().PathPrefix("/api").Subrouter()

	api.HandleFunc("/season/{season_id}/entry", createEntryHandler(c)).Methods(http.MethodPost)

	api.HandleFunc("/entry/{entry_id}/payment", updateEntryPaymentDetailsHandler(c)).Methods(http.MethodPatch)
}
