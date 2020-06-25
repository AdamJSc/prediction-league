package handlers

import (
	"encoding/json"
	"github.com/LUSHDigital/core/rest"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/pages"
)

func selectionLoginHandler(c *httph.HTTPAppContainer) func(http.ResponseWriter, *http.Request) {
	entryAgent := domain.EntryAgent{EntryAgentInjector: c}
	tokenAgent := domain.TokenAgent{TokenAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input selectionLoginRequest

		// read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
		defer closeBody(r)

		// parse request body
		if err := json.Unmarshal(body, &input); err != nil {
			responseFromError(domain.BadRequestError{Err: err}).WriteTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}
		defer cancel()

		var retrieveEntryFromInput = func(input selectionLoginRequest) (*models.Entry, error) {
			// see if we can retrieve by email
			entry, err := entryAgent.RetrieveEntryByEntrantEmail(ctx, input.EmailNickname)
			if err != nil {
				switch err.(type) {
				case domain.NotFoundError:
					// see if we can retrieve by nickname
					entry, err := entryAgent.RetrieveEntryByEntrantNickname(ctx, input.EmailNickname)
					if err != nil {
						return nil, err
					}

					return &entry, nil
				default:
					return nil, err
				}
			}

			return &entry, nil
		}

		// retrieve entry based on input
		entry, err := retrieveEntryFromInput(input)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				// credentials are invalid so convert to an unauthorized error
				rest.UnauthorizedError().WriteTo(w)
				return
			}
			responseFromError(err).WriteTo(w)
			return
		}

		// does short code match our entry?
		if entry.ShortCode != input.ShortCode {
			rest.UnauthorizedError().WriteTo(w)
			return
		}

		// generate a new auth token for our entry, and set it as a cookie
		token, err := tokenAgent.GenerateToken(ctx, models.TokenTypeAuthToken, entry.ID.String())
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}
		setAuthCookieValue(token.ID, w, r)

		rest.OKResponse(nil, nil).WriteTo(w)
	}
}

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
	data.Selections.BeingAccepted = seasonState.IsAcceptingSelections
	if seasonState.NextSelectionsWindow != nil {
		switch data.Selections.BeingAccepted {
		case true:
			data.Selections.AcceptedUntil = &seasonState.NextSelectionsWindow.Until
		default:
			data.Selections.NextAcceptedFrom = &seasonState.NextSelectionsWindow.From
		}
	}

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
	data.Teams.LastUpdated = &entrySelection.CreatedAt
	data.Entry.ID = entry.ID.String()
	data.Entry.ShortCode = entry.ShortCode

	return data
}
