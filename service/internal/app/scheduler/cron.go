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

		job, err := newRetrieveLatestStandingsJob(
			season,
			footballdataorg.NewClient(config.FootballDataAPIToken),
			container,
		)
		if err != nil {
			log.Fatalf("cannot instantiate retrieve latest standings job: %s", err.Error())
		}

		jobs = append(jobs, job)
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

		job, err := newPredictionWindowOpenJob(
			season,
			container,
		)
		if err != nil {
			log.Fatalf("cannot instantiate prediction window open job: %s", err.Error())
		}

		jobs = append(jobs, job)
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

		job, err := newPredictionWindowClosingJob(
			season,
			container,
		)
		if err != nil {
			log.Fatalf("cannot instantiate prediction window closing job: %s", err.Error())
		}

		jobs = append(jobs, job)
	}

	return jobs
}
