package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/domain"
)

func createEntryHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntryRequest

		// setup agents
		agent, err := domain.NewEntryAgent(c.EntryRepo(), c.EntryPredictionRepo())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

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

		// retrieve our model
		entry := input.ToEntryModel()

		// parse season ID from route
		var seasonID string
		if err := getRouteParam(r, "season_id", &seasonID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		if seasonID == "latest" {
			// use the current realm's season ID instead
			seasonID = domain.RealmFromContext(ctx).SeasonID
		}

		// retrieve the season we need
		season, err := domain.SeasonsDataStore.GetByID(seasonID)
		if err != nil {
			notFoundError(fmt.Errorf("invalid season: %s", seasonID)).writeTo(w)
			return
		}

		domain.GuardFromContext(ctx).SetAttempt(input.RealmPIN)

		// create entry
		createdEntry, err := agent.CreateEntry(ctx, entry, &season)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		createdResponse(&data{
			Type: "entry",
			Content: createEntryResponse{
				ID:           createdEntry.ID.String(),
				Nickname:     createdEntry.EntrantNickname,
				ShortCode:    createdEntry.ShortCode,
				NeedsPayment: c.Config().PayPalClientID != "",
			},
		}).writeTo(w)
	}
}

func updateEntryPaymentDetailsHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	commsAgent := &domain.CommunicationsAgent{CommunicationsAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var input updateEntryPaymentDetailsRequest

		entryAgent, err := domain.NewEntryAgent(c.EntryRepo(), c.EntryPredictionRepo())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

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

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		paymentDetails := domain.PaymentDetails{
			Amount:       input.PaymentAmount,
			Reference:    input.PaymentRef,
			MerchantName: input.MerchantName,
		}

		domain.GuardFromContext(ctx).SetAttempt(input.ShortCode)

		isPayPalConfigMissing := c.Config().PayPalClientID == ""

		// update payment details for entry
		entry, err := entryAgent.UpdateEntryPaymentDetails(ctx, entryID, input.PaymentMethod, input.PaymentRef, isPayPalConfigMissing)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// issue new entry email
		if err := commsAgent.IssueNewEntryEmail(ctx, &entry, &paymentDetails); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// success!
		okResponse(nil).writeTo(w)
	}
}

func createEntryPredictionHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntryPredictionRequest

		entryAgent, err := domain.NewEntryAgent(c.EntryRepo(), c.EntryPredictionRepo())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

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

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		// get entry
		entry, err := entryAgent.RetrieveEntryByID(ctx, entryID)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		entryPrediction := input.ToEntryPredictionModel()

		domain.GuardFromContext(ctx).SetAttempt(input.EntryShortCode)

		// create entry prediction for entry
		if _, err := entryAgent.AddEntryPredictionToEntry(ctx, entryPrediction, entry); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// success!
		okResponse(nil).writeTo(w)
	}
}

func retrieveLatestEntryPredictionHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		entryAgent, err := domain.NewEntryAgent(c.EntryRepo(), c.EntryPredictionRepo())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		// get entry
		entry, err := entryAgent.RetrieveEntryByID(ctx, entryID)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get entry prediction that pertains to context timestamp
		entryPrediction, err := entryAgent.RetrieveEntryPredictionByTimestamp(ctx, entry, domain.TimestampFromContext(ctx))
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get teams that correlate to entry prediction's ranking IDs
		teams, err := domain.FilterTeamsByIDs(entryPrediction.Rankings.GetIDs(), domain.TeamsDataStore)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		okResponse(&data{
			Type: "entry_prediction",
			Content: retrieveLatestEntryPredictionResponse{
				Teams:       teams,
				LastUpdated: entryPrediction.CreatedAt,
			},
		}).writeTo(w)
	}
}

func approveEntryByShortCodeHandler(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		entryAgent, err := domain.NewEntryAgent(c.EntryRepo(), c.EntryPredictionRepo())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

		// parse entry short code from route
		var entryShortCode string
		if err := getRouteParam(r, "entry_short_code", &entryShortCode); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		// approve entry
		if _, err := entryAgent.ApproveEntryByShortCode(ctx, entryShortCode); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// success!
		okResponse(nil).writeTo(w)
	}
}

func retrieveLatestScoredEntryPrediction(c *HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// setup agents
		standingsAgent, err := domain.NewStandingsAgent(c.StandingsRepo())
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		sepAgent, err := domain.NewScoredEntryPredictionAgent(
			c.EntryRepo(),
			c.EntryPredictionRepo(),
			c.StandingsRepo(),
			c.ScoredEntryPredictionRepo(),
		)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// parse round number from route
		var roundNumber int
		if err := getRouteParam(r, "round_number", &roundNumber); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		// get latest scored entry prediction by entry id and round number
		scoredEntryPredictions, err := sepAgent.RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get corresponding standings
		standings, err := standingsAgent.RetrieveStandingsByID(ctx, scoredEntryPredictions.StandingsID.String())
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		responseRankings, err := getResponseRankingsFromStandingsRankings(scoredEntryPredictions.Rankings, standings.Rankings)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		lastUpdated := standings.CreatedAt
		if standings.UpdatedAt.Valid {
			lastUpdated = standings.UpdatedAt.Time
		}
		okResponse(&data{
			Type: "scored",
			Content: retrieveLatestScoredEntryPredictionResponse{
				LastUpdated: lastUpdated,
				RoundScore:  scoredEntryPredictions.Score,
				Rankings:    responseRankings,
			},
		}).writeTo(w)
	}
}

func getResponseRankingsFromStandingsRankings(
	scoredRankings []domain.RankingWithScore,
	standingsRankings []domain.RankingWithMeta,
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
