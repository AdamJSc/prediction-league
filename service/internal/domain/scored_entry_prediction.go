package domain

import (
	"context"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"strconv"
	"strings"
	"time"
)

// ScoredEntryPredictionAgentInjector defines the dependencies required by our ScoredEntryPredictionAgent
type ScoredEntryPredictionAgentInjector interface {
	MySQL() coresql.Agent
}

// ScoredEntryPredictionAgent defines the behaviours for handling ScoredEntryStandings
type ScoredEntryPredictionAgent struct {
	ScoredEntryPredictionAgentInjector
}

// CreateScoredEntryPrediction handles the creation of a new ScoredEntryPrediction in the database
func (s ScoredEntryPredictionAgent) CreateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction models.ScoredEntryPrediction) (models.ScoredEntryPrediction, error) {
	db := s.MySQL()

	var emptyID uuid.UUID

	if scoredEntryPrediction.EntryPredictionID.String() == emptyID.String() {
		return models.ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"EntryPredictionID is empty",
		}}
	}

	if scoredEntryPrediction.StandingsID.String() == emptyID.String() {
		return models.ScoredEntryPrediction{}, ValidationError{Reasons: []string{
			"StandingsID is empty",
		}}
	}

	// ensure that entryPrediction exists
	entryPredictionRepo := repositories.NewEntryPredictionDatabaseRepository(db)
	if err := entryPredictionRepo.ExistsByID(ctx, scoredEntryPrediction.EntryPredictionID.String()); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// ensure that standings exists
	standingsRepo := repositories.NewStandingsDatabaseRepository(db)
	if err := standingsRepo.ExistsByID(ctx, scoredEntryPrediction.StandingsID.String()); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.CreatedAt = time.Now().Truncate(time.Second)
	scoredEntryPrediction.UpdatedAt = sqltypes.NullTime{}

	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(db)

	// write scoredEntryPrediction to database
	if err := scoredEntryPredictionRepo.Insert(ctx, &scoredEntryPrediction); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// RetrieveScoredEntryPredictionByIDs handles the retrieval of an existing ScoredEntryPrediction in the database by its ID
func (s ScoredEntryPredictionAgent) RetrieveScoredEntryPredictionByIDs(ctx context.Context, entryPredictionID, standingsID string) (models.ScoredEntryPrediction, error) {
	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(s.MySQL())

	retrievedScoredEntryPredictions, err := scoredEntryPredictionRepo.Select(ctx, map[string]interface{}{
		"entry_prediction_id": entryPredictionID,
		"standings_id":        standingsID,
	}, false)
	if err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return retrievedScoredEntryPredictions[0], nil
}

// RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber handles the retrieval of
// the most recently created ScoredEntryPrediction by the provided entry ID and round number
func (s ScoredEntryPredictionAgent) RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(ctx context.Context, entryID string, roundNumber int) (*models.ScoredEntryPrediction, error) {
	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(s.MySQL())

	retrievedScoredEntryPredictions, err := scoredEntryPredictionRepo.SelectByEntryIDAndRoundNumber(ctx, entryID, roundNumber)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	// results are already ordered by created date descending
	return &retrievedScoredEntryPredictions[0], nil
}

// UpdateScoredEntryPrediction handles the updating of an existing ScoredEntryPrediction in the database
func (s ScoredEntryPredictionAgent) UpdateScoredEntryPrediction(ctx context.Context, scoredEntryPrediction models.ScoredEntryPrediction) (models.ScoredEntryPrediction, error) {
	scoredEntryPredictionRepo := repositories.NewScoredEntryPredictionDatabaseRepository(s.MySQL())

	// ensure the scoredEntryPrediction exists
	if err := scoredEntryPredictionRepo.Exists(
		ctx,
		scoredEntryPrediction.EntryPredictionID.String(),
		scoredEntryPrediction.StandingsID.String(),
	); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	scoredEntryPrediction.UpdatedAt = sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	// write to database
	if err := scoredEntryPredictionRepo.Update(ctx, &scoredEntryPrediction); err != nil {
		return models.ScoredEntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return scoredEntryPrediction, nil
}

// TeamRankingsAsStrings returns the provided rankings as a slice of strings formatted with padding
func TeamRankingsAsStrings(rankings []models.RankingWithScore) ([]string, error) {
	var (
		maxTeamNameLength       int
		sequentialTeamNames     []string
		sequentialRankingScores []string
		sequentialRankingPos    []string
	)

	if len(rankings) == 0 {
		return nil, NotFoundError{errors.New("provided rankings are empty")}
	}

	// set the maximum length that our ranking ints are allowed to be when converted to strings
	const maxRankingScoreLength = 4 // 9999
	const maxRankingPosLength = 4   // 9999

	teams := datastore.Teams

	for _, rws := range rankings {
		scoreStr := strconv.Itoa(rws.Score)
		posStr := strconv.Itoa(rws.Position)

		if len(scoreStr) > maxRankingScoreLength {
			return nil, fmt.Errorf("ranking score character length cannot exceed %d: actual length %d", maxRankingScoreLength, len(scoreStr))
		}
		if len(posStr) > maxRankingPosLength {
			return nil, fmt.Errorf("ranking position character length cannot exceed %d: actual length %d", maxRankingPosLength, len(posStr))
		}
		// retrieve the team so we get its name
		team, err := teams.GetByID(rws.ID)
		if err != nil {
			return nil, NotFoundError{err}
		}
		if len(team.Name) > maxTeamNameLength {
			maxTeamNameLength = len(team.Name)
		}

		// store values as strings
		sequentialTeamNames = append(sequentialTeamNames, team.Name)
		sequentialRankingScores = append(sequentialRankingScores, scoreStr)
		sequentialRankingPos = append(sequentialRankingPos, posStr)
	}

	var fullStrings []string

	for i := 0; i < len(rankings); i++ {
		strTeamName := sequentialTeamNames[i]
		strScore := sequentialRankingScores[i]
		strPos := sequentialRankingPos[i]

		// padding on right
		paddedTeamName := fmt.Sprintf("%s%s", strTeamName, strings.Repeat(" ", maxTeamNameLength-len(strTeamName)))
		// padding on left
		paddedScore := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingScoreLength-len(strScore)), strScore)
		// padding on left
		paddedPos := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingPosLength-len(strPos)), strPos)

		fullString := fmt.Sprintf("%s    %s    %s", paddedTeamName, paddedScore, paddedPos)
		fullStrings = append(fullStrings, fullString)
	}

	// generate a header "row" with 'pts' and 'pos' given as "column" headers
	strScoreHeader := "pts"
	strPosHeader := "pos"
	paddedScoreHeader := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingScoreLength-len(strScoreHeader)), strScoreHeader)
	paddedPosHeader := fmt.Sprintf("%s%s", strings.Repeat(" ", maxRankingPosLength-len(strPosHeader)), strPosHeader)
	header := fmt.Sprintf("%s    %s    %s", strings.Repeat(" ", maxTeamNameLength), paddedScoreHeader, paddedPosHeader)

	// generate a divider "row" - i.e. a sequence of '-' characters for the full length of the other full strings
	divider := fmt.Sprintf(
		"%s----%s----%s",
		strings.Repeat("-", maxTeamNameLength),
		strings.Repeat("-", maxRankingScoreLength),
		strings.Repeat("-", maxRankingPosLength),
	)

	fullStrings = append([]string{header, divider}, fullStrings...)

	return fullStrings, nil
}
