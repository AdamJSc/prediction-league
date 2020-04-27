package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/LUSHDigital/core/rest"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/domain"
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

		ctx, err := domain.ContextFromRequest(r, c.Config())
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = ctx.Realm.SeasonID
		}

		// retrieve the season we need
		season, err := domain.Seasons().GetByID(seasonID)
		if err != nil {
			rest.NotFoundError(fmt.Errorf("invalid season: %s", seasonID)).WriteTo(w)
			return
		}

		ctx.Guard.SetAttempt(input.RealmPIN)

		// create entry
		createdEntry, err := agent.CreateEntry(ctx, entry, &season)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// generate a lookup URL for the created entry
		lookupURL := fmt.Sprintf("http://%s/%s", r.Host, createdEntry.ShortCode)

		rest.CreatedResponse(&rest.Data{
			Type: "entry",
			Content: createEntryResponse{
				ID:           createdEntry.ID.String(),
				EntrantName:  createdEntry.EntrantName,
				EntrantEmail: createdEntry.EntrantEmail,
				ShortCode:    createdEntry.ShortCode,
				ShortURL:     lookupURL,
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

		ctx, err := domain.ContextFromRequest(r, c.Config())
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		ctx.Guard.SetAttempt(input.ShortCode)

		// update payment details for entry
		if _, err := agent.UpdateEntryPaymentDetails(ctx, entryID, input.PaymentMethod, input.PaymentRef); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
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

		ctx, err := domain.ContextFromRequest(r, c.Config())
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// approve entry
		if _, err := agent.ApproveEntryByShortCode(ctx, entryShortCode); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}
