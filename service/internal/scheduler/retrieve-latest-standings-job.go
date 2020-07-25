package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/clients"
	footballdata "prediction-league/service/internal/clients/football-data-org"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"sort"
	"strings"
	"sync"
	"time"
)

// newRetrieveLatestStandingsJob returns a new job that retrieves the latest standings, pertaining to the provided season
func newRetrieveLatestStandingsJob(season models.Season, client clients.FootballDataSource, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("retrieve-latest-standings-%s", season.ID))

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// retrieve current entry predictions for provided season
		currentEntryPredictions, err := retrieveCurrentEntryPredictionsForSeason(ctx, &season, injector)
		if err != nil {
			log.Println(wrapJobError(jobName, err))
			return
		}
		if currentEntryPredictions == nil {
			// no further processing so exit early
			return
		}

		// retrieve and save standings from upstream data source for provided season
		standings, err := retrieveAndSaveStandingsForSeason(ctx, &season, client, injector)
		if err != nil {
			log.Println(wrapJobError(jobName, err))
			return
		}
		if standings == nil {
			// no further processing so exit early
			return
		}

		if standings.RoundNumber == season.MaxRounds && standings.Finalised {
			// we've already finalised the last round of our season so just exit early
			return
		}

		var scoredEntryPredictions []models.ScoredEntryPrediction

		// calculate and save ranking scores for each entry prediction based on the retrieved standings
		for _, entryPrediction := range currentEntryPredictions {
			sep, err := processEntryPredictionWithStandings(ctx, &entryPrediction, standings, injector)
			if err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}

			scoredEntryPredictions = append(scoredEntryPredictions, *sep)
		}

		switch {
		case standings.RoundNumber == season.MaxRounds:
			if !seasonIsComplete(season.MaxRounds, standings.Rankings) {
				break
			}

			// season is complete!
			// let's finalise the current (final) standings round
			standingsAgent := domain.StandingsAgent{
				StandingsAgentInjector: injector,
			}
			standings.Finalised = true
			if _, err := standingsAgent.UpdateStandings(ctx, *standings); err != nil {
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

// retrieveCurrentEntryPredictionsForSeason retrieves all current entry predictions for the provided season
func retrieveCurrentEntryPredictionsForSeason(
	ctx context.Context,
	season *models.Season,
	injector domain.EntryAgentInjector,
) ([]models.EntryPrediction, error) {
	// ensure that season is currently active
	now := time.Now()
	if !season.Active.HasBegunBy(now) || season.Active.HasElapsedBy(now) {
		// season is not currently active so don't carry on (silent fail)
		return nil, nil
	}

	// retrieve all entries for current season
	entriesAgent := domain.EntryAgent{
		EntryAgentInjector: injector,
	}
	seasonEntries, err := entriesAgent.RetrieveEntriesBySeasonID(ctx, season.ID, false)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// no entries for this season yet so don't carry on (silent fail)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "retrieve entries by season id %s", season.ID)
	}

	// get the current entry prediction for each of the entries we've just retrieved
	var currentEntryPredictions []models.EntryPrediction
	for _, entry := range seasonEntries {
		es, err := domain.GetEntryPredictionValidAtTimestamp(entry.EntryPredictions, now)
		if err != nil {
			// error indicates that no prediction has been found, so just ignore this entry and continue to the next
			continue
		}

		currentEntryPredictions = append(currentEntryPredictions, es)
	}

	return currentEntryPredictions, nil
}

// retrieveAndSaveStandingsForSeason retrieves all current entry predictions for the provided season
func retrieveAndSaveStandingsForSeason(
	ctx context.Context,
	season *models.Season,
	client clients.FootballDataSource,
	injector domain.StandingsAgentInjector,
) (*models.Standings, error) {
	// get latest standings from client
	standings, err := client.RetrieveLatestStandingsBySeason(ctx, season)
	if err != nil {
		return nil, errors.Wrapf(err, "retrieve latest standings by season id %s", season.ID)
	}

	// default standings sort (ascending by Rankings[].Position)
	sort.Sort(standings)

	// ensure that all team IDs are valid
	for _, ranking := range standings.Rankings {
		if _, err := datastore.Teams.GetByID(ranking.ID); err != nil {
			return nil, errors.Wrapf(err, "get team by id %s", ranking.ID)
		}
	}

	// save standings to database
	if err := saveStandings(ctx, injector, standings, season.ID); err != nil {
		return nil, errors.Wrapf(err, "save standings with id %s", standings.ID)
	}

	return standings, nil
}

// processEntryPredictionWithStandings scores the provided entry prediction against the provided standings and saves
func processEntryPredictionWithStandings(
	ctx context.Context,
	entryPrediction *models.EntryPrediction,
	standings *models.Standings,
	injector domain.ScoredEntryPredictionAgentInjector,
) (*models.ScoredEntryPrediction, error) {
	standingsRankingCollection := models.NewRankingCollectionFromRankingWithMetas(standings.Rankings)

	rws, err := domain.CalculateRankingsScores(entryPrediction.Rankings, standingsRankingCollection)
	if err != nil {
		return nil, errors.Wrapf(err, "calculate rankings scores for entry prediction with id %s", entryPrediction.ID)
	}

	sep := models.ScoredEntryPrediction{
		EntryPredictionID: entryPrediction.ID,
		StandingsID:       standings.ID,
		Rankings:          rws,
		Score:             rws.GetTotal(),
	}

	// save scored entry prediction
	if err := saveScoredEntryPrediction(ctx, injector, &sep); err != nil {
		return nil, errors.Wrapf(err, "save scored entry prediction with standings id %s and entry prediction id %s", sep.StandingsID, sep.EntryPredictionID)
	}

	return &sep, nil
}

// saveStandings upserts the provided Standings depending on whether or not it already exists.
// This method will also re-point the provided Standings pointer to the previous Standings, if these have not been finalised
// at the point of invocation, so that the rest of the sequence chain will act on the previous Standings instead
func saveStandings(ctx context.Context, injector domain.StandingsAgentInjector, s *models.Standings, seasonID string) error {
	agent := domain.StandingsAgent{
		StandingsAgentInjector: injector,
	}

	existingStandings, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, seasonID, s.RoundNumber)
	switch err.(type) {
	case nil:
		// we have scraped an existing standings round!
		// let's update it...
		existingStandings.Rankings = s.Rankings
		*s = existingStandings
		return updateExistingStandings(ctx, injector, s)
	case domain.NotFoundError:
		// we have scraped a new standings round!
		// we'll handle this in a minute
	default:
		// something went wrong with retrieving our existing standings...
		return err
	}

	// we now know we have a new standings round
	// let's see if we need to finalise the previous one instead
	if s.RoundNumber > 1 {
		previousStandings, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, seasonID, s.RoundNumber-1)
		switch err.(type) {
		case domain.NotFoundError:
			// this should never happen, as we should always have consecutive standings rounds
			// however, just in case, we're fine to carry on as if we don't need to finalise
			// the previous standings, so let's drop through to creating a new one below
		case nil:
			if !previousStandings.Finalised {
				// let's finalise the previous standings and continue on our quest with these instead of our newer scraped standings
				// this means that subsequent methods which create scored entry predictions will do so against the previous standings id,
				// so that we make sure these scores are affiliated with the correct (previous) standings - our newer scraped standings
				// will simply be picked up and created next time the cron job runs instead
				previousStandings.Finalised = true
				*s = previousStandings
				return updateExistingStandings(ctx, injector, s)
			}
		default:
			// something went wrong with retrieving our previous standings...
			return err
		}
	}

	// we're still here!
	// let's create our new standings
	createdStandings, err := agent.CreateStandings(ctx, *s)
	if err != nil {
		return err
	}

	*s = createdStandings
	return nil
}

