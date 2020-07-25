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
func TeamRankingsAsStrings(sepRankings []models.RankingWithScore, standingsRankings []models.RankingWithMeta) ([]string, error) {
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

	teams := datastore.Teams

	var getStandingsPosByTeamID = func(teamID string) (int, error) {
		for _, r := range standingsRankings {
			if r.ID == teamID {
				return r.Position, nil
			}
		}

		return 0, fmt.Errorf("no standings rankings entry found for team ID: %s", teamID)
	}

	totalScoreStr := strconv.Itoa(models.RankingWithScoreCollection(sepRankings).GetTotal())
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
		team, err := teams.GetByID(rws.ID)
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
