package scheduler

import (
	"context"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/clients"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"sort"
	"time"
)

// retrieveLatestStandingsJob represents our job that retrieves the latest league standings for a given season
type retrieveLatestStandingsJob struct {
	app.MySQLInjector
	season models.Season
	client clients.FootballDataSource
}

func (r retrieveLatestStandingsJob) name() string {
	return "retrieve_latest_standings"
}

func (r retrieveLatestStandingsJob) spec() string {
	// every 15 minutes
	return "15 * * * *"
}

func (r retrieveLatestStandingsJob) function() func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// get latest standings from client
		standings, err := r.client.RetrieveLatestStandingsBySeason(ctx, r.season)
		if err != nil {
			wrapped := errors.Wrapf(err, "retrieve latest standings by season id %s", r.season.ID)
			log.Println(wrapJobError(r, wrapped))
			return
		}

		sort.Sort(standings)

		// ensure that all team IDs are valid
		for _, ranking := range standings.Rankings {
			if _, err := datastore.Teams.GetByID(ranking.ID); err != nil {
				wrapped := errors.Wrapf(err, "get team by id %s", ranking.ID)
				log.Println(wrapJobError(r, wrapped))
				return
			}
		}

		// store standings
		standingsAgent := domain.StandingsAgent{
			StandingsAgentInjector: r,
		}
		if err := saveStandings(ctx, &standings, standingsAgent, r.season.ID); err != nil {
			wrapped := errors.Wrapf(err, "save standings with id %s", standings.ID)
			log.Println(wrapJobError(r, wrapped))
			return
		}

		// retrieve all entries for current season
		entriesAgent := domain.EntryAgent{
			EntryAgentInjector: r,
		}
		seasonEntries, err := entriesAgent.RetrieveEntriesBySeasonID(ctx, r.season.ID)
		if err != nil {
			wrapped := errors.Wrapf(err, "retrieve entries by season id %s", r.season.ID)
			log.Println(wrapJobError(r, wrapped))
			return
		}

		standingsTs := standings.CreatedAt
		if standings.UpdatedAt.Valid {
			standingsTs = standings.UpdatedAt.Time
		}

		// get the current entry selection for each of the entries we've just retrieved
		var currentEntrySelections []models.EntrySelection
		for _, entry := range seasonEntries {
			es, err := domain.GetEntrySelectionValidAtTimestamp(entry.EntrySelections, standingsTs)
			if err != nil {
				wrapped := errors.Wrapf(err, "entry selection for entrant nickname %s", entry.EntrantNickname)
				log.Println(wrapJobError(r, wrapped))
				return
			}

			currentEntrySelections = append(currentEntrySelections, es)
		}

		scoredEntrySelectionAgent := domain.ScoredEntrySelectionAgent{
			ScoredEntrySelectionAgentInjector: r,
		}

		// calculate ranking scores for each entry selection based on the retrieved standings
		standingsRankingCollection := models.NewRankingCollectionFromRankingWithMetas(standings.Rankings)
		for _, es := range currentEntrySelections {
			rws, err := domain.CalculateRankingsScores(es.Rankings, standingsRankingCollection)
			if err != nil {
				wrapped := errors.Wrapf(err, "calculate rankings scores for entry selection with id %s", es.ID)
				log.Println(wrapJobError(r, wrapped))
				return
			}

			ses := models.ScoredEntrySelection{
				EntrySelectionID: es.ID,
				StandingsID:      standings.ID,
				Rankings:         rws,
				Score:            rws.GetTotal(),
			}

			// store scored entry selection
			if err := saveScoredEntrySelection(ctx, &ses, scoredEntrySelectionAgent); err != nil {
				wrapped := errors.Wrapf(err, "save scored entry selection with standings id %s and entry selection id %s", ses.StandingsID, ses.EntrySelectionID)
				log.Println(wrapJobError(r, wrapped))
				return
			}

			// TODO - check for elapsed standings round and send notifications to all entries if elapsed
		}

		log.Println(wrapJobStatus(r, "done!"))
	}
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
