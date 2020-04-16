package handlers

import (
	"encoding/json"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/core/rest"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/domain"
	"time"
)

type createSeasonRequest struct {
	Name         string    `json:"name"`
	Variant      int       `json:"variant"`
	EntriesFrom  time.Time `json:"entries_from"`
	EntriesUntil time.Time `json:"entries_until"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date""`
}

func (r createSeasonRequest) ToSeasonModel() domain.Season {
	return domain.Season{
		Name:         r.Name,
		EntriesFrom:  r.EntriesFrom,
		EntriesUntil: sqltypes.ToNullTime(r.EntriesUntil),
		StartDate:    r.StartDate,
		EndDate:      r.EndDate,
	}
}

func createSeasonHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.SeasonAgent{SeasonAgentInjector: c}
	return func(w http.ResponseWriter, r *http.Request) {
		var input createSeasonRequest

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

		season := input.ToSeasonModel()
		if err := agent.CreateSeason(r.Context(), &season, input.Variant); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		rest.CreatedResponse(&rest.Data{
			Type:    "season",
			Content: season,
		}, nil).WriteTo(w)
	}
}
