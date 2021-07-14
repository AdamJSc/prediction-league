package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
)

func predictionLoginHandler(c *container) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input predictionLoginRequest

		// read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		defer closeBody(r)

		// parse request body
		if err := json.Unmarshal(body, &input); err != nil {
			responseFromError(domain.BadRequestError{Err: err}).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		// get realm from context
		realm := domain.RealmFromContext(ctx)

		// retrieve entry based on input
		entry, err := retrieveEntryByEmailOrNickname(ctx, input.EmailNickname, realm.SeasonID, realm.Name, c.entryAgent)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				// credentials are invalid so convert to an unauthorized error
				unauthorizedError().writeTo(w)
				return
			}
			responseFromError(err).writeTo(w)
			return
		}

		// does realm name match our entry?
		if domain.RealmFromContext(ctx).Name != entry.RealmName {
			unauthorizedError().writeTo(w)
			return
		}

		// does short code match our entry?
		if entry.ShortCode != input.ShortCode {
			unauthorizedError().writeTo(w)
			return
		}

		// generate a new auth token for our entry, and set it as a cookie
		token, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeAuth, entry.ID.String())
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		setAuthCookieValue(token.ID, w, r)

		okResponse(nil).writeTo(w)
	}
}

func getPredictionPageData(ctx context.Context, authToken string, entryAgent *domain.EntryAgent, tokenAgent *domain.TokenAgent, sc domain.SeasonCollection, tc domain.TeamCollection, cl domain.Clock) view.PredictionPageData {
	var data view.PredictionPageData

	// retrieve season and determine its current state
	seasonID := domain.RealmFromContext(ctx).SeasonID
	season, err := sc.GetByID(seasonID)
	if err != nil {
		data.Err = fmt.Errorf("oops! can't get season: %w", err)
		return data
	}

	ts := cl.Now()

	seasonState := season.GetState(ts)
	data.Predictions.Status = seasonState.PredictionsStatus
	data.Predictions.AcceptedFrom = season.PredictionsAccepted.From
	data.Predictions.AcceptedUntil = season.PredictionsAccepted.Until
	data.Predictions.IsClosing = seasonState.PredictionsClosing

	// default teams IDs should reflect those of the current season
	teamIDs := season.TeamIDs

	if authToken != "" {
		// retrieve the entry ID that the auth token pertains to
		token, err := tokenAgent.RetrieveTokenByID(ctx, authToken)
		if err != nil {
			switch {
			case errors.As(err, &domain.NotFoundError{}):
				data.Err = errors.New("oops! invalid auth token")
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
		entryPrediction, err := entryAgent.RetrieveEntryPredictionByTimestamp(ctx, entry, ts)
		if err == nil {
			// we have an entry prediction, let's capture what we need for our view
			data.Teams.LastUpdated = entryPrediction.CreatedAt
			teamIDs = entryPrediction.Rankings.GetIDs()
		}

		// retrieve prediction ranking limit
		lim, err := entryAgent.GetPredictionRankingLimit(ctx, entry)
		if err != nil {
			data.Err = fmt.Errorf("oops! can't get ranking limit: %w", err)
			return data
		}

		data.Predictions.Limit = lim
	}

	// filter all teams to just the IDs that we need
	teams, err := domain.FilterTeamsByIDs(teamIDs, tc)
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
