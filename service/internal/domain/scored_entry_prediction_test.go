package domain_test

import (
	"errors"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"strings"
	"testing"
	"time"
)

func TestNewScoredEntryPredictionAgent(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		tt := []struct {
			er   domain.EntryRepository
			epr  domain.EntryPredictionRepository
			sr   domain.StandingsRepository
			sepr domain.ScoredEntryPredictionRepository
		}{
			{nil, epr, sr, sepr},
			{er, nil, sr, sepr},
			{er, epr, nil, sepr},
			{er, epr, sr, nil},
		}

		for _, tc := range tt {
			_, gotErr := domain.NewScoredEntryPredictionAgent(tc.er, tc.epr, tc.sr, tc.sepr)
			if !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
			}
		}
	})
}

func TestScoredEntryPredictionAgent_CreateScoredEntryPrediction(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))
	scoredEntryPrediction := generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID)

	t.Run("create valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		createdScoredEntryPrediction, err := agent.CreateScoredEntryPrediction(ctx, scoredEntryPrediction)
		if err != nil {
			t.Fatal(err)
		}

		var emptyTime time.Time
		if createdScoredEntryPrediction.EntryPredictionID != scoredEntryPrediction.EntryPredictionID {
			expectedGot(t, scoredEntryPrediction.EntryPredictionID, createdScoredEntryPrediction.EntryPredictionID)
		}
		if createdScoredEntryPrediction.StandingsID != scoredEntryPrediction.StandingsID {
			expectedGot(t, scoredEntryPrediction.StandingsID, createdScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(createdScoredEntryPrediction.Rankings, scoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(scoredEntryPrediction.Rankings, createdScoredEntryPrediction.Rankings))
		}
		if createdScoredEntryPrediction.Score != scoredEntryPrediction.Score {
			expectedGot(t, scoredEntryPrediction.Score, createdScoredEntryPrediction.Score)
		}
		if createdScoredEntryPrediction.CreatedAt.Equal(emptyTime) {
			expectedNonEmpty(t, "CreatedAt")
		}
		if createdScoredEntryPrediction.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", *createdScoredEntryPrediction.UpdatedAt)
		}

		// inserting same scored entry prediction a second time should fail
		_, err = agent.CreateScoredEntryPrediction(ctx, scoredEntryPrediction)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create invalid scored entry prediction must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		missingEntryPredictionID := scoredEntryPrediction
		missingEntryPredictionID.EntryPredictionID = uuid.UUID{}
		_, err := agent.CreateScoredEntryPrediction(ctx, missingEntryPredictionID)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingStandingsID := scoredEntryPrediction
		missingStandingsID.StandingsID = uuid.UUID{}
		_, err = agent.CreateScoredEntryPrediction(ctx, missingStandingsID)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("create scored entry prediction for non-existent entry prediction or standings must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		invalidEntryPredictionID := scoredEntryPrediction
		invalidEntryPredictionID.EntryPredictionID = nonExistentID
		_, err = agent.CreateScoredEntryPrediction(ctx, invalidEntryPredictionID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		invalidStandingsID := scoredEntryPrediction
		invalidStandingsID.StandingsID = nonExistentID
		_, err = agent.CreateScoredEntryPrediction(ctx, invalidStandingsID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestScoredEntryPredictionAgent_UpdateScoredEntryPrediction(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))
	scoredEntryPrediction := insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID))

	t.Run("update valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		changedScoredEntryPrediction := scoredEntryPrediction
		changedScoredEntryPrediction.Score = 456
		changedScoredEntryPrediction.Rankings = domain.NewRankingWithScoreCollectionFromIDs([]string{"changedID_1", "changedID_2", "changedID_3"})

		updatedScoredEntryPrediction, err := agent.UpdateScoredEntryPrediction(ctx, changedScoredEntryPrediction)
		if err != nil {
			t.Fatal(err)
		}

		if updatedScoredEntryPrediction.EntryPredictionID != changedScoredEntryPrediction.EntryPredictionID {
			expectedGot(t, changedScoredEntryPrediction.EntryPredictionID, updatedScoredEntryPrediction.EntryPredictionID)
		}
		if updatedScoredEntryPrediction.StandingsID != changedScoredEntryPrediction.StandingsID {
			expectedGot(t, changedScoredEntryPrediction.StandingsID, updatedScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(updatedScoredEntryPrediction.Rankings, changedScoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(changedScoredEntryPrediction.Rankings, updatedScoredEntryPrediction.Rankings))
		}
		if updatedScoredEntryPrediction.Score != changedScoredEntryPrediction.Score {
			expectedGot(t, changedScoredEntryPrediction.Score, updatedScoredEntryPrediction.Score)
		}
		if !updatedScoredEntryPrediction.CreatedAt.Equal(changedScoredEntryPrediction.CreatedAt) {
			expectedGot(t, changedScoredEntryPrediction.CreatedAt, updatedScoredEntryPrediction.CreatedAt)
		}
		if updatedScoredEntryPrediction.UpdatedAt == nil {
			expectedNonEmpty(t, "UpdatedAt")
		}
	})

	t.Run("update non-existent scored entry prediction must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		nonExistentEntryPredictionID := scoredEntryPrediction
		nonExistentEntryPredictionID.EntryPredictionID = nonExistentID
		_, err = agent.UpdateScoredEntryPrediction(ctx, nonExistentEntryPredictionID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		nonExistentStandingsID := scoredEntryPrediction
		nonExistentStandingsID.StandingsID = nonExistentID
		_, err = agent.UpdateScoredEntryPrediction(ctx, nonExistentStandingsID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

// TODO - tests for ScoreEntryPredictionBasedOnStandings

func TestScoredEntryPredictionAgent_RetrieveScoredEntryPredictionByIDs(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))

	now := time.Now().Truncate(time.Second)
	scoredEntryPrediction := generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID)
	scoredEntryPrediction.CreatedAt = now
	scoredEntryPrediction = insertScoredEntryPrediction(t, scoredEntryPrediction)

	t.Run("retrieve existent scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedScoredEntryPrediction, err := agent.RetrieveScoredEntryPredictionByIDs(ctx, entryPrediction.ID.String(), standings.ID.String())
		if err != nil {
			t.Fatal(err)
		}

		if retrievedScoredEntryPrediction.EntryPredictionID != scoredEntryPrediction.EntryPredictionID {
			expectedGot(t, scoredEntryPrediction.EntryPredictionID, retrievedScoredEntryPrediction.EntryPredictionID)
		}
		if retrievedScoredEntryPrediction.StandingsID != scoredEntryPrediction.StandingsID {
			expectedGot(t, scoredEntryPrediction.StandingsID, retrievedScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(retrievedScoredEntryPrediction.Rankings, scoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(scoredEntryPrediction.Rankings, retrievedScoredEntryPrediction.Rankings))
		}
		if retrievedScoredEntryPrediction.Score != scoredEntryPrediction.Score {
			expectedGot(t, scoredEntryPrediction.Score, retrievedScoredEntryPrediction.Score)
		}
		if !retrievedScoredEntryPrediction.CreatedAt.In(utc).Equal(now.In(utc)) {
			expectedGot(t, now, retrievedScoredEntryPrediction.CreatedAt)
		}
		if retrievedScoredEntryPrediction.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", *retrievedScoredEntryPrediction.UpdatedAt)
		}
	})

	t.Run("retrieve non-existent scored entry prediction must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, nonExistentID.String(), scoredEntryPrediction.StandingsID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		_, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, scoredEntryPrediction.EntryPredictionID.String(), nonExistentID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestScoredEntryPredictionAgent_RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	standings := insertStandings(t, generateTestStandings(t))
	now := time.Now().Truncate(time.Second)

	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	scoredEntryPrediction := generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID)
	scoredEntryPrediction.CreatedAt = now
	scoredEntryPrediction = insertScoredEntryPrediction(t, scoredEntryPrediction)

	differentEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	mostRecentScoredEntryPrediction := generateTestScoredEntryPrediction(t, differentEntryPrediction.ID, standings.ID)
	mostRecentScoredEntryPrediction.CreatedAt = now.Add(time.Second) // created most recently
	mostRecentScoredEntryPrediction = insertScoredEntryPrediction(t, mostRecentScoredEntryPrediction)

	t.Run("retrieve latest scored entry prediction by existent entry id and round number must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedScoredEntryPrediction, err := agent.RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(
			ctx,
			entry.ID.String(),
			standings.RoundNumber,
		)
		if err != nil {
			t.Fatal(err)
		}

		if retrievedScoredEntryPrediction.EntryPredictionID != mostRecentScoredEntryPrediction.EntryPredictionID {
			expectedGot(t, mostRecentScoredEntryPrediction.EntryPredictionID, retrievedScoredEntryPrediction.EntryPredictionID)
		}
		if retrievedScoredEntryPrediction.StandingsID != mostRecentScoredEntryPrediction.StandingsID {
			expectedGot(t, mostRecentScoredEntryPrediction.StandingsID, retrievedScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(retrievedScoredEntryPrediction.Rankings, mostRecentScoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(mostRecentScoredEntryPrediction.Rankings, retrievedScoredEntryPrediction.Rankings))
		}
		if retrievedScoredEntryPrediction.Score != mostRecentScoredEntryPrediction.Score {
			expectedGot(t, mostRecentScoredEntryPrediction.Score, retrievedScoredEntryPrediction.Score)
		}
		if !retrievedScoredEntryPrediction.CreatedAt.In(utc).Equal(mostRecentScoredEntryPrediction.CreatedAt.In(utc)) {
			expectedGot(t, mostRecentScoredEntryPrediction.CreatedAt, retrievedScoredEntryPrediction.CreatedAt)
		}
		if retrievedScoredEntryPrediction.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", *retrievedScoredEntryPrediction.UpdatedAt)
		}
	})

	t.Run("retrieve latest scored entry prediction by non-existent entry id must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, nonExistentID.String(), standings.ID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve latest scored entry prediction by non-existent entry id must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, entry.ID.String(), nonExistentID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestTeamRankingsAsStrings(t *testing.T) {
	testRankingsWithScore := []domain.RankingWithScore{
		{
			Ranking: domain.Ranking{
				ID:       "AFC",
				Position: 1,
			},
			Score: 11111111,
		},
		{
			Ranking: domain.Ranking{
				ID:       "AVFC",
				Position: 2,
			},
			Score: 1111111,
		},
		{
			Ranking: domain.Ranking{
				ID:       "AFCB",
				Position: 3,
			},
			Score: 111111,
		},
		{
			Ranking: domain.Ranking{
				ID:       "BFC",
				Position: 4,
			},
			Score: 11111,
		},
		{
			Ranking: domain.Ranking{
				ID:       "BHAFC",
				Position: 5,
			},
			Score: 1111,
		},
		{
			Ranking: domain.Ranking{
				ID:       "CFC",
				Position: 6,
			},
			Score: 111,
		},
		{
			Ranking: domain.Ranking{
				ID:       "CPFC",
				Position: 7,
			},
			Score: 11,
		},
		{
			Ranking: domain.Ranking{
				ID:       "EFC",
				Position: 8,
			},
			Score: 1,
		},
	}

	testRankingsWithMeta := []domain.RankingWithMeta{
		{
			Ranking: domain.Ranking{
				ID:       "EFC",
				Position: 12,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "CPFC",
				Position: 34,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "CFC",
				Position: 56,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "BHAFC",
				Position: 78,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "BFC",
				Position: 90,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "AFCB",
				Position: 123,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "AVFC",
				Position: 456,
			},
		},
		{
			Ranking: domain.Ranking{
				ID:       "AFC",
				Position: 7890,
			},
		},
	}

	t.Run("generating strings from valid team rankings must succeed", func(t *testing.T) {
		actualStrings, err := domain.TeamRankingsAsStrings(testRankingsWithScore, testRankingsWithMeta, tc)
		if err != nil {
			t.Fatal(err)
		}

		expectedStrings := []string{
			"                          pts     pos",
			"-------------------------------------",
			"   1  Arsenal        11111111    7890",
			"   2  Villa           1111111     456",
			"   3  Bournemouth      111111     123",
			"   4  Burnley           11111      90",
			"   5  Brighton           1111      78",
			"   6  Chelsea             111      56",
			"   7  Palace               11      34",
			"   8  Everton               1      12",
			"-------------------------------------",
			"      TOTAL SCORE    12345678        ",
		}

		if strings.Join(actualStrings, ",") != strings.Join(expectedStrings, ",") {
			t.Fatal(gocmp.Diff(expectedStrings, actualStrings))
		}
	})

	t.Run("generating strings from empty scored entry prediction rankings must fail", func(t *testing.T) {
		_, err := domain.TeamRankingsAsStrings(nil, testRankingsWithMeta, tc)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
		_, err = domain.TeamRankingsAsStrings([]domain.RankingWithScore{}, testRankingsWithMeta, tc)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("generating strings from empty standings rankings must fail", func(t *testing.T) {
		_, err := domain.TeamRankingsAsStrings(testRankingsWithScore, nil, tc)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
		_, err = domain.TeamRankingsAsStrings(testRankingsWithScore, []domain.RankingWithMeta{}, tc)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("generating strings from team rankings with a score character length exceeding max must fail", func(t *testing.T) {
		var rwsThatShouldFail []domain.RankingWithScore
		rwsThatShouldFail = append(rwsThatShouldFail, testRankingsWithScore...)
		rwsThatShouldFail[0].Score = 100000000 // max char length is 9

		expectedErrorMessage := "total score character length cannot exceed 8: actual length 9"

		_, err := domain.TeamRankingsAsStrings(rwsThatShouldFail, testRankingsWithMeta, tc)
		if !cmp.Error(err, expectedErrorMessage)().Success() {
			expectedGot(t, expectedErrorMessage, err)
		}
	})

	t.Run("generating strings from team rankings with a position character length exceeding max must fail", func(t *testing.T) {
		var rwsThatShouldFail []domain.RankingWithScore
		rwsThatShouldFail = append(rwsThatShouldFail, testRankingsWithScore...)
		rwsThatShouldFail[0].Ranking.Position = 10000 // max char length is 4

		expectedErrorMessage := "prediction position character length cannot exceed 4: actual length 5"

		_, err := domain.TeamRankingsAsStrings(rwsThatShouldFail, testRankingsWithMeta, tc)
		if !cmp.Error(err, expectedErrorMessage)().Success() {
			expectedGot(t, expectedErrorMessage, err)
		}
	})

	t.Run("generating strings from team rankings with a non-existent ID must fail", func(t *testing.T) {
		rwsThatShouldFail := []domain.RankingWithScore{
			{
				Ranking: domain.Ranking{
					ID: "non_existent_team",
				},
			},
		}

		_, err := domain.TeamRankingsAsStrings(rwsThatShouldFail, []domain.RankingWithMeta{}, tc)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}
