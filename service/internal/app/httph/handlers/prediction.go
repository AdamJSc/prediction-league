package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/LUSHDigital/core/rest"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
)

func predictionLoginHandler(c *app.HTTPAppContainer) func(http.ResponseWriter, *http.Request) {
	entryAgent := &domain.EntryAgent{EntryAgentInjector: c}
	tokenAgent := &domain.TokenAgent{TokenAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input predictionLoginRequest

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

		// get realm from context
		realm := domain.RealmFromContext(ctx)

		// retrieve entry based on input
		entry, err := retrieveEntryByEmailOrNickname(ctx, input.EmailNickname, realm.SeasonID, realm.Name, entryAgent)
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

		// does realm name match our entry?
		if domain.RealmFromContext(ctx).Name != entry.RealmName {
			rest.UnauthorizedError().WriteTo(w)
			return
		}

		// does short code match our entry?
		if entry.ShortCode != input.ShortCode {
			rest.UnauthorizedError().WriteTo(w)
			return
		}

		// generate a new auth token for our entry, and set it as a cookie
		token, err := tokenAgent.GenerateToken(ctx, domain.TokenTypeAuth, entry.ID.String())
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}
		setAuthCookieValue(token.ID, w, r)

		rest.OKResponse(nil, nil).WriteTo(w)
	}
}

func getPredictionPageData(ctx context.Context, authToken string, entryAgent *domain.EntryAgent, tokenAgent *domain.TokenAgent) view.PredictionPageData {
	var data view.PredictionPageData

	// retrieve season and determine its current state
	seasonID := domain.RealmFromContext(ctx).SeasonID
	season, err := domain.SeasonsDataStore.GetByID(seasonID)
	if err != nil {
		data.Err = err
		return data
	}

	seasonState := season.GetState(domain.TimestampFromContext(ctx))
	data.Predictions.BeingAccepted = seasonState.IsAcceptingPredictions
	if seasonState.NextPredictionsWindow != nil {
		if data.Predictions.BeingAccepted {
			data.Predictions.AcceptedUntil = &seasonState.NextPredictionsWindow.Until
		} else {
			data.Predictions.NextAcceptedFrom = &seasonState.NextPredictionsWindow.From
		}
	}

	// default teams IDs should reflect those of the current season
	teamIDs := season.TeamIDs

	switch {
	case authToken != "":
		// retrieve the entry ID that the auth token pertains to
		token, err := tokenAgent.RetrieveTokenByID(ctx, authToken)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				data.Err = errors.New("invalid auth token")
			default:
				data.Err = err
			}
			return data
		}

		// check that entry id is valid
		entry, err := entryAgent.RetrieveEntryByID(ctx, token.Value)
		if err != nil {
			data.Err = err
			return data
		}

		// we have our entry, let's capture what we need for our view
		data.Entry.ID = entry.ID.String()
		data.Entry.ShortCode = entry.ShortCode

		// if entry has an associated entry prediction
		// then override the default season team IDs with the most recent prediction
		entryPrediction, err := entryAgent.RetrieveEntryPredictionByTimestamp(ctx, entry, domain.TimestampFromContext(ctx))
		if err == nil {
			// we have an entry prediction, let's capture what we need for our view
			data.Teams.LastUpdated = entryPrediction.CreatedAt
			teamIDs = entryPrediction.Rankings.GetIDs()
		}
	}

	// filter all teams to just the IDs that we need
	teams, err := domain.FilterTeamsByIDs(teamIDs, domain.TeamsDataStore)
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

	data.Teams.Raw = string(teamsPayload)

	return data
}
