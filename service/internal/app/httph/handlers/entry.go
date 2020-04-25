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

type createEntryRequest struct {
	EntrantName     string `json:"entrant_name"`
	EntrantNickname string `json:"entrant_nickname"`
	EntrantEmail    string `json:"entrant_email"`
	PIN             string `json:"pin"`
}

func (r createEntryRequest) ToEntryModel() domain.Entry {
	return domain.Entry{
		EntrantName:     r.EntrantName,
		EntrantNickname: r.EntrantNickname,
		EntrantEmail:    r.EntrantEmail,
	}
}

type createEntryResponse struct {
	ID           string `json:"id"`
	EntrantName  string `json:"entrant_name"`
	EntrantEmail string `json:"entrant_email"`
	LookupURL    string `json:"lookup_url"`
}

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

		ctx.Guard.SetAttempt(input.PIN)

		// create entry
		createdEntry, err := agent.CreateEntry(ctx, entry, &season)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// generate a lookup URL for the created entry
		lookupURL := fmt.Sprintf("http://%s/%s", r.Host, createdEntry.LookupRef)

		rest.CreatedResponse(&rest.Data{
			Type: "entry",
			Content: createEntryResponse{
				ID:           createdEntry.ID.String(),
				EntrantName:  createdEntry.EntrantName,
				EntrantEmail: createdEntry.EntrantEmail,
				LookupURL:    lookupURL,
			},
		}, nil).WriteTo(w)
	}
}

type updateEntryPaymentDetailsRequest struct {
	PaymentMethod string `json:"payment_method"`
	PaymentRef    string `json:"payment_ref"`
	PassCode      string `json:"pass_code"`
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

		ctx.Guard.SetAttempt(input.PassCode)

		// update payment details for entry
		if _, err := agent.UpdateEntryPaymentDetails(ctx, entryID, input.PaymentMethod, input.PaymentRef); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}
