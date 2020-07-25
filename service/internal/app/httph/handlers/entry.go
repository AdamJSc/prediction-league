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
	entryAgent := domain.EntryAgent{EntryAgentInjector: c}
	commsAgent := domain.CommunicationsAgent{CommunicationsAgentInjector: c}

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
		entry, err := entryAgent.UpdateEntryPaymentDetails(ctx, entryID, input.PaymentMethod, input.PaymentRef)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// issue new entry email
		if err := commsAgent.IssueNewEntryEmail(ctx, &entry); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}

func createEntryPredictionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	agent := domain.EntryAgent{EntryAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntryPredictionRequest

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

		entryPrediction := input.ToEntryPredictionModel()

		domain.GuardFromContext(ctx).SetAttempt(input.EntryShortCode)

		// create entry prediction for entry
		if _, err := agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// success!
		rest.OKResponse(nil, nil).WriteTo(w)
	}
}

func retrieveLatestEntryPredictionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
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

		// get entry prediction that pertains to context timestamp
		entryPrediction, err := agent.RetrieveEntryPredictionByTimestamp(ctx, entry, domain.TimestampFromContext(ctx))
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// get teams that correlate to entry prediction's ranking IDs
		teams, err := domain.FilterTeamsByIDs(entryPrediction.Rankings.GetIDs(), datastore.Teams)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		rest.OKResponse(&rest.Data{
			Type: "entry_prediction",
			Content: retrieveLatestEntryPredictionResponse{
				Teams:       teams,
				LastUpdated: entryPrediction.CreatedAt,
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

func retrieveLatestScoredEntryPrediction(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// parse round number from route
		var roundNumber int
		if err := getRouteParam(r, "round_number", &roundNumber); err != nil {
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

		// get latest scored entry prediction by entry id and round number
		scoredEntryPredictionAgent := domain.ScoredEntryPredictionAgent{
			ScoredEntryPredictionAgentInjector: c,
		}
		scoredEntryPredictions, err := scoredEntryPredictionAgent.RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		// get corresponding standings
		standingsAgent := domain.StandingsAgent{
			StandingsAgentInjector: c,
		}
		standings, err := standingsAgent.RetrieveStandingsByID(ctx, scoredEntryPredictions.StandingsID.String())
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		responseRankings, err := getResponseRankingsFromStandingsRankings(scoredEntryPredictions.Rankings, standings.Rankings)
		if err != nil {
			responseFromError(err).WriteTo(w)
			return
		}

		lastUpdated := standings.CreatedAt
		if standings.UpdatedAt.Valid {
			lastUpdated = standings.UpdatedAt.Time
		}
		rest.OKResponse(&rest.Data{
			Type: "scored",
			Content: retrieveLatestScoredEntryPredictionResponse{
				LastUpdated: lastUpdated,
				RoundScore:  scoredEntryPredictions.Score,
				Rankings:    responseRankings,
			},
		}, nil).WriteTo(w)
	}
}

func getResponseRankingsFromStandingsRankings(
	scoredRankings []models.RankingWithScore,
	standingsRankings []models.RankingWithMeta,
) ([]scoredEntryPredictionResponseRanking, error) {
	var getStandingsPositionForTeamID = func(id string) (int, error) {
		for _, r := range standingsRankings {
			if r.ID == id {
				return r.Position, nil
			}
		}

		return 0, fmt.Errorf("no standings position found for team id: %s", id)
	}

	// find standings position for each scored entry prediction ranking
	var rankingsWithStandingsPosition []scoredEntryPredictionResponseRanking
	for _, r := range scoredRankings {
		var respRanking scoredEntryPredictionResponseRanking

		metaPos, err := getStandingsPositionForTeamID(r.ID)
		if err != nil {
			return nil, err
		}

		respRanking.ID = r.ID
		respRanking.Position = r.Position
		respRanking.Score = r.Score
		respRanking.MetaPosition = metaPos

		rankingsWithStandingsPosition = append(rankingsWithStandingsPosition, respRanking)
	}

	return rankingsWithStandingsPosition, nil
}
