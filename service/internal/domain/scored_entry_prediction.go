package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ScoredEntryPrediction provides a data type for an EntryPrediction that has been scored against a Standings
type ScoredEntryPrediction struct {
	EntryPredictionID uuid.UUID          `db:"entry_prediction_id"`
	StandingsID       uuid.UUID          `db:"standings_id"`
	Rankings          []RankingWithScore `db:"rankings"`
	Score             int                `db:"score"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         *time.Time         `db:"updated_at"`
}

// ScoredEntryPredictionRepository defines the interface for transacting with our ScoredEntryPredictions data source
type ScoredEntryPredictionRepository interface {
	Insert(ctx context.Context, scoredEntryPrediction *ScoredEntryPrediction) error
	Update(ctx context.Context, scoredEntryPrediction *ScoredEntryPrediction) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]ScoredEntryPrediction, error)
	Exists(ctx context.Context, entryPredictionID, standingsID string) error
	SelectEntryCumulativeScoresByRealm(ctx context.Context, realmName string, seasonID string, roundNumber int) ([]LeaderBoardRanking, error)
	SelectByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) ([]ScoredEntryPrediction, error)
}

// ScoredEntryPredictionAgent defines the behaviours for handling ScoredEntryStandings
type ScoredEntryPredictionAgent struct {
	er   EntryRepository
	epr  EntryPredictionRepository
	sr   StandingsRepository
	sepr ScoredEntryPredictionRepository
}

// CreateScoredEntryPrediction handles the creation of a new ScoredEntryPrediction in the database
func (s *ScoredEntryPredictionAgent) CreateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction ScoredEntryPrediction) (ScoredEntryPrediction, error) {
	var emptyID uuid.UUID

	if scoredEntryPrediction.EntryPredictionID.String() == emptyID.String() {
		return ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"EntryPredictionID is empty",
		}}
	}

	if scoredEntryPrediction.StandingsID.String() == emptyID.String() {
		return ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"StandingsID is empty",
		}}
	}

	// ensure that entryPrediction exists
	if err := s.epr.ExistsByID(ctx, scoredEntryPrediction.EntryPredictionID.String()); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// ensure that standings exists
	if err := s.sr.ExistsByID(ctx, scoredEntryPrediction.StandingsID.String()); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.CreatedAt = time.Now().Truncate(time.Second)
	scoredEntryPrediction.UpdatedAt = nil

	// write scoredEntryPrediction to database
	if err := s.sepr.Insert(ctx, &scoredEntryPrediction); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// RetrieveScoredEntryPredictionByIDs handles the retrieval of an existing ScoredEntryPrediction in the database by its ID
func (s *ScoredEntryPredictionAgent) RetrieveScoredEntryPredictionByIDs(ctx context.Context, entryPredictionID, standingsID string) (ScoredEntryPrediction, error) {
	retrievedScoredEntryPredictions, err := s.sepr.Select(ctx, map[string]interface{}{
		"entry_prediction_id": entryPredictionID,
		"standings_id":        standingsID,
	}, false)
	if err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return retrievedScoredEntryPredictions[0], nil
}

// RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber handles the retrieval of
// the most recently created ScoredEntryPrediction by the provided entry ID and round number
func (s *ScoredEntryPredictionAgent) RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) (*ScoredEntryPrediction, error) {
	retrievedScoredEntryPredictions, err := s.sepr.SelectByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	// results are already ordered by created date descending
	return &retrievedScoredEntryPredictions[0], nil
}

// UpdateScoredEntryPrediction handles the updating of an existing ScoredEntryPrediction in the database
func (s *ScoredEntryPredictionAgent) UpdateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction ScoredEntryPrediction) (ScoredEntryPrediction, error) {
	// ensure the scoredEntryPrediction exists
	if err := s.sepr.Exists(
		ctx,
		scoredEntryPrediction.EntryPredictionID.String(),
		scoredEntryPrediction.StandingsID.String(),
	); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	now := time.Now().Truncate(time.Second)
	scoredEntryPrediction.UpdatedAt = &now

	// write to database
	if err := s.sepr.Update(ctx, &scoredEntryPrediction); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// NewScoredEntryPredictionAgent returns a new ScoredEntryPredictionAgent using the provided repositories
func NewScoredEntryPredictionAgent(er EntryRepository, epr EntryPredictionRepository, sr StandingsRepository, sepr ScoredEntryPredictionRepository) (*ScoredEntryPredictionAgent, error) {
	switch {
	case er == nil:
		return nil, fmt.Errorf("entry repository: %w", ErrIsNil)
	case epr == nil:
		return nil, fmt.Errorf("entry prediction repository: %w", ErrIsNil)
	case sr == nil:
		return nil, fmt.Errorf("standings repository: %w", ErrIsNil)
	case sepr == nil:
		return nil, fmt.Errorf("scored entry prediction repository: %w", ErrIsNil)
	}

	return &ScoredEntryPredictionAgent{
		er:   er,
		epr:  epr,
		sr:   sr,
		sepr: sepr,
	}, nil
}
