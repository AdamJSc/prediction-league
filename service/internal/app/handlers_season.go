package app

import (
	"fmt"
	"net/http"
	"prediction-league/service/internal/domain"
)

func retrieveSeasonHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse season ID from route
		var seasonID string
		if err := getRouteParam(r, "season_id", &seasonID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = domain.RealmFromContext(ctx).SeasonID
		}

		// retrieve the season we need
		season, err := c.Seasons().GetByID(seasonID)
		if err != nil {
			notFoundError(fmt.Errorf("invalid season: %s", seasonID)).writeTo(w)
			return
		}

		// get teams that correlate to season's team IDs
		teams, err := domain.FilterTeamsByIDs(season.TeamIDs, c.Teams())
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		okResponse(&data{
			Type: "season",
			Content: retrieveSeasonResponse{
				Name:  season.Name,
				Teams: teams,
			},
		}).writeTo(w)
	}
}

func retrieveLeaderBoardHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse season ID from route
		var seasonID string
		if err := getRouteParam(r, "season_id", &seasonID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// parse round number from route
		var roundNumber int
		if err := getRouteParam(r, "round_number", &roundNumber); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = domain.RealmFromContext(ctx).SeasonID
		}

		// retrieve leaderboard
		lbAgent, err := domain.NewLeaderBoardAgent(c.EntryRepo(), c.EntryPredictionRepo(), c.StandingsRepo(), c.ScoredEntryPredictionRepo(), c.Seasons())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		lb, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, roundNumber)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		okResponse(&data{
			Type:    "leaderboard",
			Content: lb,
		}).writeTo(w)
	}
}