// updateExistingStandings provides a helper method for updating the provided standings
func updateExistingStandings(ctx context.Context, injector domain.StandingsAgentInjector, standings *models.Standings) error {
	agent := domain.StandingsAgent{
		StandingsAgentInjector: injector,
	}

	updatedStandings, err := agent.UpdateStandings(ctx, *standings)
	if err != nil {
		return err
	}

	*standings = updatedStandings
	return nil
}

// saveScoredEntryPrediction upserts the provided ScoredEntryPrediction depending on whether or not it already exists
func saveScoredEntryPrediction(ctx context.Context, injector domain.ScoredEntryPredictionAgentInjector, ses *models.ScoredEntryPrediction) error {
	agent := domain.ScoredEntryPredictionAgent{
		ScoredEntryPredictionAgentInjector: injector,
	}

	// see if we have an existing scored entry prediction that matches our provided ses
	existingScoredEntryPrediction, err := agent.RetrieveScoredEntryPredictionByIDs(
		ctx,
		ses.EntryPredictionID.String(),
		ses.StandingsID.String(),
	)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// we have a new scored entry prediction!
			// let's create it...
			createdScoredEntryPrediction, err := agent.CreateScoredEntryPrediction(ctx, *ses)
			if err != nil {
				return err
			}

			*ses = createdScoredEntryPrediction
			return nil
		default:
			// something went wrong with retrieving our existing ScoredEntryPrediction...
			return err
		}
	}

	// we have an existing scored entry prediction!
	// let's update it...
	existingScoredEntryPrediction.Rankings = ses.Rankings
	existingScoredEntryPrediction.Score = ses.Score
	updatedScoredEntryPrediction, err := agent.UpdateScoredEntryPrediction(ctx, existingScoredEntryPrediction)
	if err != nil {
		return err
	}

	*ses = updatedScoredEntryPrediction
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

// seasonIsComplete returns true if all of the provided rankings have a games played value that matches the maximum rounds for the season
func seasonIsComplete(maxRounds int, rwm []models.RankingWithMeta) bool {
	for _, r := range rwm {
		// determine whether all teams have played the maximum number of games
		if played, ok := r.MetaData[footballdata.MetaKeyPlayedGames]; !ok || played != maxRounds {
			// no they haven't...
			return false
		}
	}

	// yes they have!
	return true
}