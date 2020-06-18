package handlers

import (
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

	// ensure that season is accepting selections currently
	state := season.GetState(domain.TimestampFromContext(ctx))
	data.IsAcceptingSelections = state.IsAcceptingSelections
	data.SelectionsNextAccepted = state.SelectionsNextAccepted

	return data
}
