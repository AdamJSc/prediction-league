package handlers

import (
	"github.com/markbates/pkger"
	"net/http"
	"prediction-league/service/internal/app/httph"
)

// RegisterRoutes attaches all routes to the router
func RegisterRoutes(c *httph.HTTPAppContainer) {
	// api endpoints
	api := c.Router().PathPrefix("/api").Subrouter()
	api.HandleFunc("/season/{season_id}/entry", createEntryHandler(c)).Methods(http.MethodPost)
	api.HandleFunc("/entry/{entry_id}/payment", updateEntryPaymentDetailsHandler(c)).Methods(http.MethodPatch)
	api.HandleFunc("/entry/{entry_short_code}/approve", approveEntryByShortCodeHandler(c)).Methods(http.MethodPatch)

	// serve static assets
	assets := pkger.Dir("/resources/dist")
	c.Router().PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(assets)))

	// frontend endpoints
	c.Router().HandleFunc("/{greeting}", frontendGreetingHandler(c)).Methods(http.MethodGet)
}
