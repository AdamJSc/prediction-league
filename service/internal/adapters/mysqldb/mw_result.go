package mysqldb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"prediction-league/service/internal/domain"
	"strings"
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
	stmt := `
	SELECT
	    mw_submission_id,
	    team_rankings,
	    score,
		created_at,
	    updated_at
	FROM
		mw_result
	WHERE
		mw_submission_id = ?
	`

	row := m.db.QueryRowContext(ctx, stmt, submissionID)
	result := &domain.MatchWeekResult{}
	var teamRankingsRaw []byte

	if err := row.Scan(
		&result.MatchWeekSubmissionID,
		&teamRankingsRaw,
		&result.Score,
		&result.CreatedAt,
		&result.UpdatedAt,
	); err != nil {
		return nil, wrapDBError(err)
	}

	if err := json.Unmarshal(teamRankingsRaw, &result.TeamRankings); err != nil {
		return nil, fmt.Errorf("cannot unmarshal raw team rankings: %w", err)
	}

	resultModifiers, err := m.getResultModifiers(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("cannot get result modifiers: %w", err)
	}

	result.Modifiers = resultModifiers

	return result, nil
}

// getResultModifiers from database
func (m *MatchWeekResultRepo) getResultModifiers(ctx context.Context, resultID uuid.UUID) ([]domain.ModifierSummary, error) {
	stmt := `
	SELECT
		code,
		value
	FROM
		mw_result_modifier
	WHERE
		mw_result_id = ?
	ORDER BY sort_order ASC
	`

	rows, err := m.db.QueryContext(ctx, stmt, resultID)
	if err != nil {
		return nil, wrapDBError(err)
	}

	summaries := make([]domain.ModifierSummary, 0)
	for rows.Next() {
		summary := domain.ModifierSummary{}
		if err := rows.Scan(
			&summary.Code,
			&summary.Value,
		); err != nil {
			return nil, fmt.Errorf("cannot scan row: %w", wrapDBError(err))
		}

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot read result modifier rows: %w", wrapDBError(err))
	}

	return summaries, nil
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
		sort_order,
		code,
		value
	) VALUES
	`

	var args []interface{}
	var placeholders []string
	for idx, modifier := range modifiers {
		placeholders = append(placeholders, ` (?,?,?,?)`)
		sortOrder := idx + 1 // sort order value matches position within slice sequence
		args = append(args, resultID, sortOrder, modifier.Code, modifier.Value)
	}

	insertStmt += strings.Join(placeholders, ",")

	if _, err := m.db.ExecContext(
		ctx,
		insertStmt,
		args...,
	); err != nil {
		return wrapDBError(err)
	}

	return nil
}

// Update the provided MatchWeekResult by its submission id
func (m *MatchWeekResultRepo) Update(ctx context.Context, mwResult *domain.MatchWeekResult) error {
	if mwResult == nil {
		return nil
	}

	teamRankingsRaw, err := json.Marshal(mwResult.TeamRankings)
	if err != nil {
		return fmt.Errorf("cannot marshal team rankings: %w", err)
	}

	updatedAt := m.timeFn()
	mwResult.UpdatedAt = &updatedAt

	stmt := `
	UPDATE mw_result
	SET
	    team_rankings = ?,
	    score = ?,
		updated_at = ?
	WHERE mw_submission_id = ?
	`

	result, err := m.db.ExecContext(
		ctx,
		stmt,
		teamRankingsRaw,
		mwResult.Score,
		mwResult.UpdatedAt,
		mwResult.MatchWeekSubmissionID,
	)
	if err != nil {
		return wrapDBError(err)
	}

	rowCount, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowCount == 0 {
		return domain.MissingDBRecordError{Err: fmt.Errorf("match week result not found: submission id %s", mwResult.MatchWeekSubmissionID)}
	}

	if err := m.insertResultModifiers(ctx, mwResult.MatchWeekSubmissionID, mwResult.Modifiers); err != nil {
		return fmt.Errorf("cannot insert result modifiers: %w", err)
	}

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
