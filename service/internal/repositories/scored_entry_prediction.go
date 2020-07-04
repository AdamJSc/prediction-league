package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
	"time"
)

// stmtSelectEntryWithTotalScore represents a partial nested statement for grouping entry IDs along with their
// cumulative score at the point of the provided round number
const stmtSelectEntryWithTotalScore = `
	SELECT
		entry_id,
		SUM(score) AS total_score
	FROM (
		SELECT
			ep.entry_id,
			s.round_number,
			sep.score
		FROM (
			SELECT *
			FROM scored_entry_prediction
			ORDER BY created_at DESC
			LIMIT 100000000 -- arbitrary limit so that order by desc row order is retained for parent query
		) sep
		INNER JOIN standings s ON sep.standings_id = s.id
		INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
		INNER JOIN entry e ON ep.entry_id = e.id
		WHERE e.realm_name = ? s.season_id = ? AND s.round_number <= ?
		GROUP BY ep.entry_id, s.round_number
		ORDER BY ep.entry_id ASC, s.round_number DESC
	) AS sub
	GROUP BY entry_id
`

// stmtSelectEntryWithScoreThisRound represents a partial nested statement for retrieving entry IDs along with their
// score for the provided round number
const stmtSelectEntryWithScoreThisRound = `
	SELECT
		ep.entry_id,
		sep.score AS score_this_round
	FROM (
		SELECT *
		FROM scored_entry_prediction
		ORDER BY created_at DESC
		LIMIT 100000000 -- arbitrary limit so that order by desc row order is retained for parent query
	) sep
	INNER JOIN standings s ON sep.standings_id = s.id
	INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
	INNER JOIN entry e ON ep.entry_id = e.id
	WHERE e.realm_name = ? s.season_id = ? AND s.round_number = ?
	GROUP BY ep.entry_id
`

// stmtSelectEntryWithMinScore represents a partial nested statement for grouping entry IDs along with their
// minimum score at the point of the provided round number
const stmtSelectEntryWithMinScore = `
	SELECT
		entry_id,
		MIN(score) AS min_score
	FROM (
		SELECT
			ep.entry_id,
			s.round_number,
			sep.score
		FROM (
			SELECT *
			FROM scored_entry_prediction
			ORDER BY created_at DESC
			LIMIT 100000000 -- arbitrary limit so that order by desc row order is retained for parent query
		) sep
		INNER JOIN standings s ON sep.standings_id = s.id
		INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
		INNER JOIN entry e ON ep.entry_id = e.id
		WHERE e.realm_name = ? s.season_id = ? AND s.round_number <= ?
		GROUP BY ep.entry_id, s.round_number
		ORDER BY ep.entry_id ASC, s.round_number DESC
	) AS sub
	GROUP BY entry_id
`

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
	SelectEntryCumulativeScores(ctx context.Context, seasonID string, roundNumber int) (models.RankingWithScoreCollection, error)
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
		return MissingDBRecordError{fmt.Errorf("scored entry prediction with standings id %s and entry prediction id %s: not found", standingsID, entryPredictionID)}
	}

	return nil
}

// SelectEntryCumulativeScoresByRealm retrieves the current score, total score and minimum score for each entry ID
// based on the provided realm name, season ID and round number
func (s ScoredEntryPredictionDatabaseRepository) SelectEntryCumulativeScoresByRealm(ctx context.Context, realmName string, seasonID string, roundNumber int) ([]models.LeaderBoardRanking, error) {
	stmt := fmt.Sprintf(`
		SELECT
			entry_with_total_score.entry_id,
			entry_with_score_this_round.score_this_round,
			entry_with_min_score.min_score,
			entry_with_total_score.total_score
		FROM (%s) AS entry_with_total_score
		INNER JOIN (%s) AS entry_with_score_this_round
			ON entry_with_total_score.entry_id = entry_with_score_this_round.entry_id
		INNER JOIN (%s) AS entry_with_min_score
			ON entry_with_total_score.entry_id = entry_with_min_score.entry_id
		ORDER BY
			entry_with_total_score.total_score ASC,
			entry_with_min_score.min_score ASC,
			entry_with_score_this_round.score_this_round ASC
		`,
		stmtSelectEntryWithTotalScore,
		stmtSelectEntryWithScoreThisRound,
		stmtSelectEntryWithMinScore,
	)

	params := []interface{}{
		// params for stmtSelectEntryWithTotalScore partial
		realmName,
		seasonID,
		roundNumber,
		// params for stmtSelectEntryWithScoreThisRound partial
		realmName,
		seasonID,
		roundNumber,
		// params for stmtSelectEntryWithMinScore partial
		realmName,
		seasonID,
		roundNumber,
	}

	rows, err := s.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var lbRankings []models.LeaderBoardRanking

	count := 0
	for rows.Next() {
		count++
		var (
			entryID      string
			totalScore   int
			currentScore int
			minScore     int
		)
		if err := rows.Scan(
			&entryID,
			&totalScore,
			&currentScore,
			&minScore,
		); err != nil {
			return nil, wrapDBError(err)
		}

		lbRankings = append(lbRankings, models.LeaderBoardRanking{
			RankingWithScore: models.RankingWithScore{
				Ranking: models.Ranking{
					ID:       entryID,
					Position: count,
				},
				Score: currentScore,
			},
			MinScore:   minScore,
			TotalScore: totalScore,
		})
	}

	if len(lbRankings) == 0 {
		return nil, MissingDBRecordError{fmt.Errorf("no cumulative scores found for round %d in season %s", roundNumber, seasonID)}
	}

	return lbRankings, nil
}

// NewScoredEntryPredictionDatabaseRepository instantiates a new ScoredEntryPredictionDatabaseRepository with the provided DB agent
func NewScoredEntryPredictionDatabaseRepository(db coresql.Agent) ScoredEntryPredictionDatabaseRepository {
	return ScoredEntryPredictionDatabaseRepository{agent: db}
}
