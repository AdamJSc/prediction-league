package scheduler

import (
	"github.com/robfig/cron/v3"
	"log"
	"prediction-league/service/internal/app/httph"
	footballdata "prediction-league/service/internal/clients/football-data-org"
	"prediction-league/service/internal/datastore"
)

// LoadCron returns our populated cron
func LoadCron(container *httph.HTTPAppContainer) *cron.Cron {
	c := cron.New()

	for _, j := range mustGeneratePredictionWindowOpenJobs(container) {
		c.AddFunc(j.spec, j.task)
	}

	if container.Config().FootballDataAPIToken == "" {
		log.Println("missing config: football data api... scheduled retrieval of latest standings will not run...")
		return c
	}

	for _, j := range mustGenerateRetrieveLatestStandingsJobs(container) {
		c.AddFunc(j.spec, j.task)
	}

	return c
}

// job provides our cron job interface
type job struct {
	spec string
	task func()
}

// mustGenerateRetrieveLatestStandingsJobs generates the RetrieveLatestStandings jobs to be used by the cron
func mustGenerateRetrieveLatestStandingsJobs(container *httph.HTTPAppContainer) []*job {
	config := container.Config()

	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []*job
	for id := range seasonIDs {
		season, err := datastore.Seasons.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		if season.ClientID == nil {
			// skip if there is no client id (e.g. FakeSeason)
			continue
		}

		jobs = append(jobs, newRetrieveLatestStandingsJob(
			season,
			footballdata.NewClient(config.FootballDataAPIToken),
			container,
		))
	}

	return jobs
}

// mustGeneratePredictionWindowOpenJobs generates the PredictionWindowOpen jobs to be used by the cron
func mustGeneratePredictionWindowOpenJobs(container *httph.HTTPAppContainer) []*job {
	config := container.Config()

	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []*job
	for id := range seasonIDs {
		season, err := datastore.Seasons.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		jobs = append(jobs, newPredictionWindowOpenJob(
			season,
			container,
		))
	}

	return jobs
}
