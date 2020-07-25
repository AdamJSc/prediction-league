package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
	"time"
)

// standingsDBFields defines the fields used regularly in Standings-related transactions
var standingsDBFields = []string{
	"season_id",
	"round_number",
	"rankings",
	"finalised",
}

// StandingsRepository defines the interface for transacting with our Standings data source
type StandingsRepository interface {
	Insert(ctx context.Context, standings *models.Standings) error
	Update(ctx context.Context, standings *models.Standings) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.Standings, error)
	ExistsByID(ctx context.Context, id string) error
	SelectLatestBySeasonIDAndTimestamp(ctx context.Context, seasonID string, ts time.Time) (models.Standings, error)
}

// StandingsDatabaseRepository defines our DB-backed Standings data store
type StandingsDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new Standings into the database
func (s StandingsDatabaseRepository) Insert(ctx context.Context, standings *models.Standings) error {
	stmt := `INSERT INTO standings (id, ` + getDBFieldsStringFromFields(standingsDBFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?, ?)`

	rankings, err := json.Marshal(&standings.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.agent.QueryContext(
		ctx,
		stmt,
		standings.ID,
		standings.SeasonID,
		standings.RoundNumber,
		rankings,
		standings.Finalised,
		standings.CreatedAt,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return nil
}

// Update updates an existing Standings in the database
func (s StandingsDatabaseRepository) Update(ctx context.Context, standings *models.Standings) error {
	stmt := `UPDATE standings
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(standingsDBFields) + `, updated_at = ?
				WHERE id = ?`

	rankings, err := json.Marshal(&standings.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.agent.QueryContext(
		ctx,
		stmt,
		standings.SeasonID,
		standings.RoundNumber,
		rankings,
		standings.Finalised,
		standings.UpdatedAt,
		standings.ID,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return nil
}

// Select retrieves Standings from our database based on the provided criteria
func (s StandingsDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.Standings, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(standingsDBFields) + `, created_at, updated_at FROM standings ` + whereStmt

	rows, err := s.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var retrievedStandings []models.Standings
	var rankings []byte

	for rows.Next() {
		standings := models.Standings{}

		if err := rows.Scan(
			&standings.ID,
			&standings.SeasonID,
			&standings.RoundNumber,
			&rankings,
			&standings.Finalised,
			&standings.CreatedAt,
			&standings.UpdatedAt,
		); err != nil {
			return nil, wrapDBError(err)
		}

		if err := json.Unmarshal(rankings, &standings.Rankings); err != nil {
			return nil, wrapDBError(err)
		}

		retrievedStandings = append(retrievedStandings, standings)
	}

	if len(retrievedStandings) == 0 {
		return nil, MissingDBRecordError{errors.New("no standings found")}
	}

	return retrievedStandings, nil
}

// ExistsByID determines whether a Standings with the provided ID exists in the database
func (s StandingsDatabaseRepository) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM standings WHERE id = ?`

	row := s.agent.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{fmt.Errorf("standings id %s: not found", id)}
	}

	return nil
}

// Select retrieves Standings from our database based on the provided criteria
func (s StandingsDatabaseRepository) SelectLatestBySeasonIDAndTimestamp(ctx context.Context, seasonID string, ts time.Time) (models.Standings, error) {
	stmt := `SELECT id, ` + getDBFieldsStringFromFields(standingsDBFields) + `, created_at, updated_at
			FROM standings
			WHERE season_id = ?
				AND created_at <= ?
			ORDER BY created_at DESC
			LIMIT 1`

	row := s.agent.QueryRowContext(ctx, stmt, seasonID, ts)

	var retrievedStandings models.Standings
	var rankings []byte

	if err := row.Scan(
		&retrievedStandings.ID,
		&retrievedStandings.SeasonID,
		&retrievedStandings.RoundNumber,
		&rankings,
		&retrievedStandings.Finalised,
		&retrievedStandings.CreatedAt,
		&retrievedStandings.UpdatedAt,
	); err != nil {
		return models.Standings{}, wrapDBError(err)
	}

	if err := json.Unmarshal(rankings, &retrievedStandings.Rankings); err != nil {
		return models.Standings{}, wrapDBError(err)
	}

	return retrievedStandings, nil
}

// NewStandingsDatabaseRepository instantiates a new StandingsDatabaseRepository with the provided DB agent
func NewStandingsDatabaseRepository(db coresql.Agent) StandingsDatabaseRepository {
	return StandingsDatabaseRepository{agent: db}
}