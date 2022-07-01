package domain

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// baseScore is the number of points that each match week result begins with
// prior to deducting any "hits" based on standings/rankings (or applying any subsequent modifiers)
const baseScore = 100

// RoundCompleteEmailIssuer defines behaviours required to issue a Round Complete email
type RoundCompleteEmailIssuer interface {
	IssueRoundCompleteEmail(ctx context.Context, sep ScoredEntryPrediction, isFinalRound bool) error
}

// RetrieveLatestStandingsWorker performs the work required to retrieve the latest standings for a provided Season
type RetrieveLatestStandingsWorker struct {
	s    Season
	tc   TeamCollection
	cl   Clock
	l    Logger
	ea   *EntryAgent
	sa   *StandingsAgent
	sepa *ScoredEntryPredictionAgent
	rcei RoundCompleteEmailIssuer
	fds  FootballDataSource
}

// DoWork implements domain.Worker
func (r *RetrieveLatestStandingsWorker) DoWork(ctx context.Context) error {
	// retrieve entry predictions for provided season
	eps, err := r.ea.RetrieveEntryPredictionsForActiveSeasonByTimestamp(ctx, r.s, r.cl.Now())
	if err != nil {
		switch {
		case errors.As(err, &NotFoundError{}):
			// no entry predictions found for current season, so exit early
			return fmt.Errorf("no entry predictions")
		case errors.As(err, &ConflictError{}):
			// season is not active, so exit early
			r.l.Debugf("season is not active: %s", r.s.ID)
			return nil
		default:
			// something else went wrong, so exit early
			return fmt.Errorf("cannot retrieve entry predictions: %w", err)
		}
	}

	// get latest standings from client
	clientStnd, err := r.fds.RetrieveLatestStandingsBySeason(ctx, r.s)
	if err != nil {
		return fmt.Errorf("cannot retrieve latest standings: %w", err)
	}

	// validate and sort
	if err := ValidateAndSortStandings(&clientStnd, r.tc); err != nil {
		return fmt.Errorf("cannot validate and sort client standings: %w", err)
	}

	// if standings retrieved from client represents a completed season, ensure that round number reflects the season's
	// max rounds - standings data from upstream client was stuck on round 37 for a 38-round PL season in 2019/20
	// so this check safeguards against that
	if r.s.IsCompletedByStandings(clientStnd) {
		clientStnd.RoundNumber = r.s.MaxRounds
	}

	var jobStnd Standings

	existStnd, err := r.sa.RetrieveStandingsBySeasonAndRoundNumber(ctx, r.s.ID, clientStnd.RoundNumber)
	switch {
	case err == nil:
		// we have existing standings
		jobStnd, err = r.ProcessExistingStandings(ctx, existStnd, clientStnd)
		if err != nil {
			return fmt.Errorf("cannot process existing standings: %w", err)
		}
	case errors.As(err, &NotFoundError{}):
		// we have new standings
		jobStnd, err = r.ProcessNewStandings(ctx, clientStnd)
		if err != nil {
			return fmt.Errorf("cannot process new standings: %w", err)
		}
	default:
		// something went wrong...
		return fmt.Errorf("cannot retrieve standings by round number %d: %w", clientStnd.RoundNumber, err)
	}

	if r.HasFinalisedLastRound(jobStnd) {
		// we've already finalised the last round of our season so just exit early
		return nil
	}

	seps := make([]ScoredEntryPrediction, 0)

	// calculate and save ranking scores for each entry prediction based on the standings
	for _, entryPrediction := range eps {
		sep, err := GenerateScoredEntryPrediction(entryPrediction, jobStnd)
		if err != nil {
			return fmt.Errorf("cannot generate scored entry prediction: %w", err)
		}
		if err := r.upsertScoredEntryPrediction(ctx, sep); err != nil {
			return fmt.Errorf("cannot upsert scored entry prediction: %w", err)
		}
		seps = append(seps, *sep)
	}

	if r.s.IsCompletedByStandings(jobStnd) {
		// last round of season!
		jobStnd.Finalised = true
		if _, err := r.sa.UpdateStandings(ctx, jobStnd); err != nil {
			return fmt.Errorf("cannot update finalised standings: %w", err)
		}
	}

	return r.IssueEmails(ctx, jobStnd, seps)
}

