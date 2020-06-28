package handlers

import (
	"net/http"
	"prediction-league/service/internal/app/httph"
)

// RegisterRoutes attaches all routes to the router
func RegisterRoutes(c *httph.HTTPAppContainer) {
	// api endpoints
	api := c.Router().PathPrefix("/api").Subrouter()
	api.HandleFunc("/selection/login", selectionLoginHandler(c)).Methods(http.MethodPost)

	api.HandleFunc("/season/{season_id}", retrieveSeasonHandler(c)).Methods(http.MethodGet)
	api.HandleFunc("/season/{season_id}/entry", createEntryHandler(c)).Methods(http.MethodPost)

	api.HandleFunc("/entry/{entry_id}/selection", createEntrySelectionHandler(c)).Methods(http.MethodPost)
	api.HandleFunc("/entry/{entry_id}/selection", retrieveLatestEntrySelectionHandler(c)).Methods(http.MethodGet)
	api.HandleFunc("/entry/{entry_id}/payment", updateEntryPaymentDetailsHandler(c)).Methods(http.MethodPatch)
	api.HandleFunc("/entry/{entry_short_code}/approve", approveEntryByShortCodeHandler(c)).Methods(http.MethodPatch)

	// serve static assets
	assets := http.Dir("./resources/dist")
	c.Router().PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(assets)))

	// frontend endpoints
	c.Router().HandleFunc("/", frontendIndexHandler(c)).Methods(http.MethodGet)
	c.Router().HandleFunc("/results", frontendResultsHandler(c)).Methods(http.MethodGet)
	c.Router().HandleFunc("/faq", frontendFAQHandler(c)).Methods(http.MethodGet)
	c.Router().HandleFunc("/enter", frontendEnterHandler(c)).Methods(http.MethodGet)
	c.Router().HandleFunc("/selection", frontendSelectionHandler(c)).Methods(http.MethodGet)

	// temporary routes
	c.Router().HandleFunc("/login", frontendLoginHandler(c)).Methods(http.MethodGet)
}
