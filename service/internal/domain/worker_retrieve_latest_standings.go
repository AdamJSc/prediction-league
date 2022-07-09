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
	season                     Season
	teamCollection             TeamCollection
	clock                      Clock
	logger                     Logger
	entryAgent                 *EntryAgent
	standingsAgent             *StandingsAgent
	scoredEntryPredictionAgent *ScoredEntryPredictionAgent
	matchWeekSubmissionAgent   *MatchWeekSubmissionAgent
	matchWeekResultAgent       *MatchWeekResultAgent
	emailIssuer                RoundCompleteEmailIssuer
	footballClient             FootballDataSource
}

// DoWork implements domain.Worker
func (r *RetrieveLatestStandingsWorker) DoWork(ctx context.Context) error {
	// retrieve entry predictions for provided season
	entryPredictions, err := r.entryAgent.RetrieveEntryPredictionsForActiveSeasonByTimestamp(ctx, r.season, r.clock.Now())
	if err != nil {
		switch {
		case errors.As(err, &NotFoundError{}):
			// no entry predictions found for current season, so exit early
			return fmt.Errorf("no entry predictions")
		case errors.As(err, &ConflictError{}):
			// season is not active, so exit early
			r.logger.Debugf("season is not active: %s", r.season.ID)
			return nil
		default:
			// something else went wrong, so exit early
			return fmt.Errorf("cannot retrieve entry predictions: %w", err)
		}
	}

	// get latest standings from client
	latestStandings, err := r.footballClient.RetrieveLatestStandingsBySeason(ctx, r.season)
	if err != nil {
		return fmt.Errorf("cannot retrieve latest standings: %w", err)
	}

	// validate and sort
	if err := ValidateAndSortStandings(&latestStandings, r.teamCollection); err != nil {
		return fmt.Errorf("cannot validate and sort client standings: %w", err)
	}

	// if standings retrieved from client represents a completed season, ensure that round number reflects the season's
	// max rounds - standings data from upstream client was stuck on round 37 for a 38-round PL season in 2019/20
	// so this check safeguards against that
	if r.season.IsCompletedByStandings(latestStandings) {
		latestStandings.RoundNumber = r.season.MaxRounds
	}

	var jobStandings Standings

	existingStandings, err := r.standingsAgent.RetrieveStandingsBySeasonAndRoundNumber(ctx, r.season.ID, latestStandings.RoundNumber)
	switch {
	case err == nil:
		// we have existing standings
		jobStandings, err = r.ProcessExistingStandings(ctx, existingStandings, latestStandings)
		if err != nil {
			return fmt.Errorf("cannot process existing standings: %w", err)
		}
	case errors.As(err, &NotFoundError{}):
		// we have new standings
		jobStandings, err = r.ProcessNewStandings(ctx, latestStandings)
		if err != nil {
			return fmt.Errorf("cannot process new standings: %w", err)
		}
	default:
		// something went wrong...
		return fmt.Errorf("cannot retrieve standings by round number %d: %w", latestStandings.RoundNumber, err)
	}

	if r.HasFinalisedLastRound(jobStandings) {
		// we've already finalised the last round of our season so just exit early
		return nil
	}

	scoredEntryPredictions := make([]ScoredEntryPrediction, 0)

	// calculate and save ranking scores for each entry prediction based on the standings
	for _, entryPrediction := range entryPredictions {
		sep, err := r.GenerateScoredEntryPrediction(ctx, entryPrediction, jobStandings)
		if err != nil {
			return fmt.Errorf("cannot generate scored entry prediction: %w", err)
		}
		if err := r.upsertScoredEntryPrediction(ctx, sep); err != nil {
			return fmt.Errorf("cannot upsert scored entry prediction: %w", err)
		}
		scoredEntryPredictions = append(scoredEntryPredictions, *sep)
	}

	if r.season.IsCompletedByStandings(jobStandings) {
		// last round of season!
		jobStandings.Finalised = true
		if _, err := r.standingsAgent.UpdateStandings(ctx, jobStandings); err != nil {
			return fmt.Errorf("cannot update finalised standings: %w", err)
		}
	}

	return r.IssueEmails(ctx, jobStandings, scoredEntryPredictions)
}

// ProcessExistingStandings updates the rankings of the provided existing standings then updates them
func (r *RetrieveLatestStandingsWorker) ProcessExistingStandings(
	ctx context.Context,
	existStnd Standings,
	clientStnd Standings,
) (Standings, error) {
	// update rankings
	existStnd.Rankings = clientStnd.Rankings
	return r.standingsAgent.UpdateStandings(ctx, existStnd)
}

// ProcessNewStandings processes the provided Standings as a new entity
func (r *RetrieveLatestStandingsWorker) ProcessNewStandings(
	ctx context.Context,
	stnd Standings,
) (Standings, error) {
	if stnd.RoundNumber == 1 {
		// this is the first time we've scraped our first round
		// just save it!
		return r.standingsAgent.CreateStandings(ctx, stnd)
	}

	// check whether we have a previous round of standings that still needs to be finalised
	rtrvdStnd, err := r.standingsAgent.RetrieveStandingsIfNotFinalised(ctx, r.season.ID, stnd.RoundNumber-1, stnd)
	if err != nil {
		return Standings{}, fmt.Errorf("cannot retrieve existing standings: %w", err)
	}

	if rtrvdStnd.RoundNumber != stnd.RoundNumber {
		// looks like we have unfinished business with our previous standings round
		// let's finalise and update it, then continue with this one
		rtrvdStnd.Finalised = true
		return r.standingsAgent.UpdateStandings(ctx, rtrvdStnd)
	}

	// previous round's standings has already been finalised, so let's create a new one and continue with this
	return r.standingsAgent.CreateStandings(ctx, stnd)
}

