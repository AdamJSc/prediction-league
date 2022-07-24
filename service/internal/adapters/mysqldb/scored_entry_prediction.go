package mysqldb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"

	"golang.org/x/net/context"
)

// stmtSelectEntryWithTotalScore represents a partial nested statement for grouping entry IDs along with their
// cumulative score at the point of the provided round number
const stmtSelectEntryWithTotalScore = `
	SELECT
		t2.entry_id,
		SUM(t2.score) AS total_score
	FROM (
		SELECT
			t1.entry_id,
			t1.score,
			t1.round_number
		FROM (
			SELECT
				e.id AS entry_id,
				sep.*,
				s.round_number
			FROM scored_entry_prediction sep
			INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
			INNER JOIN entry e ON ep.entry_id = e.id
			INNER JOIN standings s ON sep.standings_id = s.id
			WHERE
				e.realm_name = ?
				AND e.season_id = ?
				AND s.round_number <= ?
				AND e.approved_at IS NOT NULL
			ORDER BY sep.created_at DESC
			LIMIT 100000000 -- arbitrary limit so that order by desc row order is retained for parent query
		) t1
		GROUP BY t1.entry_id, t1.round_number
	) t2
	GROUP BY t2.entry_id
`

// stmtSelectEntryWithScoreThisRound represents a partial nested statement for retrieving entry IDs along with their
// score for the provided round number
const stmtSelectEntryWithScoreThisRound = `
	SELECT
		t1.entry_id,
		t1.score AS score_this_round
	FROM (
		SELECT
			e.id AS entry_id,
			sep.*
		FROM scored_entry_prediction sep
		INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
		INNER JOIN entry e ON ep.entry_id = e.id
		INNER JOIN standings s ON sep.standings_id = s.id
		WHERE
			e.realm_name = ?
			AND e.season_id = ?
			AND s.round_number = ?
			AND e.approved_at IS NOT NULL
		ORDER BY sep.created_at DESC
		LIMIT 100000000 -- arbitrary limit so that order by desc row order is retained for parent query
	) t1
	GROUP BY t1.entry_id
`

// stmtSelectEntryWithMaxScore represents a partial nested statement for grouping entry IDs along with their
// maximum score at the point of the provided round number
const stmtSelectEntryWithMaxScore = `
	SELECT
		t2.entry_id,
		MAX(t2.score) AS max_score
	FROM (
		SELECT
			t1.entry_id,
			t1.score,
			t1.round_number
		FROM (
			SELECT
				e.id AS entry_id,
				sep.*,
				s.round_number
			FROM scored_entry_prediction sep
			INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
			INNER JOIN entry e ON ep.entry_id = e.id
			INNER JOIN standings s ON sep.standings_id = s.id
			WHERE
				e.realm_name = ?
				AND e.season_id = ?
				AND s.round_number <= ?
				AND e.approved_at IS NOT NULL
			ORDER BY sep.created_at DESC
			LIMIT 100000000 -- arbitrary limit so that order by desc row order is retained for parent query
		) t1
		GROUP BY t1.entry_id, t1.round_number
	) t2
	GROUP BY t2.entry_id
`

// scoredEntryPredictionDBFields defines the fields used regularly in ScoredEntryPredictions-related transactions
var scoredEntryPredictionDBFields = []string{
	"rankings",
	"score",
}

// ScoredEntryPredictionRepo defines our DB-backed ScoredEntryPredictions data store
type ScoredEntryPredictionRepo struct {
	db *sql.DB
}

