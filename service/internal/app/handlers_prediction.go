package app

import (
	"context"
	"encoding/json"
	"errors"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
	"time"
)

func getPredictionPageData(ctx context.Context, authTknID string, entryAgent *domain.EntryAgent, tokenAgent *domain.TokenAgent, sc domain.SeasonCollection, tc domain.TeamCollection, cl domain.Clock, l domain.Logger) view.PredictionPageData {
	var data view.PredictionPageData

	// retrieve season and determine its current state
	seasonID := domain.RealmFromContext(ctx).SeasonID
	season, err := sc.GetByID(seasonID)
	if err != nil {
		l.Errorf("prediction page: can't get season: %s", err.Error())
		data.Err = genericErr
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

	if authTknID != "" {
		// enrich based on auth token
		teamIDs, err = enrichAuthPredictionPageData(ctx, authTknID, teamIDs, &data, ts, entryAgent, tokenAgent, l)
		if err != nil {
			data.Err = err
			return data
		}
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

func enrichAuthPredictionPageData(ctx context.Context, authTknID string, teamIDs []string, data *view.PredictionPageData, ts time.Time, ea *domain.EntryAgent, ta *domain.TokenAgent, l domain.Logger) ([]string, error) {
	// retrieve the entry ID that the auth token pertains to
	authTkn, err := ta.RetrieveTokenByID(ctx, authTknID)
	if err != nil {
		if errors.As(err, &domain.NotFoundError{}) {
			l.Errorf("prediction page: auth token not found '%s'", authTknID)
			return nil, genericErr
		}
		l.Errorf("prediction page: error retrieving auth token '%s': %s", authTknID, err)
		return nil, genericErr
	}

	// check that entry id is valid
	entry, err := ea.RetrieveEntryByID(ctx, authTkn.Value)
	if err != nil {
		l.Errorf("prediction page: cannot retrieve entry id '%s' from auth token '%s': %s", authTkn.Value, authTkn.ID, err.Error())
		return nil, genericErr
	}

	// we have our entry, let's capture what we need for our view
	data.Entry.ID = entry.ID.String()

	// if entry has an associated entry prediction
	// then override the default season team IDs with the most recent prediction
	entryPrediction, err := ea.RetrieveEntryPredictionByTimestamp(ctx, entry, ts)
	if err == nil {
		// we have an entry prediction, let's capture what we need for our view
		data.Teams.LastUpdated = entryPrediction.CreatedAt
		teamIDs = entryPrediction.Rankings.GetIDs()
	}

	// retrieve prediction ranking limit
	lim, err := ea.GetPredictionRankingLimit(ctx, entry)
	if err != nil {
		l.Errorf("prediction page: cannot get ranking limit: %s", err.Error())
		return nil, genericErr
	}

	data.Predictions.Limit = lim

	if lim == 0 {
		// no predictions are allowed to be changed
		// exit early before generating prediction token
		return teamIDs, nil
	}

	// purge unused prediction tokens for this user
	if _, err := ta.DeleteInFlightTokens(ctx, domain.TokenTypePrediction, entry.ID.String()); err != nil {
		l.Errorf("prediction page: cannot purge in-flight prediction tokens: %s", err.Error())
		return nil, genericErr
	}

	// generate new prediction token
	predTkn, err := ta.GenerateToken(ctx, domain.TokenTypePrediction, entry.ID.String())
	if err != nil {
		l.Errorf("prediction page: cannot generate prediction token: %s", err.Error())
		return nil, genericErr
	}

	data.Entry.PredictionToken = predTkn.ID

	return teamIDs, nil
}
