package handlers

import (
	"encoding/json"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/pages"
)

func getSelectionPageData(r *http.Request, c *httph.HTTPAppContainer) pages.SelectionPageData {
	var data pages.SelectionPageData

	agent := domain.EntryAgent{EntryAgentInjector: c}

	ctx, cancel, err := contextFromRequest(r, c)
	if err != nil {
		data.Err = err
		return data
	}
	defer cancel()

	// retrieve season and determine its current state
	seasonID := domain.RealmFromContext(ctx).SeasonID
	season, err := datastore.Seasons.GetByID(seasonID)
	if err != nil {
		data.Err = err
		return data
	}

	seasonState := season.GetState(domain.TimestampFromContext(ctx))

	data.Season.IsAcceptingSelections = seasonState.IsAcceptingSelections
	data.Season.SelectionsNextAccepted = seasonState.SelectionsNextAccepted

	// retrieve cookie value
	authCookieValue, err := getAuthCookieValue(r)
	if err != nil {
		// no auth, so return early
		// we already have our season-related data so this is fine to do
		return data
	}

	// default teams IDs should reflect those of the current season
	var teamIDs = season.TeamIDs

	// check that entry id is valid
	entry, err := agent.RetrieveEntryByID(ctx, authCookieValue)
	if err != nil {
		data.Err = err
		return data
	}

	// if entry has an associated entry selection
	// then override the team IDs with the most recent selection
	entrySelection, err := agent.RetrieveEntrySelectionByTimestamp(ctx, entry, domain.TimestampFromContext(ctx))
	if err == nil {
		teamIDs = entrySelection.Rankings.GetIDs()
	}

	// retrieve teams
	teams, err := domain.FilterTeamsByIDs(teamIDs, datastore.Teams)
	if err != nil {
		// something went wrong
		data.Err = err
		return data
	}

	teamsPayload, err := json.Marshal(teams)
	if err != nil {
		// something went wrong
		data.Err = err
		return data
	}

	// populate remaining teams and entry data
	data.Teams.Raw = string(teamsPayload)
	data.Teams.LastUpdated = entrySelection.CreatedAt
	data.Entry.ID = entry.ID.String()
	data.Entry.ShortCode = entry.ShortCode

	return data
}
