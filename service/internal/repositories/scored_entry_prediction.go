package repositories

import (
	"encoding/json"
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
	"time"
)

// scoredEntryPredictionDBFields defines the fields used regularly in ScoredEntryPredictions-related transactions
var scoredEntryPredictionDBFields = []string{
	"rankings",
	"score",
}

// ScoredEntryPredictionRepository defines the interface for transacting with our ScoredEntryPredictions data source
type ScoredEntryPredictionRepository interface {
	Insert(ctx context.Context, scoredEntryPrediction *models.ScoredEntryPrediction) error
	Update(ctx context.Context, scoredEntryPrediction *models.ScoredEntryPrediction) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.ScoredEntryPrediction, error)
	Exists(ctx context.Context, entryPredictionID, standingsID string) error
}

// ScoredEntryPredictionDatabaseRepository defines our DB-backed ScoredEntryPredictions data store
type ScoredEntryPredictionDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new ScoredEntryPrediction into the database
func (s ScoredEntryPredictionDatabaseRepository) Insert(ctx context.Context, scoredEntryPrediction *models.ScoredEntryPrediction) error {
	stmt := `INSERT INTO scored_entry_prediction (entry_prediction_id, standings_id,` + getDBFieldsStringFromFields(scoredEntryPredictionDBFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	rawRankings, err := json.Marshal(&scoredEntryPrediction.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.agent.QueryContext(
		ctx,
		stmt,
		scoredEntryPrediction.EntryPredictionID,
		scoredEntryPrediction.StandingsID,
		rawRankings,
		scoredEntryPrediction.Score,
		now,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	scoredEntryPrediction.CreatedAt = now

	return nil
}

// Update updates an existing ScoredEntryPrediction in the database
func (s ScoredEntryPredictionDatabaseRepository) Update(ctx context.Context, scoredEntryPrediction *models.ScoredEntryPrediction) error {
	stmt := `UPDATE scored_entry_prediction
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(scoredEntryPredictionDBFields) + `, updated_at = ?
				WHERE entry_prediction_id = ? AND standings_id = ?`

	now := sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	rawRankings, err := json.Marshal(&scoredEntryPrediction.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.agent.QueryContext(
		ctx,
		stmt,
		rawRankings,
		scoredEntryPrediction.Score,
		now,
		scoredEntryPrediction.EntryPredictionID,
		scoredEntryPrediction.StandingsID,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	scoredEntryPrediction.UpdatedAt = now

	return nil
}

// Select retrieves ScoredEntryPredictions from our database based on the provided criteria
func (s ScoredEntryPredictionDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.ScoredEntryPrediction, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT entry_prediction_id, standings_id, ` + getDBFieldsStringFromFields(scoredEntryPredictionDBFields) + `, created_at, updated_at
				FROM scored_entry_prediction ` + whereStmt

	rows, err := s.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var scoredEntryPredictions []models.ScoredEntryPrediction
	var rawRankings []byte

	for rows.Next() {
		scoredEntryPrediction := models.ScoredEntryPrediction{}

		if err := rows.Scan(
			&scoredEntryPrediction.EntryPredictionID,
			&scoredEntryPrediction.StandingsID,
			&rawRankings,
			&scoredEntryPrediction.Score,
			&scoredEntryPrediction.CreatedAt,
			&scoredEntryPrediction.UpdatedAt,
		); err != nil {
			return nil, wrapDBError(err)
		}

		if err := json.Unmarshal(rawRankings, &scoredEntryPrediction.Rankings); err != nil {
			return nil, err
		}

		scoredEntryPredictions = append(scoredEntryPredictions, scoredEntryPrediction)
	}

	if len(scoredEntryPredictions) == 0 {
		return nil, MissingDBRecordError{errors.New("no scored entry predictions found")}
	}

	return scoredEntryPredictions, nil
}

// Exists determines whether a ScoredEntryPrediction with the provided ID exists in the database
func (s ScoredEntryPredictionDatabaseRepository) Exists(ctx context.Context, entryPredictionID, standingsID string) error {
	stmt := `SELECT COUNT(*) FROM scored_entry_prediction WHERE entry_prediction_id = ? AND standings_id = ?`

	row := s.agent.QueryRowContext(ctx, stmt, entryPredictionID, standingsID)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{errors.New("scored entry prediction not found")}
	}

	return nil
}

// NewScoredEntryPredictionDatabaseRepository instantiates a new ScoredEntryPredictionDatabaseRepository with the provided DB agent
func NewScoredEntryPredictionDatabaseRepository(db coresql.Agent) ScoredEntryPredictionDatabaseRepository {
	return ScoredEntryPredictionDatabaseRepository{agent: db}
}
