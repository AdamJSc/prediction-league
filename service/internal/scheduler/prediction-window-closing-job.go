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

// predictionWindowClosingCronSpec determines the frequency by which the PredictionWindowClosingJob will run
// (i.e. every day at 4:48pm)
const predictionWindowClosingCronSpec = "48 16 * * *"

// predictionWindowClosingTimeRangeOffset determines the offset of the timeframe lower limit
// (i.e. from 4:48am)
// this is because most prediction windows will finish at 23:59, but we want to give as much notice as possible to entrants
// this enables us to issue an email at 16:48 on 01/01/1970, to give notice of a window ending at 23:59 on 02/01/1970
const predictionWindowClosingTimeRangeOffset = 12 * time.Hour

// predictionWindowClosingTimeRangeUpper determines the upper limit of the timeframe for each cron job run
// (i.e. 24 hours since last job run)
const predictionWindowClosingTimeRangeUpper = 24 * time.Hour

// newPredictionWindowClosingJob returns a new job that issues emails to entrants
// when an active Prediction Window is due to close for the provided season
func newPredictionWindowClosingJob(season models.Season, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-closing-%s", season.ID))

	entryAgent := domain.EntryAgent{
		EntryAgentInjector: injector,
	}

	commsAgent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// determine required timeframe for this job run
		now := time.Now()
		from := now.Add(predictionWindowClosingTimeRangeOffset)
		tf := models.TimeFrame{
			From:  from,
			Until: from.Add(predictionWindowClosingTimeRangeUpper),
		}

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

			err := commsAgent.IssuePredictionWindowClosingEmail(ctx, &entry, window)
			if err != nil {
				errChan <- err
			}
		}(entry)
	}

	wg.Wait()
	close(errChan)
}