// ProcessExistingStandings updates the rankings of the provided existing standings then updates them
func (r *RetrieveLatestStandingsWorker) ProcessExistingStandings(
	ctx context.Context,
	existStnd Standings,
	clientStnd Standings,
) (Standings, error) {
	// update rankings
	existStnd.Rankings = clientStnd.Rankings
	return r.sa.UpdateStandings(ctx, existStnd)
}

// ProcessNewStandings processes the provided Standings as a new entity
func (r *RetrieveLatestStandingsWorker) ProcessNewStandings(
	ctx context.Context,
	stnd Standings,
) (Standings, error) {
	if stnd.RoundNumber == 1 {
		// this is the first time we've scraped our first round
		// just save it!
		return r.sa.CreateStandings(ctx, stnd)
	}

	// check whether we have a previous round of standings that still needs to be finalised
	rtrvdStnd, err := r.sa.RetrieveStandingsIfNotFinalised(ctx, r.s.ID, stnd.RoundNumber-1, stnd)
	if err != nil {
		return Standings{}, fmt.Errorf("cannot retrieve existing standings: %w", err)
	}

	if rtrvdStnd.RoundNumber != stnd.RoundNumber {
		// looks like we have unfinished business with our previous standings round
		// let's finalise and update it, then continue with this one
		rtrvdStnd.Finalised = true
		return r.sa.UpdateStandings(ctx, rtrvdStnd)
	}

	// previous round's standings has already been finalised, so let's create a new one and continue with this
	return r.sa.CreateStandings(ctx, stnd)
}

// HasFinalisedLastRound returns true if the provided Standings represents the worker's Season having been finalised, otherwise false
func (r *RetrieveLatestStandingsWorker) HasFinalisedLastRound(stnd Standings) bool {
	return r.s.IsCompletedByStandings(stnd) && stnd.Finalised
}