// HasFinalisedLastRound returns true if the provided Standings represents the worker's Season having been finalised, otherwise false
func (r *RetrieveLatestStandingsWorker) HasFinalisedLastRound(stnd Standings) bool {
	return r.season.IsCompletedByStandings(stnd) && stnd.Finalised
}

// GenerateScoredEntryPrediction generates a scored entry prediction from the provided entry prediction and standings
func (r *RetrieveLatestStandingsWorker) GenerateScoredEntryPrediction(ctx context.Context, ep EntryPrediction, s Standings) (*ScoredEntryPrediction, error) {
	// TODO: migrate to MatchWeekSubmission entity + deprecate EntryPrediction

	mwSubmission := newMatchWeekSubmissionFromEntryPredictionAndStandings(ep, s)

	// TODO: migrate to MatchWeekStandings entity + deprecate Standings
	// TODO: migrate to MatchWeekResult entity + deprecate ScoredEntryPrediction

	mwStandings := newMatchWeekStandingsFromStandings(s)
	mwResult, err := NewMatchWeekResult(
		mwSubmission.ID,
		BaseScoreModifier(baseScore),
		TeamRankingsHitModifier(mwSubmission, mwStandings),
	)
	if err != nil {
		return nil, err
	}

	if err := r.matchWeekSubmissionAgent.UpsertByLegacy(ctx, mwSubmission); err != nil {
		return nil, err
	}

	mwResult.MatchWeekSubmissionID = mwSubmission.ID
	if err := r.matchWeekResultAgent.UpsertBySubmissionID(ctx, mwResult); err != nil {
		return nil, err
	}

	sep := ScoredEntryPrediction{
		EntryPredictionID: ep.ID,
		StandingsID:       s.ID,
		Rankings:          newRankingsWithScoreFromResultTeamRankings(mwResult.TeamRankings),
		Score:             int(mwResult.Score),
	}

	return &sep, nil
}

// IssueEmails issues emails to entrants based on the provided Standings and ScoredEntryPredictions
func (r *RetrieveLatestStandingsWorker) IssueEmails(ctx context.Context, stnd Standings, seps []ScoredEntryPrediction) error {
	chDone := make(chan struct{}, 1)
	chErr := make(chan error, 1)

	switch {
	case r.season.IsCompletedByStandings(stnd):
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
	existSep, err := r.scoredEntryPredictionAgent.RetrieveScoredEntryPredictionByIDs(
		ctx,
		sep.EntryPredictionID.String(),
		sep.StandingsID.String(),
	)
	if err != nil {
		switch {
		case errors.As(err, &NotFoundError{}):
			// we have a new scored entry prediction!
			// let's create it...
			crSep, crErr := r.scoredEntryPredictionAgent.CreateScoredEntryPrediction(ctx, *sep)
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
	updSep, err := r.scoredEntryPredictionAgent.UpdateScoredEntryPrediction(ctx, existSep)
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

			if err := r.emailIssuer.IssueRoundCompleteEmail(ctx, sep, isFinalRound); err != nil {
				chErr <- err
			}
		}(sep)
	}

	wg.Wait()
}

type RetrieveLatestStandingsWorkerParams struct {
	Season                     Season
	TeamCollection             TeamCollection
	Clock                      Clock
	Logger                     Logger
	EntryAgent                 *EntryAgent
	StandingsAgent             *StandingsAgent
	ScoredEntryPredictionAgent *ScoredEntryPredictionAgent
	MatchWeekSubmissionAgent   *MatchWeekSubmissionAgent
	MatchWeekResultAgent       *MatchWeekResultAgent
	EmailIssuer                RoundCompleteEmailIssuer
	FootballClient             FootballDataSource
}

func NewRetrieveLatestStandingsWorker(params RetrieveLatestStandingsWorkerParams) (*RetrieveLatestStandingsWorker, error) {
	if params.TeamCollection == nil {
		return nil, fmt.Errorf("team collection: %w", ErrIsNil)
	}
	if params.Clock == nil {
		return nil, fmt.Errorf("clock: %w", ErrIsNil)
	}
	if params.Logger == nil {
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	if params.EntryAgent == nil {
		return nil, fmt.Errorf("entry agent: %w", ErrIsNil)
	}
	if params.StandingsAgent == nil {
		return nil, fmt.Errorf("standings agent: %w", ErrIsNil)
	}
	if params.ScoredEntryPredictionAgent == nil {
		return nil, fmt.Errorf("scored entry predictions agent: %w", ErrIsNil)
	}
	if params.MatchWeekSubmissionAgent == nil {
		return nil, fmt.Errorf("match week submission agent: %w", ErrIsNil)
	}
	if params.MatchWeekResultAgent == nil {
		return nil, fmt.Errorf("match week result agent: %w", ErrIsNil)
	}
	if params.EmailIssuer == nil {
		return nil, fmt.Errorf("email issuer: %w", ErrIsNil)
	}
	if params.FootballClient == nil {
		return nil, fmt.Errorf("football data client: %w", ErrIsNil)
	}
	return &RetrieveLatestStandingsWorker{
		season:                     params.Season,
		teamCollection:             params.TeamCollection,
		clock:                      params.Clock,
		logger:                     params.Logger,
		entryAgent:                 params.EntryAgent,
		standingsAgent:             params.StandingsAgent,
		scoredEntryPredictionAgent: params.ScoredEntryPredictionAgent,
		matchWeekSubmissionAgent:   params.MatchWeekSubmissionAgent,
		matchWeekResultAgent:       params.MatchWeekResultAgent,
		emailIssuer:                params.EmailIssuer,
		footballClient:             params.FootballClient,
	}, nil
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
