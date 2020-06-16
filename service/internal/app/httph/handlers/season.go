package handlers

import (
	"fmt"
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
)

func retrieveSeasonHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse season ID from route
		var seasonID string
		if err := getRouteParam(r, "season_id", &seasonID); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}
		defer cancel()

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = domain.RealmFromContext(ctx).SeasonID
		}

		// retrieve the season we need
		season, err := datastore.Seasons.GetByID(seasonID)
		if err != nil {
			rest.NotFoundError(fmt.Errorf("invalid season: %s", seasonID)).WriteTo(w)
			return
		}

		// get teams that correlate to season's team IDs
		var teams []models.Team
		for _, teamID := range season.TeamIDs {
			team, err := datastore.Teams.GetByID(teamID)
			if err != nil {
				rest.NotFoundError(fmt.Errorf("invalid team: %s", teamID)).WriteTo(w)
				return
			}

			teams = append(teams, team)
		}

		rest.OKResponse(&rest.Data{
			Type: "season",
			Content: retrieveSeasonResponse{
				Name:  season.Name,
				Teams: teams,
			},
		}, nil).WriteTo(w)
	}
}
