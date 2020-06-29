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

// scoredEntrySelectionDBFields defines the fields used regularly in ScoredEntrySelections-related transactions
var scoredEntrySelectionDBFields = []string{
	"rankings",
	"score",
}

// ScoredEntrySelectionRepository defines the interface for transacting with our ScoredEntrySelections data source
type ScoredEntrySelectionRepository interface {
	Insert(ctx context.Context, scoredEntrySelection *models.ScoredEntrySelection) error
	Update(ctx context.Context, scoredEntrySelection *models.ScoredEntrySelection) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.ScoredEntrySelection, error)
	Exists(ctx context.Context, entrySelectionID, standingsID string) error
}

// ScoredEntrySelectionDatabaseRepository defines our DB-backed ScoredEntrySelections data store
type ScoredEntrySelectionDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new ScoredEntrySelection into the database
func (s ScoredEntrySelectionDatabaseRepository) Insert(ctx context.Context, scoredEntrySelection *models.ScoredEntrySelection) error {
	stmt := `INSERT INTO scored_entry_selection (entry_selection_id, standings_id,` + getDBFieldsStringFromFields(scoredEntrySelectionDBFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	rawRankings, err := json.Marshal(&scoredEntrySelection.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.agent.QueryContext(
		ctx,
		stmt,
		scoredEntrySelection.EntrySelectionID,
		scoredEntrySelection.StandingsID,
		rawRankings,
		scoredEntrySelection.Score,
		now,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	scoredEntrySelection.CreatedAt = now

	return nil
}

// Update updates an existing ScoredEntrySelection in the database
func (s ScoredEntrySelectionDatabaseRepository) Update(ctx context.Context, scoredEntrySelection *models.ScoredEntrySelection) error {
	stmt := `UPDATE scored_entry_selection
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(scoredEntrySelectionDBFields) + `, updated_at = ?
				WHERE entry_selection_id = ? AND standings_id = ?`

	now := sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	rawRankings, err := json.Marshal(&scoredEntrySelection.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.agent.QueryContext(
		ctx,
		stmt,
		rawRankings,
		scoredEntrySelection.Score,
		now,
		scoredEntrySelection.EntrySelectionID,
		scoredEntrySelection.StandingsID,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	scoredEntrySelection.UpdatedAt = now

	return nil
}

// Select retrieves ScoredEntrySelections from our database based on the provided criteria
func (s ScoredEntrySelectionDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.ScoredEntrySelection, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT entry_selection_id, standings_id, ` + getDBFieldsStringFromFields(scoredEntrySelectionDBFields) + `, created_at, updated_at
				FROM scored_entry_selection ` + whereStmt

	rows, err := s.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return []models.ScoredEntrySelection{}, wrapDBError(err)
	}
	defer rows.Close()

	var scoredEntrySelections []models.ScoredEntrySelection
	var rawRankings []byte

	for rows.Next() {
		scoredEntrySelection := models.ScoredEntrySelection{}

		if err := rows.Scan(
			&scoredEntrySelection.EntrySelectionID,
			&scoredEntrySelection.StandingsID,
			&rawRankings,
			&scoredEntrySelection.Score,
			&scoredEntrySelection.CreatedAt,
			&scoredEntrySelection.UpdatedAt,
		); err != nil {
			return []models.ScoredEntrySelection{}, wrapDBError(err)
		}

		if err := json.Unmarshal(rawRankings, &scoredEntrySelection.Rankings); err != nil {
			return []models.ScoredEntrySelection{}, err
		}

		scoredEntrySelections = append(scoredEntrySelections, scoredEntrySelection)
	}

	if len(scoredEntrySelections) == 0 {
		return []models.ScoredEntrySelection{}, MissingDBRecordError{errors.New("no entries found")}
	}

	return scoredEntrySelections, nil
}

// Exists determines whether a ScoredEntrySelection with the provided ID exists in the database
func (s ScoredEntrySelectionDatabaseRepository) Exists(ctx context.Context, entrySelectionID, standingsID string) error {
	stmt := `SELECT COUNT(*) FROM scored_entry_selection WHERE entry_selection_id = ? AND standings_id = ?`

	row := s.agent.QueryRowContext(ctx, stmt, entrySelectionID, standingsID)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{errors.New("scored entry selection not found")}
	}

	return nil
}

// NewScoredEntrySelectionDatabaseRepository instantiates a new ScoredEntrySelectionDatabaseRepository with the provided DB agent
func NewScoredEntrySelectionDatabaseRepository(db coresql.Agent) ScoredEntrySelectionDatabaseRepository {
	return ScoredEntrySelectionDatabaseRepository{agent: db}
}
