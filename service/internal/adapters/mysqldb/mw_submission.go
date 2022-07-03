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

type (
	idFunc   func() (uuid.UUID, error) // custom function to return a uuid
	timeFunc func() time.Time          // custom function to return a timestamp
)

// MatchWeekSubmissionRepo implements operations against a sql db
type MatchWeekSubmissionRepo struct {
	db     *sql.DB
	idFn   idFunc
	timeFn timeFunc
}

// GetByLegacyIDAndMatchWeekNumber returns the MatchWeekSubmission that matches the provided entry prediction id and match week number
func (m *MatchWeekSubmissionRepo) GetByLegacyIDAndMatchWeekNumber(ctx context.Context, entryPredictionID uuid.UUID, mwNumber uint16) (*domain.MatchWeekSubmission, error) {
	stmt := `
	SELECT
    	mws.id,
    	mws.entry_id,
	    mws.mw_number,
	    mws.team_rankings,
	    mws.legacy_entry_prediction_id,
		mws.created_at,
		mws.updated_at
	FROM
		mw_submission AS mws
	WHERE
		mws.legacy_entry_prediction_id = ?
	AND
	    mws.mw_number = ?
	ORDER BY mws.created_at DESC
	LIMIT 1
	`

	row := m.db.QueryRowContext(ctx, stmt, entryPredictionID, mwNumber)
	submission := &domain.MatchWeekSubmission{}
	var teamRankingsRaw []byte

	if err := row.Scan(
		&submission.ID,
		&submission.EntryID,
		&submission.MatchWeekNumber,
		&teamRankingsRaw,
		&submission.LegacyEntryPredictionID,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	); err != nil {
		return nil, wrapDBError(err)
	}

	if err := json.Unmarshal(teamRankingsRaw, &submission.TeamRankings); err != nil {
		return nil, fmt.Errorf("cannot unmarshal raw team rankings: %w", err)
	}

	return submission, nil
}

// Insert the provided MatchWeekSubmission into the database
func (m *MatchWeekSubmissionRepo) Insert(ctx context.Context, submission *domain.MatchWeekSubmission) error {
	if submission == nil {
		return nil
	}

	teamRankingsRaw, err := json.Marshal(submission.TeamRankings)
	if err != nil {
		return fmt.Errorf("cannot marshal team rankings: %w", err)
	}

	newID, err := m.idFn()
	if err != nil {
		return fmt.Errorf("cannot get uuid: %w", err)
	}

	createdAt := m.timeFn()

	submission.ID = newID
	submission.CreatedAt = createdAt

	stmt := `
	INSERT INTO mw_submission (
    	id,
    	entry_id,
	    mw_number,
	    team_rankings,
	    legacy_entry_prediction_id,
		created_at
	) VALUES (?,?,?,?,?,?)
	`

	if _, err := m.db.ExecContext(
		ctx,
		stmt,
		submission.ID,
		submission.EntryID,
		submission.MatchWeekNumber,
		teamRankingsRaw,
		submission.LegacyEntryPredictionID,
		submission.CreatedAt,
	); err != nil {
		return fmt.Errorf("cannot insert submission: %w", err)
	}

	return nil
}

// Update the provided MatchWeekSubmission by its id
func (m *MatchWeekSubmissionRepo) Update(ctx context.Context, submission *domain.MatchWeekSubmission) error {
	if submission == nil {
		return nil
	}

	teamRankingsRaw, err := json.Marshal(submission.TeamRankings)
	if err != nil {
		return fmt.Errorf("cannot marshal team rankings: %w", err)
	}

	updatedAt := m.timeFn()
	submission.UpdatedAt = &updatedAt

	stmt := `
	UPDATE mw_submission
	SET
    	entry_id = ?,
	    mw_number = ?,
	    team_rankings = ?,
	    legacy_entry_prediction_id = ?,
		updated_at = ?
	WHERE id = ?
	`

	if _, err := m.db.ExecContext(
		ctx,
		stmt,
		submission.EntryID,
		submission.MatchWeekNumber,
		teamRankingsRaw,
		submission.LegacyEntryPredictionID,
		submission.UpdatedAt,
		submission.ID,
	); err != nil {
		return fmt.Errorf("cannot update submission: %w", err)
	}

	return nil
}

// NewMatchWeekSubmissionRepo instantiates a new MatchWeekSubmissionRepo with the provided attributes
func NewMatchWeekSubmissionRepo(db *sql.DB, idFn idFunc, timeFn timeFunc) (*MatchWeekSubmissionRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db: %w", domain.ErrIsNil)
	}
	if idFn == nil {
		idFn = uuid.NewUUID
	}
	if timeFn == nil {
		timeFn = time.Now
	}
	return &MatchWeekSubmissionRepo{
		db:     db,
		idFn:   idFn,
		timeFn: timeFn,
	}, nil
}
