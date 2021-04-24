package scheduler

import (
	"github.com/robfig/cron/v3"
	"log"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
)

// LoadCron returns our populated cron
func LoadCron(container *app.HTTPAppContainer) *cron.Cron {
	c := cron.New()

	for _, j := range mustGeneratePredictionWindowOpenJobs(container) {
		c.AddFunc(j.spec, j.task)
	}

	for _, j := range mustGeneratePredictionWindowClosingJobs(container) {
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
func mustGenerateRetrieveLatestStandingsJobs(container *app.HTTPAppContainer) []*job {
	config := container.Config()

	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []*job
	for id := range seasonIDs {
		season, err := domain.SeasonsDataStore.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		if season.ClientID == nil {
			// skip if there is no client id (e.g. FakeSeason)
			continue
		}

		jobs = append(jobs, newRetrieveLatestStandingsJob(
			season,
			footballdataorg.NewClient(config.FootballDataAPIToken),
			container,
		))
	}

	return jobs
}

// mustGeneratePredictionWindowOpenJobs generates the PredictionWindowOpen jobs to be used by the cron
func mustGeneratePredictionWindowOpenJobs(container *app.HTTPAppContainer) []*job {
	config := container.Config()

	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []*job
	for id := range seasonIDs {
		season, err := domain.SeasonsDataStore.GetByID(id)
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

// mustGeneratePredictionWindowClosingJobs generates the PredictionWindowClosing jobs to be used by the cron
func mustGeneratePredictionWindowClosingJobs(container *app.HTTPAppContainer) []*job {
	config := container.Config()

	// get the current season ID for all realms
	var seasonIDs = make(map[string]struct{})
	for _, realm := range config.Realms {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	// add a job for each unique season ID that retrieves the latest standings
	var jobs []*job
	for id := range seasonIDs {
		season, err := domain.SeasonsDataStore.GetByID(id)
		if err != nil {
			log.Fatal(err)
		}

		jobs = append(jobs, newPredictionWindowClosingJob(
			season,
			container,
		))
	}

	return jobs
}
