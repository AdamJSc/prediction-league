package repositories

import (
	"encoding/json"
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
	"time"
)

// entryPredictionDBFields defines the fields used regularly in EntryPredictions-related transactions
var entryPredictionDBFields = []string{
	"entry_id",
	"rankings",
}

// EntryPredictionRepository defines the interface for transacting with our EntryPredictions data source
type EntryPredictionRepository interface {
	Insert(ctx context.Context, entryPrediction *models.EntryPrediction) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.EntryPrediction, error)
	ExistsByID(ctx context.Context, id string) error
}

// EntryPredictionDatabaseRepository defines our DB-backed EntryPredictions data store
type EntryPredictionDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new EntryPrediction into the database
func (e EntryPredictionDatabaseRepository) Insert(ctx context.Context, entryPrediction *models.EntryPrediction) error {
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

	rows, err := e.agent.QueryContext(
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
func (e EntryPredictionDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.EntryPrediction, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entryPredictionDBFields) + `, created_at FROM entry_prediction ` + whereStmt

	rows, err := e.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var entryPredictions []models.EntryPrediction
	var rawRankings []byte

	for rows.Next() {
		entryPrediction := models.EntryPrediction{}

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
		return nil, MissingDBRecordError{errors.New("no entry predictions found")}
	}

	return entryPredictions, nil
}

// ExistsByID determines whether an EntryPrediction with the provided ID exists in the database
func (e EntryPredictionDatabaseRepository) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM entry_prediction WHERE id = ?`

	row := e.agent.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{errors.New("entry prediction not found")}
	}

	return nil
}

// SelectByIDAndTimestamp retrieves the most recent EntryPrediction that exists for the provided entry ID and timestamp
func (e EntryPredictionDatabaseRepository) SelectByEntryIDAndTimestamp(ctx context.Context, entryID string, ts time.Time) (models.EntryPrediction, error) {
	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entryPredictionDBFields) + `, created_at FROM entry_prediction
			WHERE entry_id = ?
			AND created_at <= ?
			ORDER BY created_at DESC
			LIMIT 1`

	row := e.agent.QueryRowContext(ctx, stmt, entryID, ts)

	var entryPrediction models.EntryPrediction
	var rawRankings []byte

	if err := row.Scan(
		&entryPrediction.ID,
		&entryPrediction.EntryID,
		&rawRankings,
		&entryPrediction.CreatedAt,
	); err != nil {
		return models.EntryPrediction{}, wrapDBError(err)
	}

	if err := json.Unmarshal(rawRankings, &entryPrediction.Rankings); err != nil {
		return models.EntryPrediction{}, err
	}

	return entryPrediction, nil
}

// NewEntryPredictionDatabaseRepository instantiates a new EntryPredictionDatabaseRepository with the provided DB agent
func NewEntryPredictionDatabaseRepository(db coresql.Agent) EntryPredictionDatabaseRepository {
	return EntryPredictionDatabaseRepository{agent: db}
}
