package domain

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PredictionWindowClosingEmailIssuer defines behaviours required to issue a Prediction Window Closing email
type PredictionWindowClosingEmailIssuer interface {
	IssuePredictionWindowClosingEmail(ctx context.Context, e *Entry, stf SequencedTimeFrame) error
}

// PredictionWindowClosingWorker performs the work required to notify entries of a new Prediction Window closing
type PredictionWindowClosingWorker struct {
	s     Season
	cl    Clock
	ea    *EntryAgent
	pwcei PredictionWindowClosingEmailIssuer
}

// DoWork implements domain.Worker
func (p *PredictionWindowClosingWorker) DoWork(ctx context.Context) error {
	// previous 24-hour window
	tf := GenerateTimeFrameForPredictionWindowClosingQuery(p.cl.Now())

	// see if a prediction window is due to close within this timeframe for the provided season
	window, err := p.s.GetPredictionWindowEndsWithin(tf)
	if err != nil {
		// no active prediction windows are due to close since the last job run
		// exit early
		return nil
	}

	// retrieve entries for season
	entries, err := p.ea.RetrieveEntriesBySeasonID(ctx, p.s.ID, true)
	if err != nil {
		return fmt.Errorf("cannot retrieve entries for active season id '%s': %w", p.s.ID, err)
	}

	return p.IssueEmails(ctx, entries, window)
}

// IssueEmails issues emails to entrants based on the provided entries and prediction window
func (p *PredictionWindowClosingWorker) IssueEmails(ctx context.Context, entries []Entry, window SequencedTimeFrame) error {
	chDone := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	go func() {
		defer func() { chDone <- struct{}{} }()
		p.issuePredictionWindowClosingEmails(ctx, entries, window, chErr)
	}()

	errs := make([]error, 0)

	for {
		select {
		case err := <-chErr:
			errs = append(errs, err)
		case <-chDone:
			// issue emails complete
			if len(errs) > 0 {
				return MultiError{Errs: errs}
			}
			return nil
		}
	}
}

// issuePredictionWindowClosingEmails issues a series of emails to the provided Entries
func (p *PredictionWindowClosingWorker) issuePredictionWindowClosingEmails(
	ctx context.Context,
	entries []Entry,
	window SequencedTimeFrame,
	chErr chan error,
) {
	sem := make(chan struct{}, 10) // send a maximum of 10 concurrent emails
	wg := &sync.WaitGroup{}
	wg.Add(len(entries))

	for idx := range entries {
		e := entries[idx]
		sem <- struct{}{}

		go func(e Entry) {
			defer func() {
				wg.Done()
				<-sem
			}()

			if err := p.pwcei.IssuePredictionWindowClosingEmail(ctx, &e, window); err != nil {
				chErr <- err
			}
		}(e)
	}

	wg.Wait()
}

func NewPredictionWindowClosingWorker(
	s Season,
	cl Clock,
	ea *EntryAgent,
	ca *CommunicationsAgent,
) (*PredictionWindowClosingWorker, error) {
	if cl == nil {
		return nil, fmt.Errorf("clock: %w", ErrIsNil)
	}
	if ea == nil {
		return nil, fmt.Errorf("entry agent: %w", ErrIsNil)
	}
	if ca == nil {
		return nil, fmt.Errorf("communications agent: %w", ErrIsNil)
	}
	return &PredictionWindowClosingWorker{s, cl, ea, ca}, nil
}

// GenerateTimeFrameForPredictionWindowClosingQuery returns the timeframe required for querying
// Prediction Windows within the PredictionWindowClosing cron job
func GenerateTimeFrameForPredictionWindowClosingQuery(t time.Time) TimeFrame {
	// from 12 hours after base time
	// until 24 hours after from time, less a minute
	from := t.Add(12 * time.Hour)
	return TimeFrame{
		From:  from,
		Until: from.Add(24 * time.Hour).Add(-time.Minute),
	}
}