// IssueEmails issues emails to entrants based on the provided Standings and ScoredEntryPredictions
func (r *RetrieveLatestStandingsWorker) IssueEmails(ctx context.Context, stnd Standings, seps []ScoredEntryPrediction) error {
	chDone := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	switch {
	case r.s.IsCompletedByStandings(stnd):
		go func() {
			defer func() { chDone <- struct{}{} }()
			r.issueRoundCompleteEmails(ctx, seps, true, chErr)
		}()
	case stnd.Finalised:
		go func() {
			defer func() { chDone <- struct{}{} }()
			r.issueRoundCompleteEmails(ctx, seps, false, chErr)
		}()
	default:
		go func() {
			// no emails to issue
			chDone <- struct{}{}
		}()
	}

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

// upsertScoredEntryPrediction creates or updates the provided ScoredEntryPrediction depending on whether or not it already exists
func (r *RetrieveLatestStandingsWorker) upsertScoredEntryPrediction(ctx context.Context, sep *ScoredEntryPrediction) error {
	// see if we have an existing scored entry prediction that matches our provided sep
	existSep, err := r.sepa.RetrieveScoredEntryPredictionByIDs(
		ctx,
		sep.EntryPredictionID.String(),
		sep.StandingsID.String(),
	)
	if err != nil {
		switch {
		case errors.As(err, &NotFoundError{}):
			// we have a new scored entry prediction!
			// let's create it...
			crSep, crErr := r.sepa.CreateScoredEntryPrediction(ctx, *sep)
			if crErr != nil {
				return fmt.Errorf("cannot created scored entry prediction: %w", crErr)
			}

			*sep = crSep
			return nil
		default:
			// something went wrong with retrieving our existing ScoredEntryPrediction...
			return fmt.Errorf("cannot retrieve scored entry prediction: %w", err)
		}
	}

	// we have an existing scored entry prediction!
	// let's update it...
	existSep.Rankings = sep.Rankings
	existSep.Score = sep.Score
	updSep, err := r.sepa.UpdateScoredEntryPrediction(ctx, existSep)
	if err != nil {
		return fmt.Errorf("cannot update existing scored entry prediction: %w", err)
	}

	*sep = updSep
	return nil
}

// issueRoundCompleteEmails issues a series of round complete emails to the provided scored entry predictions
func (r *RetrieveLatestStandingsWorker) issueRoundCompleteEmails(
	ctx context.Context,
	seps []ScoredEntryPrediction,
	isFinalRound bool,
	chErr chan error,
) {
	sem := make(chan struct{}, 10) // send a maximum of 10 concurrent emails
	wg := &sync.WaitGroup{}
	wg.Add(len(seps))

	for _, sep := range seps {
		sem <- struct{}{}

		go func(sep ScoredEntryPrediction) {
			defer func() {
				wg.Done()
				<-sem
			}()

			if err := r.rcei.IssueRoundCompleteEmail(ctx, sep, isFinalRound); err != nil {
				chErr <- err
			}
		}(sep)
	}

	wg.Wait()
}

func NewRetrieveLatestStandingsWorker(
	s Season,
	tc TeamCollection,
	cl Clock,
	l Logger,
	ea *EntryAgent,
	sa *StandingsAgent,
	sepa *ScoredEntryPredictionAgent,
	ca *CommunicationsAgent,
	fds FootballDataSource,
) (*RetrieveLatestStandingsWorker, error) {
	if tc == nil {
		return nil, fmt.Errorf("team collection: %w", ErrIsNil)
	}
	if cl == nil {
		return nil, fmt.Errorf("clock: %w", ErrIsNil)
	}
	if l == nil {
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	if ea == nil {
		return nil, fmt.Errorf("entry agent: %w", ErrIsNil)
	}
	if sa == nil {
		return nil, fmt.Errorf("standings agent: %w", ErrIsNil)
	}
	if sepa == nil {
		return nil, fmt.Errorf("scored entry predictions agent: %w", ErrIsNil)
	}
	if ca == nil {
		return nil, fmt.Errorf("communications agent: %w", ErrIsNil)
	}
	if fds == nil {
		return nil, fmt.Errorf("football data source: %w", ErrIsNil)
	}
	return &RetrieveLatestStandingsWorker{s, tc, cl, l, ea, sa, sepa, ca, fds}, nil
}

// TODO - tests: move to domain package and remove this constructor
func NewTestRetrieveLatestStandingsWorker(
	s Season,
	tc TeamCollection,
	cl Clock,
	l Logger,
	ea *EntryAgent,
	sa *StandingsAgent,
	sepa *ScoredEntryPredictionAgent,
	rcei RoundCompleteEmailIssuer,
	fds FootballDataSource,
) *RetrieveLatestStandingsWorker {
	return &RetrieveLatestStandingsWorker{s, tc, cl, l, ea, sa, sepa, rcei, fds}
}

// GenerateScoredEntryPrediction generates a scored entry prediction from the provided entry prediction and standings
func GenerateScoredEntryPrediction(ep EntryPrediction, s Standings) (*ScoredEntryPrediction, error) {
	// TODO: migrate to MatchWeekSubmission entity + deprecate EntryPrediction
	mwSubmission := newMatchWeekSubmissionFromEntryPredictionAndStandings(ep, s)

	// TODO: migrate to MatchWeekStandings entity + deprecate Standings
	mwStandings := newMatchWeekStandingsFromStandings(s)
	mwResult, err := NewMatchWeekResult(
		mwSubmission.ID,
		BaseScoreModifier(baseScore),
		TeamRankingsHitModifier(mwSubmission, mwStandings),
	)
	if err != nil {
		return nil, err
	}

	// TODO: migrate to MatchWeekResult entity + deprecate ScoredEntryPrediction

	sep := ScoredEntryPrediction{
		EntryPredictionID: ep.ID,
		StandingsID:       s.ID,
		Rankings:          newRankingsWithScoreFromResultTeamRankings(mwResult.TeamRankings),
		Score:             int(mwResult.Score),
	}

	return &sep, nil
}

// ValidateAndSortStandings sorts and validates the provided standings
func ValidateAndSortStandings(stnd *Standings, tc TeamCollection) error {
	if stnd == nil {
		return fmt.Errorf("standings: %w", ErrIsNil)
	}

	// ensure that all team IDs are valid
	for _, rnk := range stnd.Rankings {
		if _, err := tc.GetByID(rnk.ID); err != nil {
			return NotFoundError{err}
		}
	}

	// default standings sort (ascending by Rankings[].Position)
	sort.Sort(stnd)

	return nil
}
