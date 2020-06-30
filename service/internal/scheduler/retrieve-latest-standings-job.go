package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/clients"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"sort"
	"strings"
	"time"
)

// newRetrieveLatestStandingsJob returns a new job that retrieves the latest standings, pertaining to the provided season
func newRetrieveLatestStandingsJob(season models.Season, client clients.FootballDataSource, injector app.MySQLInjector) job {
	jobName := strings.ToLower(fmt.Sprintf("retrieve-latest-standings-%s", season.ID))

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// retrieve current entry selections for provided season
		currentEntrySelections, err := retrieveCurrentEntrySelectionsForSeason(ctx, &season, injector)
		if err != nil {
			log.Println(wrapJobError(jobName, err))
			return
		}
		if currentEntrySelections == nil {
			// silent fail
			return
		}

		// retrieve and save standings from upstream data source for provided season
		standings, err := retrieveAndSaveStandingsForSeason(ctx, &season, client, injector)
		if err != nil {
			log.Println(wrapJobError(jobName, err))
			return
		}
		if standings == nil {
			// silent fail
			return
		}

		// calculate and save ranking scores for each entry selection based on the retrieved standings
		for _, entrySelection := range currentEntrySelections {
			if err := processEntrySelectionWithStandings(ctx, &entrySelection, standings, injector); err != nil {
				log.Println(wrapJobError(jobName, err))
				return
			}
		}

		log.Println(wrapJobStatus(jobName, "done!"))
	}

	return job{
		name: jobName,
		spec: "15 * * * *",
		task: task,
	}
}

// retrieveCurrentEntrySelectionsForSeason retrieves all current entry selections for the provided season
func retrieveCurrentEntrySelectionsForSeason(
	ctx context.Context,
	season *models.Season,
	injector app.MySQLInjector,
) ([]models.EntrySelection, error) {
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
	seasonEntries, err := entriesAgent.RetrieveEntriesBySeasonID(ctx, season.ID)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// no entries for this season yet so don't carry on (silent fail)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "retrieve entries by season id %s", season.ID)
	}

	// get the current entry selection for each of the entries we've just retrieved
	var currentEntrySelections []models.EntrySelection
	for _, entry := range seasonEntries {
		es, err := domain.GetEntrySelectionValidAtTimestamp(entry.EntrySelections, now)
		if err != nil {
			return nil, errors.Wrapf(err, "entry selection for entrant nickname %s", entry.EntrantNickname)
		}

		currentEntrySelections = append(currentEntrySelections, es)
	}

	return currentEntrySelections, nil
}

// retrieveAndSaveStandingsForSeason retrieves all current entry selections for the provided season
func retrieveAndSaveStandingsForSeason(
	ctx context.Context,
	season *models.Season,
	client clients.FootballDataSource,
	injector app.MySQLInjector,
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
	standingsAgent := domain.StandingsAgent{
		StandingsAgentInjector: injector,
	}
	if err := saveStandings(ctx, standings, standingsAgent, season.ID); err != nil {
		return nil, errors.Wrapf(err, "save standings with id %s", standings.ID)
	}

	return standings, nil
}

// processEntrySelectionWithStandings
func processEntrySelectionWithStandings(
	ctx context.Context,
	entrySelection *models.EntrySelection,
	standings *models.Standings,
	injector app.MySQLInjector,
) error {
	standingsRankingCollection := models.NewRankingCollectionFromRankingWithMetas(standings.Rankings)

	rws, err := domain.CalculateRankingsScores(entrySelection.Rankings, standingsRankingCollection)
	if err != nil {
		return errors.Wrapf(err, "calculate rankings scores for entry selection with id %s", entrySelection.ID)
	}

	ses := models.ScoredEntrySelection{
		EntrySelectionID: entrySelection.ID,
		StandingsID:      standings.ID,
		Rankings:         rws,
		Score:            rws.GetTotal(),
	}

	// store scored entry selection
	scoredEntrySelectionAgent := domain.ScoredEntrySelectionAgent{
		ScoredEntrySelectionAgentInjector: injector,
	}
	if err := saveScoredEntrySelection(ctx, &ses, scoredEntrySelectionAgent); err != nil {
		return errors.Wrapf(err, "save scored entry selection with standings id %s and entry selection id %s", ses.StandingsID, ses.EntrySelectionID)
	}

	// TODO - check for elapsed standings round and send notifications to all entries if elapsed

	return nil
}

// saveStandings upserts the provided Standings depending on whether or not it already exists
func saveStandings(ctx context.Context, s *models.Standings, agent domain.StandingsAgent, seasonID string) error {
	existingStandings, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, seasonID, s.RoundNumber)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// we have scraped a new standings round!
			// let's create it...
			createdStandings, err := agent.CreateStandings(ctx, *s)
			if err != nil {
				return err
			}

			*s = createdStandings
			return nil
		default:
			// something went wrong with retrieving our existing standings...
			return err
		}
	}

	// we have scraped an existing standings round!
	// let's update it...
	existingStandings.Rankings = s.Rankings
	updatedStandings, err := agent.UpdateStandings(ctx, existingStandings)
	if err != nil {
		return err
	}

	*s = updatedStandings
	return nil
}

// saveScoredEntrySelection upserts the provided ScoredEntrySelection depending on whether or not it already exists
func saveScoredEntrySelection(ctx context.Context, ses *models.ScoredEntrySelection, agent domain.ScoredEntrySelectionAgent) error {
	existingScoredEntrySelection, err := agent.RetrieveScoredEntrySelectionByIDs(ctx, ses.EntrySelectionID.String(), ses.StandingsID.String())
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// we have a new scored entry selection!
			// let's create it...
			createdScoredEntrySelection, err := agent.CreateScoredEntrySelection(ctx, *ses)
			if err != nil {
				return err
			}

			*ses = createdScoredEntrySelection
			return nil
		default:
			// something went wrong with retrieving our existing ScoredEntrySelection...
			return err
		}
	}

	// we have an existing scored entry selection!
	// let's update it...
	existingScoredEntrySelection.Rankings = ses.Rankings
	existingScoredEntrySelection.Score = ses.Score
	updatedScoredEntrySelection, err := agent.UpdateScoredEntrySelection(ctx, existingScoredEntrySelection)
	if err != nil {
		return err
	}

	*ses = updatedScoredEntrySelection
	return nil
}
