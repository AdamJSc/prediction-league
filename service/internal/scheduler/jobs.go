package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/clients"
	footballdata "prediction-league/service/internal/clients/football-data-org"
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
	teams := domain.Teams()
	for _, ranking := range standings.Rankings {
		if _, err := teams.GetByResourceID(ranking.ID); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("team id '%s':", ranking.ID.Value()))
		}
	}

	// TODO - store Standings

	// TODO - retrieve all Entries for r.Season

	// TODO - generate new Standings score per team for each entry based on standings

	return fmt.Sprint("%+v", standings), nil
}

// MustGenerateCronJobs generates the cron jobs to be used by the scheduler
func MustGenerateCronJobs(config domain.Config) []Job {
	seasons := domain.Seasons()

	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []Job
	for id := range seasonIDs {
		season, err := seasons.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		jobs = append(jobs, RetrieveLatestStandings{
			Season:   season,
			Client: footballdata.NewClient(config.FootballDataAPIToken),
		})
	}

	return jobs
}
