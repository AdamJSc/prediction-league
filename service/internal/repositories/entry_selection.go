package repositories

import (
	"encoding/json"
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
	"time"
)

// entrySelectionDBFields defines the fields used regularly in EntrySelections-related transactions
var entrySelectionDBFields = []string{
	"entry_id",
	"rankings",
}

// EntrySelectionRepository defines the interface for transacting with our EntrySelections data source
type EntrySelectionRepository interface {
	Insert(ctx context.Context, entrySelection *models.EntrySelection) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.EntrySelection, error)
	ExistsByID(ctx context.Context, id string) error
}

// EntrySelectionDatabaseRepository defines our DB-backed EntrySelections data store
type EntrySelectionDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new EntrySelection into the database
func (e EntrySelectionDatabaseRepository) Insert(ctx context.Context, entrySelection *models.EntrySelection) error {
	stmt := `INSERT INTO entry_selection (id, ` + getDBFieldsStringFromFields(entrySelectionDBFields) + `, created_at)
					VALUES (?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	rawRankings, err := json.Marshal(&entrySelection.Rankings)
	if err != nil {
		return err
	}

	if _, err := e.agent.QueryContext(
		ctx,
		stmt,
		entrySelection.ID,
		entrySelection.EntryID,
		rawRankings,
		now,
	); err != nil {
		return wrapDBError(err)
	}

	entrySelection.CreatedAt = now

	return nil
}

// Select retrieves EntrySelections from our database based on the provided criteria
func (e EntrySelectionDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.EntrySelection, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entrySelectionDBFields) + `, created_at FROM entry_selection ` + whereStmt

	rows, err := e.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return []models.EntrySelection{}, wrapDBError(err)
	}

	var entrySelections []models.EntrySelection
	var rawRankings []byte

	for rows.Next() {
		entrySelection := models.EntrySelection{}

		if err := rows.Scan(
			&entrySelection.ID,
			&entrySelection.EntryID,
			&rawRankings,
			&entrySelection.CreatedAt,
		); err != nil {
			return []models.EntrySelection{}, wrapDBError(err)
		}

		if err := json.Unmarshal(rawRankings, &entrySelection.Rankings); err != nil {
			return []models.EntrySelection{}, err
		}

		entrySelections = append(entrySelections, entrySelection)
	}

	if len(entrySelections) == 0 {
		return []models.EntrySelection{}, MissingDBRecordError{errors.New("no entry selections found")}
	}

	return entrySelections, nil
}

// ExistsByID determines whether an EntrySelection with the provided ID exists in the database
func (e EntrySelectionDatabaseRepository) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM entry_selection WHERE id = ?`

	row := e.agent.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{errors.New("entry selection not found")}
	}

	return nil
}

// NewEntrySelectionDatabaseRepository instantiates a new EntrySelectionDatabaseRepository with the provided DB agent
func NewEntrySelectionDatabaseRepository(db coresql.Agent) EntrySelectionDatabaseRepository {
	return EntrySelectionDatabaseRepository{agent: db}
}
