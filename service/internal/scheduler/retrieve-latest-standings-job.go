package scheduler

import (
	"context"
	"fmt"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/clients"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"strings"
	"sync"
	"time"
)

// newRetrieveLatestStandingsJob returns a new job that retrieves the latest standings, pertaining to the provided season
func newRetrieveLatestStandingsJob(season models.Season, client clients.FootballDataSource, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("retrieve-latest-standings-%s", season.ID))

	standingsAgent := domain.StandingsAgent{
		StandingsAgentInjector: injector,
	}

	entryAgent := domain.EntryAgent{
		EntryAgentInjector: injector,
	}

	sepAgent := domain.ScoredEntryPredictionAgent{
		ScoredEntryPredictionAgentInjector: injector,
	}

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// retrieve entry predictions for provided season
		latestEntryPredictions, err := entryAgent.RetrieveEntryPredictionsForActiveSeasonByTimestamp(
			ctx,
			season,
			nil, // defaults to current timestamp
		)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				// no entry predictions found for current season, so exit early
				return
			case domain.ConflictError:
				// season is not active, so exit early
				return
			default:
				// something else went wrong, so exit early
				log.Println(wrapJobError(jobName, err))
				return
			}
		}

		// get latest standings from client
		clientStandings, err := client.RetrieveLatestStandingsBySeason(ctx, &season)
		if err != nil {
			log.Println(wrapJobError(jobName, err))
			return
		}
		// validate and sort
		if err := domain.ValidateAndSortStandings(clientStandings); err != nil {
			log.Println(wrapJobError(jobName, err))
			return
		}

		var standings models.Standings

		existingStandings, err := standingsAgent.RetrieveStandingsBySeasonAndRoundNumber(ctx, season.ID, clientStandings.RoundNumber)
		switch err.(type) {
		case nil:
			// we have existing standings
			standings, err = processExistingStandings(ctx, existingStandings, *clientStandings, standingsAgent)
			if err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}
		case domain.NotFoundError:
			// we have new standings
			standings, err = processNewStandings(ctx, *clientStandings, season, standingsAgent)
			if err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}
		default:
			// something went wrong...
			log.Println(wrapJobError(jobName, err))
			return
		}

		if season.IsComplete(standings) && standings.Finalised {
			// we've already finalised the last round of our season so just exit early
			return
		}

		var scoredEntryPredictions []models.ScoredEntryPrediction

		// calculate and save ranking scores for each entry prediction based on the standings
		for _, entryPrediction := range latestEntryPredictions {
			sep, err := domain.ScoreEntryPredictionBasedOnStandings(entryPrediction, standings)
			if err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}
			if err := upsertScoredEntryPrediction(ctx, sep, sepAgent); err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}
			scoredEntryPredictions = append(scoredEntryPredictions, *sep)
		}

		switch {
		case season.IsComplete(standings):
			// finalise final round
			standings.Finalised = true
			if _, err := standingsAgent.UpdateStandings(ctx, standings); err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}

			// issue final round complete emails
			issueRoundCompleteEmails(ctx, injector, true, scoredEntryPredictions, jobName)

		case standings.Finalised:
			// issue round complete emails
			issueRoundCompleteEmails(ctx, injector, false, scoredEntryPredictions, jobName)
		}

		// job is complete!
	}

	return &job{
		spec: "@every 0h15m",
		task: task,
	}
}

func processExistingStandings(
	ctx context.Context,
	existingStandings models.Standings,
	clientStandings models.Standings,
	standingsAgent domain.StandingsAgent,
) (models.Standings, error) {
	// update rankings
	existingStandings.Rankings = clientStandings.Rankings
	return standingsAgent.UpdateStandings(ctx, existingStandings)
}

func processNewStandings(
	ctx context.Context,
	clientStandings models.Standings,
	season models.Season,
	standingsAgent domain.StandingsAgent,
) (models.Standings, error) {
	if clientStandings.RoundNumber == 1 {
		// this is the first time we've scraped our first round
		// just save it!
		return standingsAgent.CreateStandings(ctx, clientStandings)
	}

	// check whether we have a previous round of standings that still needs to be finalised
	retrievedStandings, err := standingsAgent.RetrieveStandingsIfNotFinalised(ctx, season.ID, clientStandings.RoundNumber-1, clientStandings)
	if err != nil {
		return models.Standings{}, err
	}

	if retrievedStandings.RoundNumber != clientStandings.RoundNumber {
		// looks like we have unfinished business with our previous standings round
		// let's finalise and update it, then continue with this
		retrievedStandings.Finalised = true
		return standingsAgent.UpdateStandings(ctx, retrievedStandings)
	}

	// previous round's standings has already been finalised, so let's create a new one and continue with this
	return standingsAgent.CreateStandings(ctx, clientStandings)
}

// upsertScoredEntryPrediction creates or updates the provided ScoredEntryPrediction depending on whether or not it already exists
func upsertScoredEntryPrediction(ctx context.Context, sep *models.ScoredEntryPrediction, sepAgent domain.ScoredEntryPredictionAgent) error {
	// see if we have an existing scored entry prediction that matches our provided sep
	existingScoredEntryPrediction, err := sepAgent.RetrieveScoredEntryPredictionByIDs(
		ctx,
		sep.EntryPredictionID.String(),
		sep.StandingsID.String(),
	)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// we have a new scored entry prediction!
			// let's create it...
			createdScoredEntryPrediction, err := sepAgent.CreateScoredEntryPrediction(ctx, *sep)
			if err != nil {
				return err
			}

			*sep = createdScoredEntryPrediction
			return nil
		default:
			// something went wrong with retrieving our existing ScoredEntryPrediction...
			return err
		}
	}

	// we have an existing scored entry prediction!
	// let's update it...
	existingScoredEntryPrediction.Rankings = sep.Rankings
	existingScoredEntryPrediction.Score = sep.Score
	updatedScoredEntryPrediction, err := sepAgent.UpdateScoredEntryPrediction(ctx, existingScoredEntryPrediction)
	if err != nil {
		return err
	}

	*sep = updatedScoredEntryPrediction
	return nil
}

// issueRoundCompleteEmails issues a series of round complete emails to the provided scored entry predictions
func issueRoundCompleteEmails(
	ctx context.Context,
	injector domain.CommunicationsAgentInjector,
	finalRound bool,
	scoredEntryPredictions []models.ScoredEntryPrediction,
	jobName string,
) bool {
	commsAgent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	var wg sync.WaitGroup
	var sem = make(chan struct{}, 10)
	var chErrs = make(chan error, len(scoredEntryPredictions))

	for _, sep := range scoredEntryPredictions {
		wg.Add(1)
		sem <- struct{}{}

		go func(pred models.ScoredEntryPrediction) {
			defer wg.Done()
			defer func() { <-sem }()

			err := commsAgent.IssueRoundCompleteEmail(ctx, &pred, finalRound)
			if err != nil {
				chErrs <- err
			}
		}(sep)
	}

	wg.Wait()
	close(chErrs)

	if len(chErrs) == 0 {
		return true
	}

	for err := range chErrs {
		log.Println(wrapJobError(jobName, err))
	}
	return false
}
