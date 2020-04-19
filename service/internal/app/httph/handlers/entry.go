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

// createEntryRequest defines the required structure for our create Entry request
type createEntryRequest struct {
	EntrantName     string `json:"entrant_name"`
	EntrantNickname string `json:"entrant_nickname"`
	EntrantEmail    string `json:"entrant_email"`
	PIN             string `json:"pin"`
}

// ToEntryModel transforms the request to an Entry object
func (r createEntryRequest) ToEntryModel() domain.Entry {
	return domain.Entry{
		EntrantName:     r.EntrantName,
		EntrantNickname: r.EntrantNickname,
		EntrantEmail:    r.EntrantEmail,
	}
}

// createEntryResponse defines the required structure for our create Entry response
type createEntryResponse struct {
	EntrantName  string `json:"entrant_name"`
	EntrantEmail string `json:"entrant_email"`
	LookupURL    string `json:"lookup_url"`
}

// createEntryHandler defines our create Entry request handler
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

		ctx := domain.ContextFromRequest(r)
		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = ctx.GetRealmSeasonID()
		}

		// retrieve the season we need
		season, err := domain.Seasons().GetByID(seasonID)
		if err != nil {
			rest.NotFoundError(fmt.Errorf("invalid season: %s", seasonID)).WriteTo(w)
			return
		}

		// create entry
		createdEntry, err := agent.CreateEntry(ctx, entry, &season, input.PIN)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// generate a lookup URL for the created entry
		lookupURL := fmt.Sprintf("http://%s/%s", r.Host, createdEntry.LookupRef)

		rest.CreatedResponse(&rest.Data{
			Type: "entrant",
			Content: createEntryResponse{
				EntrantName:  createdEntry.EntrantName,
				EntrantEmail: createdEntry.EntrantEmail,
				LookupURL:    lookupURL,
			},
		}, nil).WriteTo(w)
	}
}
