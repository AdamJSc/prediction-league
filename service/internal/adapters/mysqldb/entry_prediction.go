package mysqldb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"prediction-league/service/internal/domain"
	"time"
)

// entryPredictionDBFields defines the fields used regularly in EntryPredictions-related transactions
var entryPredictionDBFields = []string{
	"entry_id",
	"rankings",
}

// EntryPredictionRepo defines our DB-backed EntryPredictions data store
type EntryPredictionRepo struct {
	db *sql.DB
}

// Insert inserts a new EntryPrediction into the database
func (e *EntryPredictionRepo) Insert(ctx context.Context, entryPrediction *domain.EntryPrediction) error {
	stmt := `INSERT INTO entry_prediction (id, ` + getDBFieldsStringFromFields(entryPredictionDBFields) + `, created_at)
					VALUES (?, ?, ?, ?)`

	var emptyTime time.Time
	if entryPrediction.CreatedAt.Equal(emptyTime) {
		entryPrediction.CreatedAt = time.Now()
	}

	entryPrediction.CreatedAt = entryPrediction.CreatedAt.Truncate(time.Second)

	rawRankings, err := json.Marshal(&entryPrediction.Rankings)
	if err != nil {
		return err
	}

	rows, err := e.db.QueryContext(
		ctx,
		stmt,
		entryPrediction.ID,
		entryPrediction.EntryID,
		rawRankings,
		entryPrediction.CreatedAt,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return nil
}

// Select retrieves EntryPredictions from our database based on the provided criteria
func (e *EntryPredictionRepo) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]domain.EntryPrediction, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entryPredictionDBFields) + `, created_at FROM entry_prediction ` + whereStmt

	rows, err := e.db.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var entryPredictions []domain.EntryPrediction
	var rawRankings []byte

	for rows.Next() {
		entryPrediction := domain.EntryPrediction{}

		if err := rows.Scan(
			&entryPrediction.ID,
			&entryPrediction.EntryID,
			&rawRankings,
			&entryPrediction.CreatedAt,
		); err != nil {
			return nil, wrapDBError(err)
		}

		if err := json.Unmarshal(rawRankings, &entryPrediction.Rankings); err != nil {
			return nil, err
		}

		entryPredictions = append(entryPredictions, entryPrediction)
	}

	if len(entryPredictions) == 0 {
		return nil, domain.MissingDBRecordError{Err: errors.New("no entry predictions found")}
	}

	return entryPredictions, nil
}

// ExistsByID determines whether an EntryPrediction with the provided ID exists in the database
func (e *EntryPredictionRepo) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM entry_prediction WHERE id = ?`

	row := e.db.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return domain.MissingDBRecordError{Err: fmt.Errorf("entry prediction id %s: not found", id)}
	}

	return nil
}

// SelectByIDAndTimestamp retrieves the most recent EntryPrediction that exists for the provided entry ID and timestamp
func (e *EntryPredictionRepo) SelectByEntryIDAndTimestamp(ctx context.Context, entryID string, ts time.Time) (domain.EntryPrediction, error) {
	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entryPredictionDBFields) + `, created_at FROM entry_prediction
			WHERE entry_id = ?
			AND created_at <= ?
			ORDER BY created_at DESC
			LIMIT 1`

	row := e.db.QueryRowContext(ctx, stmt, entryID, ts)

	var entryPrediction domain.EntryPrediction
	var rawRankings []byte

	if err := row.Scan(
		&entryPrediction.ID,
		&entryPrediction.EntryID,
		&rawRankings,
		&entryPrediction.CreatedAt,
	); err != nil {
		return domain.EntryPrediction{}, wrapDBError(err)
	}

	if err := json.Unmarshal(rawRankings, &entryPrediction.Rankings); err != nil {
		return domain.EntryPrediction{}, err
	}

	return entryPrediction, nil
}

// NewEntryPredictionRepo instantiates a new EntryPredictionRepo with the provided DB agent
func NewEntryPredictionRepo(db *sql.DB) *EntryPredictionRepo {
	return &EntryPredictionRepo{db: db}
}
