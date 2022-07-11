package app

import (
	"fmt"
	"net/http"
	"prediction-league/service/internal/domain"
)

func retrieveSeasonHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
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

		realm := domain.RealmFromContext(ctx)

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = realm.Config.SeasonID
		}

		// retrieve the season we need
		season, err := c.seasons.GetByID(seasonID)
		if err != nil {
			notFoundError(fmt.Errorf("invalid season: %s", seasonID)).writeTo(w)
			return
		}

		// get teams that correlate to season's team IDs
		teams, err := domain.FilterTeamsByIDs(season.TeamIDs, c.teams)
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

func retrieveLeaderBoardHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
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

		realm := domain.RealmFromContext(ctx)

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = realm.Config.SeasonID
		}

		// retrieve leaderboard
		lb, err := c.lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, roundNumber)
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
