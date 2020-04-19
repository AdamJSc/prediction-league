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
	EntrantName  string `json:"entrant_name"`
	EntrantEmail string `json:"entrant_email"`
	LookupURL    string `json:"lookup_url"`
}

func createEntryHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntryRequest

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
		defer closeBody(r)

		if err := json.Unmarshal(body, &input); err != nil {
			responseFromError(domain.BadRequestError{Err: err}).WriteTo(w)
			return
		}

		entry := input.ToEntryModel()

		var seasonID string
		if err := getRouteParam(r, "season_id", &seasonID); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		ctx := domain.ContextFromRequest(r)

		if seasonID == "latest" {
			seasonID = ctx.GetRealmSeasonID()
		}

		season, err := domain.Seasons().GetByID(seasonID)
		if err != nil {
			rest.NotFoundError(fmt.Errorf("invalid season: %s", seasonID)).WriteTo(w)
			return
		}

		createdEntry, err := agent.CreateEntry(ctx, entry, &season, input.PIN)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

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
