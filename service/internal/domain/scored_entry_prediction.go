package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"prediction-league/service/internal/adapters/mysqldb/sqltypes"
	"strconv"
	"strings"
	"time"
)

// ScoredEntryPrediction provides a data type for an EntryPrediction that has been scored against a Standings
type ScoredEntryPrediction struct {
	EntryPredictionID uuid.UUID          `db:"entry_prediction_id"`
	StandingsID       uuid.UUID          `db:"standings_id"`
	Rankings          []RankingWithScore `db:"rankings"`
	Score             int                `db:"score"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         sqltypes.NullTime  `db:"updated_at"`
}

// ScoredEntryPredictionRepository defines the interface for transacting with our ScoredEntryPredictions data source
type ScoredEntryPredictionRepository interface {
	Insert(ctx context.Context, scoredEntryPrediction *ScoredEntryPrediction) error
	Update(ctx context.Context, scoredEntryPrediction *ScoredEntryPrediction) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]ScoredEntryPrediction, error)
	Exists(ctx context.Context, entryPredictionID, standingsID string) error
	SelectEntryCumulativeScoresByRealm(ctx context.Context, realmName string, seasonID string, roundNumber int) ([]LeaderBoardRanking, error)
	SelectByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) ([]ScoredEntryPrediction, error)
}

// ScoredEntryPredictionAgent defines the behaviours for handling ScoredEntryStandings
type ScoredEntryPredictionAgent struct {
	er   EntryRepository
	epr  EntryPredictionRepository
	sr   StandingsRepository
	sepr ScoredEntryPredictionRepository
}

// CreateScoredEntryPrediction handles the creation of a new ScoredEntryPrediction in the database
func (s *ScoredEntryPredictionAgent) CreateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction ScoredEntryPrediction) (ScoredEntryPrediction, error) {
	var emptyID uuid.UUID

	if scoredEntryPrediction.EntryPredictionID.String() == emptyID.String() {
		return ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"EntryPredictionID is empty",
		}}
	}

	if scoredEntryPrediction.StandingsID.String() == emptyID.String() {
		return ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"StandingsID is empty",
		}}
	}

	// ensure that entryPrediction exists
	if err := s.epr.ExistsByID(ctx, scoredEntryPrediction.EntryPredictionID.String()); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// ensure that standings exists
	if err := s.sr.ExistsByID(ctx, scoredEntryPrediction.StandingsID.String()); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.CreatedAt = time.Now().Truncate(time.Second)
	scoredEntryPrediction.UpdatedAt = sqltypes.NullTime{}

	// write scoredEntryPrediction to database
	if err := s.sepr.Insert(ctx, &scoredEntryPrediction); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// RetrieveScoredEntryPredictionByIDs handles the retrieval of an existing ScoredEntryPrediction in the database by its ID
func (s *ScoredEntryPredictionAgent) RetrieveScoredEntryPredictionByIDs(ctx context.Context, entryPredictionID, standingsID string) (ScoredEntryPrediction, error) {
	retrievedScoredEntryPredictions, err := s.sepr.Select(ctx, map[string]interface{}{
		"entry_prediction_id": entryPredictionID,
		"standings_id":        standingsID,
	}, false)
	if err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return retrievedScoredEntryPredictions[0], nil
}

// RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber handles the retrieval of
// the most recently created ScoredEntryPrediction by the provided entry ID and round number
func (s *ScoredEntryPredictionAgent) RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) (*ScoredEntryPrediction, error) {
	retrievedScoredEntryPredictions, err := s.sepr.SelectByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	// results are already ordered by created date descending
	return &retrievedScoredEntryPredictions[0], nil
}

// UpdateScoredEntryPrediction handles the updating of an existing ScoredEntryPrediction in the database
func (s *ScoredEntryPredictionAgent) UpdateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction ScoredEntryPrediction) (ScoredEntryPrediction, error) {
	// ensure the scoredEntryPrediction exists
	if err := s.sepr.Exists(
		ctx,
		scoredEntryPrediction.EntryPredictionID.String(),
		scoredEntryPrediction.StandingsID.String(),
	); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.UpdatedAt = sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	// write to database
	if err := s.sepr.Update(ctx, &scoredEntryPrediction); err != nil {
		return ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// NewScoredEntryPredictionAgent returns a new ScoredEntryPredictionAgent using the provided repositories
func NewScoredEntryPredictionAgent(er EntryRepository, epr EntryPredictionRepository, sr StandingsRepository, sepr ScoredEntryPredictionRepository) (*ScoredEntryPredictionAgent, error) {
	switch {
	case er == nil:
		return nil, fmt.Errorf("entry repository: %w", ErrIsNil)
	case epr == nil:
		return nil, fmt.Errorf("entry prediction repository: %w", ErrIsNil)
	case sr == nil:
		return nil, fmt.Errorf("standings repository: %w", ErrIsNil)
	case sepr == nil:
		return nil, fmt.Errorf("scored entry prediction repository: %w", ErrIsNil)
	}

	return &ScoredEntryPredictionAgent{
		er:   er,
		epr:  epr,
		sr:   sr,
		sepr: sepr,
	}, nil
}

// ScoreEntryPredictionBasedOnStandings generates a scored entry prediction from the provided entry prediction and standings
func ScoreEntryPredictionBasedOnStandings(
	entryPrediction EntryPrediction,
	standings Standings,
) (*ScoredEntryPrediction, error) {
	standingsRankingCollection := NewRankingCollectionFromRankingWithMetas(standings.Rankings)

	rws, err := CalculateRankingsScores(entryPrediction.Rankings, standingsRankingCollection)
	if err != nil {
		return nil, err
	}

	sep := ScoredEntryPrediction{
		EntryPredictionID: entryPrediction.ID,
		StandingsID:       standings.ID,
		Rankings:          *rws,
		Score:             rws.GetTotal(),
	}

	return &sep, nil
}

// TeamRankingsAsStrings returns the provided rankings as a slice of strings formatted with padding
func TeamRankingsAsStrings(sepRankings []RankingWithScore, standingsRankings []RankingWithMeta, tc TeamCollection) ([]string, error) {
	if len(sepRankings) == 0 {
		return nil, NotFoundError{errors.New("provided scored entry predictions are empty")}
	}

	if len(standingsRankings) == 0 {
		return nil, NotFoundError{errors.New("provided standings rankings are empty")}
	}

	// set the maximum length that our ranking ints are allowed to be when converted to strings
	const totalScoreLabel = "TOTAL SCORE"
	const maxPredictionPosLength = 4 // 9999
	const maxRankingScoreLength = 8  // 99999999
	const maxStandingsPosLength = 4  // 9999

	var (
		maxNameValueLength      = len(totalScoreLabel)
		sequentialPredictionPos []string
		sequentialTeamNames     []string
		sequentialRankingScores []string
		sequentialStandingsPos  []string
	)

	var getStandingsPosByTeamID = func(teamID string) (int, error) {
		for _, r := range standingsRankings {
			if r.ID == teamID {
				return r.Position, nil
			}
		}

		return 0, fmt.Errorf("no standings rankings entry found for team ID: %s", teamID)
	}

	totalScoreStr := strconv.Itoa(RankingWithScoreCollection(sepRankings).GetTotal())
	if len(totalScoreStr) > maxRankingScoreLength {
		return nil, fmt.Errorf("total score character length cannot exceed %d: actual length %d", maxRankingScoreLength, len(totalScoreStr))
	}

	for _, rws := range sepRankings {
		standingsPosVal, err := getStandingsPosByTeamID(rws.ID)
		if err != nil {
			return nil, err
		}

		predictionPosStr := strconv.Itoa(rws.Position)
		scoreStr := strconv.Itoa(rws.Score)
		standingsPosStr := strconv.Itoa(standingsPosVal)

		if len(predictionPosStr) > maxPredictionPosLength {
			return nil, fmt.Errorf("prediction position character length cannot exceed %d: actual length %d", maxPredictionPosLength, len(predictionPosStr))
		}
		if len(scoreStr) > maxRankingScoreLength {
			return nil, fmt.Errorf("ranking score character length cannot exceed %d: actual length %d", maxRankingScoreLength, len(scoreStr))
		}
		if len(standingsPosStr) > maxStandingsPosLength {
			return nil, fmt.Errorf("standingsRankings position character length cannot exceed %d: actual length %d", maxStandingsPosLength, len(standingsPosStr))
		}
		// retrieve the team so we get its name
		team, err := tc.GetByID(rws.ID)
		if err != nil {
			return nil, NotFoundError{err}
		}

		teamName := team.ShortName
		if len(teamName) > maxNameValueLength {
			maxNameValueLength = len(teamName)
		}

		// store values as strings
		sequentialPredictionPos = append(sequentialPredictionPos, predictionPosStr)
		sequentialTeamNames = append(sequentialTeamNames, teamName)
		sequentialRankingScores = append(sequentialRankingScores, scoreStr)
		sequentialStandingsPos = append(sequentialStandingsPos, standingsPosStr)
	}

	var fullStrings []string

	for i := 0; i < len(sepRankings); i++ {
		predictionPosStr := sequentialPredictionPos[i]
		teamNameStr := sequentialTeamNames[i]
		scoreStr := sequentialRankingScores[i]
		standingsPosStr := sequentialStandingsPos[i]

		// padding on left
		paddedPredictionPos := fmt.Sprintf("%s%s", strings.Repeat(" ", maxPredictionPosLength-len(predictionPosStr)), predictionPosStr)
		// padding on right
		paddedTeamName := fmt.Sprintf("%s%s", teamNameStr, strings.Repeat(" ", maxNameValueLength-len(teamNameStr)))
		// padding on left
		paddedScore := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingScoreLength-len(scoreStr)), scoreStr)
		// padding on left
		paddedStandingsPos := fmt.Sprintf("%s%s", strings.Repeat(" ", maxStandingsPosLength-len(standingsPosStr)), standingsPosStr)

		fullString := fmt.Sprintf("%s  %s    %s    %s", paddedPredictionPos, paddedTeamName, paddedScore, paddedStandingsPos)
		fullStrings = append(fullStrings, fullString)
	}

	// generate a header "row" with 'pts' and 'pos' given as "column" headers
	strPtsHeadingLabel := "pts"
	strPosHeadingLabel := "pos"
	// padding on left
	paddedPtsHeading := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingScoreLength-len(strPtsHeadingLabel)), strPtsHeadingLabel)
	// padding on left
	paddedPosHeading := fmt.Sprintf("%s%s", strings.Repeat(" ", maxStandingsPosLength-len(strPosHeadingLabel)), strPosHeadingLabel)
	header := fmt.Sprintf(
		"%s  %s    %s    %s",
		strings.Repeat(" ", maxPredictionPosLength),
		strings.Repeat(" ", maxNameValueLength),
		paddedPtsHeading,
		paddedPosHeading,
	)

	// generate a divider "row" - i.e. a sequence of '-' characters for the full length of the other full strings
	divider := fmt.Sprintf(
		"%s--%s----%s----%s",
		strings.Repeat("-", maxPredictionPosLength),
		strings.Repeat("-", maxNameValueLength),
		strings.Repeat("-", maxRankingScoreLength),
		strings.Repeat("-", maxStandingsPosLength),
	)

	// generate a total score "row" where total score is aligned with pts column
	// padding on right
	paddedTotalScoreLabel := fmt.Sprintf("%s%s", totalScoreLabel, strings.Repeat(" ", maxNameValueLength-len(totalScoreLabel)))
	// padding on left
	paddedTotalScore := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingScoreLength-len(totalScoreStr)), totalScoreStr)
	totalScoreRow := fmt.Sprintf(
		"%s  %s    %s    %s",
		strings.Repeat(" ", maxPredictionPosLength),
		paddedTotalScoreLabel,
		paddedTotalScore,
		strings.Repeat(" ", maxStandingsPosLength),
	)

	fullStrings = append([]string{header, divider}, append(fullStrings, divider, totalScoreRow)...)

	return fullStrings, nil
}