// Insert inserts a new ScoredEntryPrediction into the database
func (s *ScoredEntryPredictionRepo) Insert(ctx context.Context, scoredEntryPrediction *domain.ScoredEntryPrediction) error {
	stmt := `INSERT INTO scored_entry_prediction (entry_prediction_id, standings_id,` + getDBFieldsStringFromFields(scoredEntryPredictionDBFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?)`

	rawRankings, err := json.Marshal(&scoredEntryPrediction.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.db.QueryContext(
		ctx,
		stmt,
		scoredEntryPrediction.EntryPredictionID,
		scoredEntryPrediction.StandingsID,
		rawRankings,
		scoredEntryPrediction.Score,
		scoredEntryPrediction.CreatedAt,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return nil
}

// Update updates an existing ScoredEntryPrediction in the database
func (s *ScoredEntryPredictionRepo) Update(ctx context.Context, scoredEntryPrediction *domain.ScoredEntryPrediction) error {
	stmt := `UPDATE scored_entry_prediction
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(scoredEntryPredictionDBFields) + `, updated_at = ?
				WHERE entry_prediction_id = ? AND standings_id = ?`

	rawRankings, err := json.Marshal(&scoredEntryPrediction.Rankings)
	if err != nil {
		return err
	}

	rows, err := s.db.QueryContext(
		ctx,
		stmt,
		rawRankings,
		scoredEntryPrediction.Score,
		scoredEntryPrediction.UpdatedAt,
		scoredEntryPrediction.EntryPredictionID,
		scoredEntryPrediction.StandingsID,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return nil
}

// Select retrieves ScoredEntryPredictions from our database based on the provided criteria
func (s *ScoredEntryPredictionRepo) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]domain.ScoredEntryPrediction, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT entry_prediction_id, standings_id, ` + getDBFieldsStringFromFields(scoredEntryPredictionDBFields) + `, created_at, updated_at
				FROM scored_entry_prediction ` + whereStmt

	rows, err := s.db.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var scoredEntryPredictions []domain.ScoredEntryPrediction
	var rawRankings []byte

	for rows.Next() {
		scoredEntryPrediction := domain.ScoredEntryPrediction{}

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
		return nil, domain.MissingDBRecordError{Err: errors.New("no scored entry predictions found")}
	}

	return scoredEntryPredictions, nil
}

// Exists determines whether a ScoredEntryPrediction with the provided ID exists in the database
func (s *ScoredEntryPredictionRepo) Exists(ctx context.Context, entryPredictionID, standingsID string) error {
	stmt := `SELECT COUNT(*) FROM scored_entry_prediction WHERE entry_prediction_id = ? AND standings_id = ?`

	row := s.db.QueryRowContext(ctx, stmt, entryPredictionID, standingsID)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return domain.MissingDBRecordError{Err: fmt.Errorf("scored entry prediction with standings id %s and entry prediction id %s: not found", standingsID, entryPredictionID)}
	}

	return nil
}

// SelectEntryCumulativeScoresByRealm retrieves the current score, total score and minimum score for each entry ID
// based on the provided realm name, season ID and round number
func (s *ScoredEntryPredictionRepo) SelectEntryCumulativeScoresByRealm(ctx context.Context, realmName string, seasonID string, roundNumber int) ([]domain.LeaderBoardRanking, error) {
	// compose main statement from provided partials
	// note that partials are ordering their scored entry predictions subqueries
	// by created_at date descending rather than updated_at date
	// this is because we will always have a created_at date, whereas updated_at may be null if the
	// scored entry prediction has been created without being updated yet
	// once a scored entry prediction has been created for a superseding entry_prediction/round_number combination,
	// none of the pre-existing scored entry predictions for this combination are operated on again
	// therefore, the created_at date for a superseding scored entry prediction should *always* occur later than
	// the updated_at date for any previous scored entry prediction for the same combination
	// ergo, the created_at date is fine to trust as a reliable real-world sort order
	stmt := fmt.Sprintf(`
		SELECT
			entry_with_total_score.entry_id,
			entry_with_total_score.total_score,
			entry_with_score_this_round.score_this_round,
			entry_with_max_score.max_score
		FROM (%s) AS entry_with_total_score
		INNER JOIN (%s) AS entry_with_score_this_round
			ON entry_with_total_score.entry_id = entry_with_score_this_round.entry_id
		INNER JOIN (%s) AS entry_with_max_score
			ON entry_with_total_score.entry_id = entry_with_max_score.entry_id
		ORDER BY
			entry_with_total_score.total_score DESC,
			entry_with_max_score.max_score DESC,
			entry_with_score_this_round.score_this_round DESC
		`,
		stmtSelectEntryWithTotalScore,
		stmtSelectEntryWithScoreThisRound,
		stmtSelectEntryWithMaxScore,
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
		// params for stmtSelectEntryWithMaxScore partial
		realmName,
		seasonID,
		roundNumber,
	}

	rows, err := s.db.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var lbRankings []domain.LeaderBoardRanking

	count := 0
	for rows.Next() {
		count++
		var (
			entryID      string
			totalScore   int
			currentScore int
			maxScore     int
		)
		if err := rows.Scan(
			&entryID,
			&totalScore,
			&currentScore,
			&maxScore,
		); err != nil {
			return nil, wrapDBError(err)
		}

		lbRankings = append(lbRankings, domain.LeaderBoardRanking{
			RankingWithScore: domain.RankingWithScore{
				Ranking: domain.Ranking{
					ID:       entryID,
					Position: count,
				},
				Score: currentScore,
			},
			MaxScore:   maxScore,
			TotalScore: totalScore,
		})
	}

	if len(lbRankings) == 0 {
		return nil, domain.MissingDBRecordError{Err: fmt.Errorf("no cumulative scores found for round %d in season %s", roundNumber, seasonID)}
	}

	return lbRankings, nil
}

// SelectByEntryIDAndRoundNumber retrieves ScoredEntryPredictions from our database based on the provided entry ID and round number
// ordered by their created_at date descending
func (s *ScoredEntryPredictionRepo) SelectByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) ([]domain.ScoredEntryPrediction, error) {
	stmt := `SELECT sep.entry_prediction_id, sep.standings_id, 
			` + getDBFieldsStringFromFieldsWithTablePrefix(scoredEntryPredictionDBFields, "sep") + `,
				sep.created_at, sep.updated_at
			FROM scored_entry_prediction sep
			INNER JOIN entry_prediction ep ON sep.entry_prediction_id = ep.id
			INNER JOIN standings s ON sep.standings_id = s.id
			WHERE ep.entry_id = ? AND s.round_number = ?
			ORDER BY sep.created_at DESC`

	rows, err := s.db.QueryContext(ctx, stmt, entryID, roundNumber)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var scoredEntryPredictions []domain.ScoredEntryPrediction
	var rawRankings []byte

	for rows.Next() {
		scoredEntryPrediction := domain.ScoredEntryPrediction{}

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
		return nil, domain.MissingDBRecordError{Err: fmt.Errorf("no scored entry predictions found: entry id %s, round number %d", entryID, roundNumber)}
	}

	return scoredEntryPredictions, nil
}

// NewScoredEntryPredictionRepo instantiates a new ScoredEntryPredictionRepo with the provided DB agent
func NewScoredEntryPredictionRepo(db *sql.DB) (*ScoredEntryPredictionRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db: %w", domain.ErrIsNil)
	}
	return &ScoredEntryPredictionRepo{db: db}, nil
}
