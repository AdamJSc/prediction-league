package handlers

import (
	"encoding/json"
	"github.com/LUSHDigital/core/rest"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/domain"
	"time"
)

type createEntryRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	PIN      int    `json:"pin"`
}

func (r createEntryRequest) ToEntryModel() domain.Entry {
	return domain.Entry{
		EntrantName:     r.Name,
		EntrantNickname: r.Nickname,
		EntrantEmail:    r.Email,
	}
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

		// retrieve current season
		season, err := domain.Seasons().GetByTimestamp(time.Now())
		if err != nil {
			responseFromError(domain.ValidationError{
				Reasons: []string{"No current season"},
			}).WriteTo(w)
			return
		}

		createdEntry, err := agent.CreateEntryForSeason(domain.ContextFromRequest(r), entry, season, input.PIN)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		rest.CreatedResponse(&rest.Data{
			Type:    "entry",
			Content: createdEntry,
		}, nil).WriteTo(w)
	}
}
