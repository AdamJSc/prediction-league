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

// ScoredEntrySelectionAgentInjector defines the dependencies required by our ScoredEntrySelectionAgent
type ScoredEntrySelectionAgentInjector interface {
	MySQL() coresql.Agent
}

// ScoredEntrySelectionAgent defines the behaviours for handling ScoredEntryStandings
type ScoredEntrySelectionAgent struct {
	ScoredEntrySelectionAgentInjector
}

// CreateScoredEntrySelection handles the creation of a new ScoredEntrySelection in the database
func (s ScoredEntrySelectionAgent) CreateScoredEntrySelection(ctx context.Context, scoredEntrySelection models.ScoredEntrySelection) (models.ScoredEntrySelection, error) {
	db := s.MySQL()

	var emptyID uuid.UUID

	if scoredEntrySelection.EntrySelectionID.String() == emptyID.String() {
		return models.ScoredEntrySelection{}, ValidationError{Reasons: []string{
			"EntrySelectionID is empty",
		}}
	}

	if scoredEntrySelection.StandingsID.String() == emptyID.String() {
		return models.ScoredEntrySelection{}, ValidationError{Reasons: []string{
			"StandingsID is empty",
		}}
	}

	// ensure that entrySelection exists
	entrySelectionRepo := repositories.NewEntrySelectionDatabaseRepository(db)
	if err := entrySelectionRepo.ExistsByID(ctx, scoredEntrySelection.EntrySelectionID.String()); err != nil {
		return models.ScoredEntrySelection{}, domainErrorFromDBError(err)
	}

	// ensure that standings exists
	standingsRepo := repositories.NewStandingsDatabaseRepository(db)
	if err := standingsRepo.ExistsByID(ctx, scoredEntrySelection.StandingsID.String()); err != nil {
		return models.ScoredEntrySelection{}, domainErrorFromDBError(err)
	}

	// override these values
	scoredEntrySelection.CreatedAt = time.Time{}
	scoredEntrySelection.UpdatedAt = sqltypes.NullTime{}

	scoredEntrySelectionRepo := repositories.NewScoredEntrySelectionDatabaseRepository(db)

	// write scoredEntrySelection to database
	if err := scoredEntrySelectionRepo.Insert(ctx, &scoredEntrySelection); err != nil {
		return models.ScoredEntrySelection{}, domainErrorFromDBError(err)
	}

	return scoredEntrySelection, nil
}

// RetrieveScoredEntrySelectionByIDs handles the retrieval of an existing ScoredEntrySelection in the database by its ID
func (s ScoredEntrySelectionAgent) RetrieveScoredEntrySelectionByIDs(ctx context.Context, entrySelectionID, standingsID string) (models.ScoredEntrySelection, error) {
	scoredEntrySelectionRepo := repositories.NewScoredEntrySelectionDatabaseRepository(s.MySQL())

	retrievedScoredEntrySelections, err := scoredEntrySelectionRepo.Select(ctx, map[string]interface{}{
		"entry_selection_id": entrySelectionID,
		"standings_id":       standingsID,
	}, false)
	if err != nil {
		return models.ScoredEntrySelection{}, domainErrorFromDBError(err)
	}

	return retrievedScoredEntrySelections[0], nil
}

// UpdateScoredEntrySelection handles the updating of an existing ScoredEntrySelection in the database
func (s ScoredEntrySelectionAgent) UpdateScoredEntrySelection(ctx context.Context, scoredEntrySelection models.ScoredEntrySelection) (models.ScoredEntrySelection, error) {
	scoredEntrySelectionRepo := repositories.NewScoredEntrySelectionDatabaseRepository(s.MySQL())

	// ensure the scoredEntrySelection exists
	if err := scoredEntrySelectionRepo.Exists(
		ctx,
		scoredEntrySelection.EntrySelectionID.String(),
		scoredEntrySelection.StandingsID.String(),
	); err != nil {
		return models.ScoredEntrySelection{}, domainErrorFromDBError(err)
	}

	// write to database
	if err := scoredEntrySelectionRepo.Update(ctx, &scoredEntrySelection); err != nil {
		return models.ScoredEntrySelection{}, domainErrorFromDBError(err)
	}

	return scoredEntrySelection, nil
}
