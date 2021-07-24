package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/domain"
)

func createEntryHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input createEntryRequest

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
		season, err := c.seasons.GetByID(seasonID)
		if err != nil {
			notFoundError(fmt.Errorf("invalid season: %s", seasonID)).writeTo(w)
			return
		}

		domain.GuardFromContext(ctx).SetAttempt(input.RealmPIN)

		// create entry
		createdEntry, err := c.entryAgent.CreateEntry(ctx, entry, &season)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		regToken, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeEntryRegistration, createdEntry.ID.String())
		if err != nil {
			c.logger.Errorf("cannot generate entry reg token for entry '%s': %s", createdEntry.ID.String(), err.Error())
			responseFromError(err).writeTo(w)
			return
		}

		createdResponse(&data{
			Type: "entry",
			Content: createEntryResponse{
				ID:                createdEntry.ID.String(),
				Nickname:          createdEntry.EntrantNickname,
				RegistrationToken: regToken.ID,
				NeedsPayment:      c.config.PayPalClientID != "",
			},
		}).writeTo(w)
	}
}

func updateEntryPaymentDetailsHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var input updateEntryPaymentDetailsRequest

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

		isPayPalConfigMissing := c.config.PayPalClientID == ""

		// retrieve registration token
		tkn, err := c.tokenAgent.RetrieveTokenByID(ctx, input.RegistrationToken)
		if err != nil {
			c.logger.Errorf("cannot retrieve token '%s': %s", input.RegistrationToken, err.Error())
			responseFromError(domain.BadRequestError{Err: errors.New("invalid token")}).writeTo(w)
			return
		}

		// check token validity
		if !c.tokenAgent.IsTokenValid(tkn, domain.TokenTypeEntryRegistration, entryID) {
			responseFromError(domain.BadRequestError{Err: errors.New("invalid token")}).writeTo(w)
			return
		}

		// update payment details for entry
		entry, err := c.entryAgent.UpdateEntryPaymentDetails(ctx, entryID, input.PaymentMethod, input.PaymentRef, isPayPalConfigMissing)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// issue new entry email
		if err := c.commsAgent.IssueNewEntryEmail(ctx, &entry, &paymentDetails); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// delete token
		// TODO - feat: replace with token redeem
		if err := c.tokenAgent.DeleteToken(ctx, *tkn); err != nil {
			// log error and continue
			responseFromError(err).writeTo(w)
			return
		}

		// success!
		okResponse(nil).writeTo(w)
	}
}

func createEntryPredictionHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		defer cancel()

		// parse entry ID from route
		var entryID string
		if err := getRouteParam(r, "entry_id", &entryID); err != nil {
			responseFromError(err).writeTo(w)
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
		var input createEntryPredictionRequest
		if err := json.Unmarshal(body, &input); err != nil {
			responseFromError(domain.BadRequestError{Err: err}).writeTo(w)
			return
		}

		// transform to entry prediction model
		newEP := input.ToEntryPredictionModel()

		// get entry
		entry, err := c.entryAgent.RetrieveEntryByID(ctx, entryID)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// check current ranking limit
		limit, err := c.entryAgent.GetPredictionRankingLimit(ctx, entry)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		if err := c.entryAgent.CheckRankingLimit(ctx, limit, newEP, entry); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// set guard attempt for next agent method
		// TODO - ShortCode: migrate auth to use session token
		domain.GuardFromContext(ctx).SetAttempt(input.EntryShortCode)

		// create entry prediction for entry
		if _, err := c.entryAgent.AddEntryPredictionToEntry(ctx, newEP, entry); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// success!
		okResponse(nil).writeTo(w)
	}
}

func retrieveLatestEntryPredictionHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		entry, err := c.entryAgent.RetrieveEntryByID(ctx, entryID)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get entry prediction that pertains to context timestamp
		entryPrediction, err := c.entryAgent.RetrieveEntryPredictionByTimestamp(ctx, entry, c.clock.Now())
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get teams that correlate to entry prediction's ranking IDs
		teams, err := domain.FilterTeamsByIDs(entryPrediction.Rankings.GetIDs(), c.teams)
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

func approveEntryByShortCodeHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	// TODO - ShortCode: migrate to approving by entry id
	return func(w http.ResponseWriter, r *http.Request) {
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
		if _, err := c.entryAgent.ApproveEntryByShortCode(ctx, entryShortCode); err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// success!
		okResponse(nil).writeTo(w)
	}
}

func retrieveLatestScoredEntryPrediction(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		scoredEntryPredictions, err := c.sepAgent.RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}

		// get corresponding standings
		standings, err := c.standingsAgent.RetrieveStandingsByID(ctx, scoredEntryPredictions.StandingsID.String())
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
		if standings.UpdatedAt != nil {
			lastUpdated = *standings.UpdatedAt
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
