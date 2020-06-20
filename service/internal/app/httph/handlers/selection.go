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

	// retrieve cookie value
	authCookieValue, err := getAuthCookieValue(r)
	if err != nil {
		data.Err = err
		return data
	}

	ctx, cancel, err := contextFromRequest(r, c)
	if err != nil {
		data.Err = err
		return data
	}
	defer cancel()

	// check that entry id is valid
	entry, err := agent.RetrieveEntryByID(ctx, authCookieValue)
	if err != nil {
		data.Err = err
		return data
	}

	// retrieve season
	season, err := datastore.Seasons.GetByID(entry.SeasonID)
	if err != nil {
		data.Err = err
		return data
	}

	// default teams IDs should be the current season
	var teamIDs = season.TeamIDs

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

	// ensure that season is accepting selections currently
	state := season.GetState(domain.TimestampFromContext(ctx))
	data.IsAcceptingSelections = state.IsAcceptingSelections

	// populate remaining data
	data.SelectionsNextAccepted = state.SelectionsNextAccepted
	data.Teams.Raw = string(teamsPayload)
	data.Teams.LastUpdated = entrySelection.CreatedAt
	data.Entry.ID = entry.ID.String()
	data.Entry.ShortCode = entry.ShortCode

	return data
}
