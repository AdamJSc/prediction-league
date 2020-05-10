package scheduler

import (
	"context"
	"prediction-league/service/internal/domain"
)

// Job defines the interface that a job must have
type Job interface {
	Name() string
	IntervalInSeconds() int
	Run(ctx context.Context) (string, error)
}

// RetrieveLatestLeagueStandings represents our job that retrieves the latest league standings for a given season
type RetrieveLatestLeagueStandings struct {
	SeasonID string
}

func (r RetrieveLatestLeagueStandings) Name() string {
	return "retrieve_latest_standings"
}

func (r RetrieveLatestLeagueStandings) IntervalInSeconds() int {
	// 15 minutes
	return 900
}

func (r RetrieveLatestLeagueStandings) Run(_ context.Context) (string, error) {
	return "to implement...", nil
}

// GenerateCronJobs generates the cron jobs to be used by the scheduler
func GenerateCronJobs(config domain.Config) []Job {
	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest league standings
	var jobs []Job
	for id := range seasonIDs {
		jobs = append(jobs, RetrieveLatestLeagueStandings{SeasonID: id})
	}

	return jobs
}
