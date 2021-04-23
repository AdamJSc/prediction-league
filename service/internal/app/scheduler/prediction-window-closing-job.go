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

// predictionWindowClosingCronSpec determines the frequency by which the PredictionWindowClosingJob will run
// (i.e. every day at 4:48pm)
const predictionWindowClosingCronSpec = "48 16 * * *"

// newPredictionWindowClosingJob returns a new job that issues emails to entrants
// when an active Prediction Window is due to close for the provided season
func newPredictionWindowClosingJob(season domain.Season, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-closing-%s", season.ID))

	entryAgent := &domain.EntryAgent{
		EntryAgentInjector: injector,
	}

	commsAgent := &domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// from 4:48am tomorrow, until 4:47am the following day
		tf := GenerateTimeFrameForPredictionWindowClosingQuery(time.Now())

		// see if a prediction window is due to close within this timeframe for the provided season
		window, err := season.GetPredictionWindowEndsWithin(tf)
		if err != nil {
			// no active prediction windows are due to close since the last job run
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
		issuePredictionWindowClosingEmails(ctx, entries, window, errChan, commsAgent)

		for err := range errChan {
			log.Println(wrapJobError(
				jobName,
				errors.Wrap(err, "issue email: prediction window closing"),
			))
		}

		// job is complete!
	}

	return &job{
		spec: predictionWindowClosingCronSpec,
		task: task,
	}
}

// issuePredictionWindowClosingEmails issues a series of prediction window closing emails to the provided entries
func issuePredictionWindowClosingEmails(
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

			err := commsAgent.IssuePredictionWindowClosingEmail(ctx, &entry, window)
			if err != nil {
				errChan <- err
			}
		}(entry)
	}

	wg.Wait()
	close(errChan)
}

// GenerateTimeFrameForPredictionWindowClosingQuery returns the timeframe required for querying
// Prediction Windows within the PredictionWindowClosing cron job
func GenerateTimeFrameForPredictionWindowClosingQuery(t time.Time) domain.TimeFrame {
	// from 12 hours after base time
	// until 24 hours after from time, less a minute
	from := t.Add(12 * time.Hour)
	return domain.TimeFrame{
		From:  from,
		Until: from.Add(24 * time.Hour).Add(-time.Minute),
	}
}
