package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/LUSHDigital/core/rest"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
)

func createEntryHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntryRequest

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

		// retrieve our model
		entry := input.ToEntryModel()

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

		domain.GuardFromContext(ctx).SetAttempt(input.RealmPIN)

		// create entry
		createdEntry, err := agent.CreateEntry(ctx, entry, &season)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		rest.CreatedResponse(&rest.Data{
			Type: "entry",
			Content: createEntryResponse{
				ID:        createdEntry.ID.String(),
				Nickname:  createdEntry.EntrantNickname,
				ShortCode: createdEntry.ShortCode,
			},
		}, nil).WriteTo(w)
	}
}

func updateEntryPaymentDetailsHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input updateEntryPaymentDetailsRequest

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

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
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

		domain.GuardFromContext(ctx).SetAttempt(input.EntryID)

		// update payment details for entry
		if _, err := agent.UpdateEntryPaymentDetails(ctx, entryID, input.PaymentMethod, input.PaymentRef); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}

func createEntrySelectionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntrySelectionRequest

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

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
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

		// get entry
		entry, err := agent.RetrieveEntryByID(ctx, entryID)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		entrySelection := input.ToEntrySelectionModel()

		domain.GuardFromContext(ctx).SetAttempt(input.EntryShortCode)

		// create entry selection for entry
		if _, err := agent.AddEntrySelectionToEntry(ctx, entrySelection, entry); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}

func retrieveLatestEntrySelectionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
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

		// get entry
		entry, err := agent.RetrieveEntryByID(ctx, entryID)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// get entry selection that pertains to context timestamp
		entrySelection, err := agent.RetrieveEntrySelectionByTimestamp(ctx, entry, domain.TimestampFromContext(ctx))
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// get teams that correlate to entry selection's ranking IDs
		var teams []models.Team
		for _, teamID := range entrySelection.Rankings.GetIDs() {
			team, err := datastore.Teams.GetByID(teamID)
			if err != nil {
				rest.NotFoundError(fmt.Errorf("invalid team: %s", teamID)).WriteTo(w)
				return
			}

			teams = append(teams, team)
		}

		rest.OKResponse(&rest.Data{
			Type: "entry_selection",
			Content: retrieveLatestEntrySelectionResponse{
				Teams:       teams,
				LastUpdated: entrySelection.CreatedAt,
			},
		}, nil).WriteTo(w)
	}
}

func approveEntryByShortCodeHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		// parse entry short code from route
		var entryShortCode string
		if err := getRouteParam(r, "entry_short_code", &entryShortCode); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}
		defer cancel()

		// approve entry
		if _, err := agent.ApproveEntryByShortCode(ctx, entryShortCode); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}
