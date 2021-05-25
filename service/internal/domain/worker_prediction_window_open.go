package domain

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PredictionWindowOpenEmailIssuer defines behaviours required to issue a Prediction Window Open email
type PredictionWindowOpenEmailIssuer interface {
	IssuePredictionWindowOpenEmail(ctx context.Context, e *Entry, stf SequencedTimeFrame) error
}

// PredictionWindowOpenWorker performs the work required to notify entries of a new Prediction Window opening
type PredictionWindowOpenWorker struct {
	s     Season
	cl    Clock
	ea    *EntryAgent
	pwoei PredictionWindowOpenEmailIssuer
}

// DoWork implements domain.Worker
func (p *PredictionWindowOpenWorker) DoWork(ctx context.Context) error {
	// previous 24-hour window
	tf := GenerateTimeFrameForPredictionWindowOpenQuery(p.cl.Now())

	// see if a prediction window has opened within this timeframe for the provided season
	window, err := p.s.GetPredictionWindowBeginsWithin(tf)
	if err != nil {
		// no new prediction windows have opened since the last job run
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
func (p *PredictionWindowOpenWorker) IssueEmails(ctx context.Context, entries []Entry, window SequencedTimeFrame) error {
	chDone := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	go func() {
		defer func() { chDone <- struct{}{} }()
		p.issuePredictionWindowOpenEmails(ctx, entries, window, chErr)
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

// issuePredictionWindowOpenEmails issues a series of emails to the provided Entries
func (p *PredictionWindowOpenWorker) issuePredictionWindowOpenEmails(
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

			if err := p.pwoei.IssuePredictionWindowOpenEmail(ctx, &e, window); err != nil {
				chErr <- err
			}
		}(e)
	}

	wg.Wait()
}

func NewPredictionWindowOpenWorker(
	s Season,
	cl Clock,
	ea *EntryAgent,
	ca *CommunicationsAgent,
) (*PredictionWindowOpenWorker, error) {
	if cl == nil {
		return nil, fmt.Errorf("clock: %w", ErrIsNil)
	}
	if ea == nil {
		return nil, fmt.Errorf("entry agent: %w", ErrIsNil)
	}
	if ca == nil {
		return nil, fmt.Errorf("communications agent: %w", ErrIsNil)
	}
	return &PredictionWindowOpenWorker{s, cl, ea, ca}, nil
}

// GenerateTimeFrameForPredictionWindowOpenQuery returns the timeframe required for querying
// Prediction Windows within the PredictionWindowOpen cron job
func GenerateTimeFrameForPredictionWindowOpenQuery(t time.Time) TimeFrame {
	// from 24 hours prior to base time
	// until one minute before base time
	return TimeFrame{
		From:  t.Add(-24 * time.Hour),
		Until: t.Add(-time.Minute),
	}
}
