package domain

import (
	"context"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"time"
)

// ScoredEntryPredictionAgentInjector defines the dependencies required by our ScoredEntryPredictionAgent
type ScoredEntryPredictionAgentInjector interface {
	MySQL() coresql.Agent
}

// ScoredEntryPredictionAgent defines the behaviours for handling ScoredEntryStandings
type ScoredEntryPredictionAgent struct {
	ScoredEntryPredictionAgentInjector
}

// CreateScoredEntryPrediction handles the creation of a new ScoredEntryPrediction in the database
func (s ScoredEntryPredictionAgent) CreateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction models.ScoredEntryPrediction) (models.ScoredEntryPrediction, error) {
	db := s.MySQL()

	var emptyID uuid.UUID

	if scoredEntryPrediction.EntryPredictionID.String() == emptyID.String() {
		return models.ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"EntryPredictionID is empty",
		}}
	}

	if scoredEntryPrediction.StandingsID.String() == emptyID.String() {
		return models.ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"StandingsID is empty",
		}}
	}

	// ensure that entryPrediction exists
	entryPredictionRepo := repositories.NewEntryPredictionDatabaseRepository(db)
	if err := entryPredictionRepo.ExistsByID(ctx, scoredEntryPrediction.EntryPredictionID.String()); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// ensure that standings exists
	standingsRepo := repositories.NewStandingsDatabaseRepository(db)
	if err := standingsRepo.ExistsByID(ctx, scoredEntryPrediction.StandingsID.String()); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.CreatedAt = time.Now().Truncate(time.Second)
	scoredEntryPrediction.UpdatedAt = sqltypes.NullTime{}

	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(db)

	// write scoredEntryPrediction to database
	if err := scoredEntryPredictionRepo.Insert(ctx, &scoredEntryPrediction); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// RetrieveScoredEntryPredictionByIDs handles the retrieval of an existing ScoredEntryPrediction in the database by its ID
func (s ScoredEntryPredictionAgent) RetrieveScoredEntryPredictionByIDs(ctx context.Context, entryPredictionID, standingsID string) (models.ScoredEntryPrediction, error) {
	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(s.MySQL())

	retrievedScoredEntryPredictions, err := scoredEntryPredictionRepo.Select(ctx, map[string]interface{}{
		"entry_prediction_id": entryPredictionID,
		"standings_id":        standingsID,
	}, false)
	if err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return retrievedScoredEntryPredictions[0], nil
}

// RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber handles the retrieval of
// the most recently created ScoredEntryPrediction by the provided entry ID and round number
func (s ScoredEntryPredictionAgent) RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) (*models.ScoredEntryPrediction, error) {
	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(s.MySQL())

	retrievedScoredEntryPredictions, err := scoredEntryPredictionRepo.SelectByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	// results are already ordered by created date descending
	return &retrievedScoredEntryPredictions[0], nil
}

// UpdateScoredEntryPrediction handles the updating of an existing ScoredEntryPrediction in the database
func (s ScoredEntryPredictionAgent) UpdateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction models.ScoredEntryPrediction) (models.ScoredEntryPrediction, error) {
	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(s.MySQL())

	// ensure the scoredEntryPrediction exists
	if err := scoredEntryPredictionRepo.Exists(
		ctx,
		scoredEntryPrediction.EntryPredictionID.String(),
		scoredEntryPrediction.StandingsID.String(),
	); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.UpdatedAt = sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	// write to database
	if err := scoredEntryPredictionRepo.Update(ctx, &scoredEntryPrediction); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}
