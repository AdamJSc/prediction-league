package app

import (
	"net/http"
)

// RegisterRoutes attaches all routes to the router
func RegisterRoutes(c *container) {
	// api endpoints
	api := c.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/prediction/login", predictionLoginHandler(c)).Methods(http.MethodPost)

	api.HandleFunc("/season/{season_id}", retrieveSeasonHandler(c)).Methods(http.MethodGet)
	api.HandleFunc("/season/{season_id}/entry", createEntryHandler(c)).Methods(http.MethodPost)
	api.HandleFunc("/season/{season_id}/leaderboard/{round_number:[0-9]+}", retrieveLeaderBoardHandler(c)).Methods(http.MethodGet)

	api.HandleFunc("/entry/{entry_id}/prediction", createEntryPredictionHandler(c)).Methods(http.MethodPost)
	api.HandleFunc("/entry/{entry_id}/prediction", retrieveLatestEntryPredictionHandler(c)).Methods(http.MethodGet)
	api.HandleFunc("/entry/{entry_id}/scored/{round_number:[0-9]+}", retrieveLatestScoredEntryPrediction(c)).Methods(http.MethodGet)
	api.HandleFunc("/entry/{entry_id}/payment", updateEntryPaymentDetailsHandler(c)).Methods(http.MethodPatch)
	api.HandleFunc("/entry/{entry_short_code}/approve", approveEntryByShortCodeHandler(c)).Methods(http.MethodPatch)

	// serve static assets
	assets := http.Dir("./resources/dist")
	c.router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(assets)))

	// frontend endpoints
	c.router.HandleFunc("/", frontendIndexHandler(c)).Methods(http.MethodGet)
	c.router.HandleFunc("/leaderboard", frontendLeaderBoardHandler(c)).Methods(http.MethodGet)
	c.router.HandleFunc("/faq", frontendFAQHandler(c)).Methods(http.MethodGet)
	c.router.HandleFunc("/join", frontendJoinHandler(c)).Methods(http.MethodGet)
	c.router.HandleFunc("/prediction", frontendPredictionHandler(c)).Methods(http.MethodGet)

	c.router.HandleFunc("/reset", frontendShortCodeResetBeginHandler(c)).Methods(http.MethodPost)
	c.router.HandleFunc("/reset/{reset_token}", frontendShortCodeResetCompleteHandler(c)).Methods(http.MethodGet)
}
