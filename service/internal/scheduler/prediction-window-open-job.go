package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"strings"
	"sync"
	"time"
)

// predictionWindowOpenCronSpec determines the frequency by which the PredictionWindowOpenJob will run
// (i.e. every day at 12:34pm)
const predictionWindowOpenCronSpec = "34 12 * * *"

// newPredictionWindowOpenJob returns a new job that issues emails to entrants
// when a new Prediction Window has been opened for the provided season
func newPredictionWindowOpenJob(season domain.Season, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-open-%s", season.ID))

	entryAgent := &domain.EntryAgent{
		EntryAgentInjector: injector,
	}

	commsAgent := &domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// from 12:34pm previous day, until 12:33pm today
		tf := GenerateTimeFrameForPredictionWindowOpenQuery(time.Now())

		// see if a prediction window has opened within this timeframe for the provided season
		window, err := season.GetPredictionWindowBeginsWithin(tf)
		if err != nil {
			// no new prediction windows have opened since the last job run
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
		spec: predictionWindowOpenCronSpec,
		task: task,
	}
}

// issuePredictionWindowOpenEmails issues a series of prediction window open emails to the provided entries
func issuePredictionWindowOpenEmails(
	ctx context.Context,
	entries []domain.Entry,
	window domain.SequencedTimeFrame,
	errChan chan error,
	commsAgent *domain.CommunicationsAgent,
) {
	var wg sync.WaitGroup
	var sem = make(chan struct{}, 10) // send a maximum of 10 concurrent emails

	errChan = make(chan error, len(entries))

	for _, entry := range entries {
		wg.Add(1)
		sem <- struct{}{}

		go func(entry domain.Entry) {
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

// GenerateTimeFrameForPredictionWindowOpenQuery returns the timeframe required for querying
// Prediction Windows within the PredictionWindowOpen cron job
func GenerateTimeFrameForPredictionWindowOpenQuery(t time.Time) domain.TimeFrame {
	// from 24 hours prior to base time
	// until one minute before base time
	return domain.TimeFrame{
		From:  t.Add(-24 * time.Hour),
		Until: t.Add(-time.Minute),
	}
}
