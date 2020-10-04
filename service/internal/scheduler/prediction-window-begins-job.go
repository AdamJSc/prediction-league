package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"strings"
	"sync"
	"time"
)

const predictionWindowBeginCronSpec = "5 17 * * *"
const predictionWindowJobFrequency = 24 * time.Hour

// newPredictionWindowBeginsJob returns a new job that issues emails to entrants
// when a new Prediction Window has been opened for the provided season
func newPredictionWindowBeginsJob(season models.Season, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-begins-%s", season.ID))

	entryAgent := domain.EntryAgent{
		EntryAgentInjector: injector,
	}

	commsAgent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// determine required timeframe since last job run
		now := time.Now()
		tf := models.TimeFrame{
			From:  now.Add(-predictionWindowJobFrequency),
			Until: now,
		}

		// see if a prediction window has begun within this timeframe for the provided season
		window, err := season.GetPredictionWindowBeginsWithin(tf)
		if err != nil {
			// no new prediction windows have begun since the last job run
			// exit early
			return
		}

		// retrieve entries for season
		entries, err := entryAgent.RetrieveEntriesBySeasonID(ctx, season.ID, true)
		if err != nil {
			log.Println(wrapJobError(
				jobName,
				errors.Wrapf(err, "retrieve entries for active season id: %s", season.ID),
			))
			return
		}

		var errChan chan error

		// issue our emails
		issuePredictionWindowOpenEmails(ctx, entries, window, errChan, commsAgent)

		for err := range errChan {
			log.Println(wrapJobError(
				jobName,
				errors.Wrap(err, "issue email: prediction window open"),
			))
		}

		// job is complete!
	}

	return &job{
		spec: predictionWindowBeginCronSpec,
		task: task,
	}
}

// issuePredictionWindowOpenEmails issues a series of prediction window open emails to the provided entries
func issuePredictionWindowOpenEmails(
	ctx context.Context,
	entries []models.Entry,
	window models.SequencedTimeFrame,
	errChan chan error,
	commsAgent domain.CommunicationsAgent,
) {
	var wg sync.WaitGroup
	var sem = make(chan struct{}, 10) // send a maximum of 10 concurrent emails

	errChan = make(chan error, len(entries))

	for _, entry := range entries {
		wg.Add(1)
		sem <- struct{}{}

		go func(entry models.Entry) {
			defer wg.Done()
			defer func() { <-sem }()

			err := commsAgent.IssuePredictionWindowOpenEmail(ctx, &entry, window)
			if err != nil {
				errChan <- err
			}
		}(entry)
	}

	wg.Wait()
	close(errChan)
}
