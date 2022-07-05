package mysqldb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"prediction-league/service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// MatchWeekResultRepo implements operations against a sql db
type MatchWeekResultRepo struct {
	db     *sql.DB
	timeFn timeFunc
}

// GetBySubmissionID returns the MatchWeekSubmission that matches the provided id
func (m *MatchWeekResultRepo) GetBySubmissionID(ctx context.Context, submissionID uuid.UUID) (*domain.MatchWeekResult, error) {
	// TODO: feat - implement repo method
	return nil, nil
}

// Insert the provided MatchWeekResult into the database
func (m *MatchWeekResultRepo) Insert(ctx context.Context, mwResult *domain.MatchWeekResult) error {
	if mwResult == nil {
		return nil
	}

	teamRankingsRaw, err := json.Marshal(mwResult.TeamRankings)
	if err != nil {
		return fmt.Errorf("cannot marshal team rankings: %w", err)
	}

	createdAt := m.timeFn()
	mwResult.CreatedAt = createdAt

	stmt := `
	INSERT INTO mw_result (
    	id,
	    mw_submission_id,
	    team_rankings,
	    score,
		created_at
	) VALUES (?,?,?,?,?)
	`

	if _, err := m.db.ExecContext(
		ctx,
		stmt,
		mwResult.MatchWeekSubmissionID, // repeat value for id
		mwResult.MatchWeekSubmissionID,
		teamRankingsRaw,
		mwResult.Score,
		mwResult.CreatedAt,
	); err != nil {
		return wrapDBError(err)
	}

	if err := m.insertResultModifiers(ctx, mwResult.MatchWeekSubmissionID, mwResult.Modifiers); err != nil {
		return fmt.Errorf("cannot insert result modifiers: %w", err)
	}

	return nil
}

// insertResultModifiers into database
func (m *MatchWeekResultRepo) insertResultModifiers(ctx context.Context, resultID uuid.UUID, modifiers []domain.ModifierSummary) error {
	// prune existing modifiers for provided result id
	truncateStmt := `
	DELETE FROM mw_result_modifier
	WHERE
		mw_result_id = ?
	`

	if _, err := m.db.ExecContext(ctx, truncateStmt, resultID); err != nil {
		return fmt.Errorf("cannot prune existing result modifiers: %w", wrapDBError(err))
	}

	if len(modifiers) == 0 {
		// no modifiers to insert
		return nil
	}

	// bulk-insert provided modifiers
	insertStmt := `
	INSERT INTO mw_result_modifier (
		mw_result_id,
		code,
		value
	) VALUES
	`

	var args []interface{}
	for _, modifier := range modifiers {
		insertStmt += ` (?,?,?)`
		args = append(args, resultID, modifier.Code, modifier.Value)
	}

	if _, err := m.db.ExecContext(
		ctx,
		insertStmt,
		args,
	); err != nil {
		return wrapDBError(err)
	}

	return nil
}

// Update the provided MatchWeekResult by its submission id
func (m *MatchWeekResultRepo) Update(ctx context.Context, mwResult *domain.MatchWeekResult) error {
	// TODO: feat - implement repo method
	return nil
}

// NewMatchWeekResultRepo instantiates a new MatchWeekSubmissionRepo with the provided attributes
func NewMatchWeekResultRepo(db *sql.DB, timeFn timeFunc) (*MatchWeekResultRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db: %w", domain.ErrIsNil)
	}

	if timeFn == nil {
		timeFn = time.Now
	}

	return &MatchWeekResultRepo{
		db:     db,
		timeFn: timeFn,
	}, nil
}
