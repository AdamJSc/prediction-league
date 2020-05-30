package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/clients"
	footballdata "prediction-league/service/internal/clients/football-data-org"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"sort"
)

// Job defines the interface that a job must have
type Job interface {
	Name() string
	IntervalInSeconds() int
	Run(ctx context.Context) (string, error)
}

// RetrieveLatestStandings represents our job that retrieves the latest league standings for a given season
type RetrieveLatestStandings struct {
	app.MySQLInjector
	Season models.Season
	Client clients.FootballDataSource
}

func (r RetrieveLatestStandings) Name() string {
	return "retrieve_latest_standings"
}

func (r RetrieveLatestStandings) IntervalInSeconds() int {
	// 15 minutes
	return 900
}

func (r RetrieveLatestStandings) Run(ctx context.Context) (string, error) {
	// get latest standings from client
	standings, err := r.Client.RetrieveLatestStandingsBySeason(ctx, r.Season)
	if err != nil {
		return "", err
	}

	sort.Sort(standings)

	// ensure that all team IDs are valid
	for _, ranking := range standings.Rankings {
		if _, err := datastore.Teams.GetByID(ranking.ID); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("team id '%s':", ranking.ID))
		}
	}

	// store standings
	standingsAgent := domain.StandingsAgent{
		StandingsAgentInjector: r,
	}
	if err := saveStandings(ctx, &standings, standingsAgent, r.Season.ID); err != nil {
		return "", err
	}

	// retrieve all entries for current season
	entriesAgent := domain.EntryAgent{
		EntryAgentInjector: r,
	}
	seasonEntries, err := entriesAgent.RetrieveEntriesBySeasonID(ctx, r.Season.ID)
	if err != nil {
		return "", err
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
			return "", errors.Wrapf(err, "entry selection for entrant nickname: %s", entry.EntrantNickname)
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
			return "", err
		}

		ses := models.ScoredEntrySelection{
			EntrySelectionID: es.ID,
			StandingsID:      standings.ID,
			Rankings:         rws,
			Score:            rws.GetTotal(),
		}

		// store scored entry selection
		if err := saveScoredEntrySelection(ctx, &ses, scoredEntrySelectionAgent); err != nil {
			return "", err
		}

		// TODO - check for elapsed standings round and send notifications to all entries if elapsed
	}

	return "done!", nil
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

// MustGenerateCronJobs generates the cron jobs to be used by the scheduler
func MustGenerateCronJobs(config domain.Config, container *httph.HTTPAppContainer) []Job {
	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []Job
	for id := range seasonIDs {
		season, err := datastore.Seasons.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		jobs = append(jobs, RetrieveLatestStandings{
			MySQLInjector: container,
			Season:        season,
			Client:        footballdata.NewClient(config.FootballDataAPIToken),
		})
	}

	return jobs
}
