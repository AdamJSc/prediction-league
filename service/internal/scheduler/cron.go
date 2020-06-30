package scheduler

import (
	"log"
	"prediction-league/service/internal/app/httph"
	footballdata "prediction-league/service/internal/clients/football-data-org"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"

	"github.com/robfig/cron/v3"
)

// LoadCron returns our populated cron
func LoadCron(config domain.Config, container *httph.HTTPAppContainer) *cron.Cron {
	c := cron.New()

	for _, j := range mustGenerateRetrieveLatestStandingsJobs(config, container) {
		c.AddFunc(j.spec, j.task)
	}

	return c
}

// job provides our cron job interface
type job struct {
	name string
	spec string
	task func()
}

// mustGenerateRetrieveLatestStandingsJobs generates the RetrieveLatestStandings jobs to be used by the cron
func mustGenerateRetrieveLatestStandingsJobs(config domain.Config, container *httph.HTTPAppContainer) []job {
	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []job
	for id := range seasonIDs {
		season, err := datastore.Seasons.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		jobs = append(jobs, newRetrieveLatestStandingsJob(
			season,
			footballdata.NewClient(config.FootballDataAPIToken),
			container,
		))
	}

	return jobs
}
