package domain_test

import (
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
)

func TestNewLeaderBoardAgent(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		tt := []struct {
			er      domain.EntryRepository
			epr     domain.EntryPredictionRepository
			sr      domain.StandingsRepository
			sepr    domain.ScoredEntryPredictionRepository
			sc      domain.SeasonCollection
			wantErr error
		}{
			{nil, epr, sr, sepr, sc, domain.ErrIsNil},
			{er, nil, sr, sepr, sc, domain.ErrIsNil},
			{er, epr, nil, sepr, sc, domain.ErrIsNil},
			{er, epr, sr, nil, sc, domain.ErrIsNil},
			{er, epr, sr, sepr, nil, domain.ErrIsNil},
			{er, epr, sr, sepr, sc, nil},
		}

		for idx, tc := range tt {
			agent, gotErr := domain.NewLeaderBoardAgent(tc.er, tc.epr, tc.sr, tc.sepr, tc.sc)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && agent == nil {
				t.Fatalf("tc #%d: want non-empty agent, got nil", idx)
			}
		}
	})
}

func TestLeaderBoardAgent_RetrieveLeaderBoardBySeasonAndRoundNumber(t *testing.T) {
	t.Cleanup(truncate)

	now := time.Now().Truncate(time.Second)

	// <-- seed standings rounds -->

	// start at round 2, so that we can check round 1 produces an empty leaderboard
	var standingsRounds = make(map[int]domain.Standings)
	for i := 2; i <= 4; i++ {
		s := generateTestStandings(t)
		s.SeasonID = testSeason.ID
		s.RoundNumber = i
		s.CreatedAt = now.Add(time.Duration(i) * 24 * time.Hour)
		s = insertStandings(t, s)
		switch {
		case i > 2:
			// later on, we can check that round 2's leaderboard has a last updated date that
			// matches the created_at date of standings round 2
			// otherwise, leaderboard should match the standings round's updated_at date instead
			addMin := s.CreatedAt.Add(time.Minute)
			s.UpdatedAt = &addMin
			s = updateStandings(t, s)
		}
		standingsRounds[i] = s
	}

	// <-- seed entries that should appear within the leaderboard -->

	harryEntry := generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	)
	harryEntry.ApprovedAt = &now
	harryEntry = insertEntry(t, harryEntry)

	jamieEntry := generateTestEntry(t,
		"Jamie Redknapp",
		"MrJamieR",
		"jamie.redknapp@football.net",
	)
	jamieEntry.ApprovedAt = &now
	jamieEntry = insertEntry(t, jamieEntry)

	frankEntry := generateTestEntry(t,
		"Frank Lampard",
		"FrankieLamps",
		"frank.lampard@football.net",
	)
	frankEntry.ApprovedAt = &now
	frankEntry = insertEntry(t, frankEntry)

	harryEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, harryEntry.ID))
	jamieEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, jamieEntry.ID))
	frankEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, frankEntry.ID))

	// <-- define scores in advance for convenience -->

	type tieredScores struct {
		min int
		mid int
		max int
	}
	harryScores := tieredScores{
		min: 122,
		mid: 124,
		max: 126,
	}
	jamieScores := tieredScores{
		min: 121,
		mid: 123,
		max: 125,
	}
	frankScores := tieredScores{
		min: 119,
		mid: 124,
		max: 125,
	}

	// <-- seed scored entry predictions -->

	var harryScoredEntryPredictions = make(map[int]domain.ScoredEntryPrediction)
	var harryScoreSequence = []int{harryScores.min, harryScores.mid, harryScores.max}
	for i := 2; i <= 4; i++ {
		idx := i - 2
		sep := generateTestScoredEntryPrediction(t, harryEntryPrediction.ID, standingsRounds[i].ID)
		sep.Score = harryScoreSequence[idx]
		sep.CreatedAt = time.Now().Add(time.Duration(i) * 24 * time.Hour)
		harryScoredEntryPredictions[i] = insertScoredEntryPrediction(t, sep)
	}
	var jamieScoredEntryPredictions = make(map[int]domain.ScoredEntryPrediction)
	var jamieScoreSequence = []int{jamieScores.max, jamieScores.min, jamieScores.mid}
	for i := 2; i <= 4; i++ {
		idx := i - 2
		sep := generateTestScoredEntryPrediction(t, jamieEntryPrediction.ID, standingsRounds[i].ID)
		sep.Score = jamieScoreSequence[idx]
		sep.CreatedAt = time.Now().Add(time.Duration(i) * 24 * time.Hour)
		jamieScoredEntryPredictions[i] = insertScoredEntryPrediction(t, sep)
	}
	var frankScoredEntryPredictions = make(map[int]domain.ScoredEntryPrediction)
	var frankScoreSequence = []int{frankScores.mid, frankScores.max, frankScores.min}
	for i := 2; i <= 4; i++ {
		idx := i - 2
		sep := generateTestScoredEntryPrediction(t, frankEntryPrediction.ID, standingsRounds[i].ID)
		sep.Score = frankScoreSequence[idx]
		sep.CreatedAt = time.Now().Add(time.Duration(i) * 24 * time.Hour)
		frankScoredEntryPredictions[i] = insertScoredEntryPrediction(t, sep)
	}

	// <-- seed some scored entry predictions that should never appear within the leaderboard
	// because their `updated at` date occurs before the prediction we already have in our map
	// for the given round number (leaderboard should include the most recent one for each round) -->

	// harry
	ep := insertEntryPrediction(t, generateTestEntryPrediction(t, harryEntry.ID))
	sep := generateTestScoredEntryPrediction(t, ep.ID, standingsRounds[2].ID)
	sep.Score = 100000                                                       // something ludicrous
	sep.CreatedAt = harryScoredEntryPredictions[2].CreatedAt.Add(-time.Hour) // occurs BEFORE harry's existing round 2 prediction
	insertScoredEntryPrediction(t, sep)
	// jamie
	ep = insertEntryPrediction(t, generateTestEntryPrediction(t, jamieEntry.ID))
	sep = generateTestScoredEntryPrediction(t, ep.ID, standingsRounds[3].ID)
	sep.Score = 123456789                                                    // something ludicrous
	sep.CreatedAt = jamieScoredEntryPredictions[3].CreatedAt.Add(-time.Hour) // occurs BEFORE jamie's existing round 3 prediction
	insertScoredEntryPrediction(t, sep)
	// frank
	ep = insertEntryPrediction(t, generateTestEntryPrediction(t, frankEntry.ID))
	sep = generateTestScoredEntryPrediction(t, ep.ID, standingsRounds[4].ID)
	sep.Score = 55378008                                                     // something ludicrous
	sep.CreatedAt = frankScoredEntryPredictions[4].CreatedAt.Add(-time.Hour) // occurs BEFORE frank's existing round 4 prediction
	insertScoredEntryPrediction(t, sep)

	// <-- seed entries that definitely should never appear within the leaderboard -->

	robbieEntry := generateTestEntry(t,
		"Robbie Savage",
		"MrRobbieS",
		"robbie.savage@football.net",
	)
	robbieEntry.ApprovedAt = &now
	robbieEntry.SeasonID = "NotSameID" // different season ID to the others
	robbieEntry = insertEntry(t, robbieEntry)
	robbieEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, robbieEntry.ID))
	insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, robbieEntryPrediction.ID, standingsRounds[2].ID))

	joeyEntry := generateTestEntry(t,
		"Joey Barton",
		"MrJoeyB",
		"joey.barton@football.net",
	)
	joeyEntry.ApprovedAt = &now
	joeyEntry.RealmName = "NotSameRealm" // different realm name to the others
	joeyEntry = insertEntry(t, joeyEntry)
	joeyEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, joeyEntry.ID))
	insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, joeyEntryPrediction.ID, standingsRounds[2].ID))

	ericEntry := generateTestEntry(t,
		"Eric Cantona",
		"MonsieurEric",
		"eric.cantona@football.net",
	)
	// no changes to eric, he doesn't have an approved at date, so shouldn't appear in the leaderboard
	ericEntry = insertEntry(t, ericEntry)
	ericEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, ericEntry.ID))
	insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, ericEntryPrediction.ID, standingsRounds[2].ID))

	// store season ID arbitrarily from one of the valid entries (they should all belong to the same one, apart from robbie)
	seasonID := harryEntry.SeasonID

	lbAgent, err := domain.NewLeaderBoardAgent(er, epr, sr, sepr, sc)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("harry's entry: %s", harryEntry.ID.String())
	t.Logf("jamie's entry: %s", jamieEntry.ID.String())
	t.Logf("frank's entry: %s", frankEntry.ID.String())
	t.Logf("robbie's entry: %s", robbieEntry.ID.String())
	t.Logf("joey's entry: %s", joeyEntry.ID.String())
	t.Logf("eric's entry: %s", ericEntry.ID.String())

	t.Run("retrieve leaderboard for a round number that pre-dates the rounds we have must return empty leaderboard", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// empty leaderboard should be sorted by entrants' nicknames
		expectedLeaderBoard := &domain.LeaderBoard{
			RoundNumber: 1,
			Rankings: []domain.LeaderBoardRanking{
				generateTestLeaderBoardRanking(1, 0, frankEntry.ID.String(), 0, 0, 0),
				generateTestLeaderBoardRanking(2, 0, harryEntry.ID.String(), 0, 0, 0),
				generateTestLeaderBoardRanking(3, 0, jamieEntry.ID.String(), 0, 0, 0),
			},
		}

		actualLeaderBoard, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, 1)
		if err != nil {
			t.Fatal(err)
		}

		if !gocmp.Equal(actualLeaderBoard, expectedLeaderBoard) {
			t.Fatal(gocmp.Diff(expectedLeaderBoard, actualLeaderBoard))
		}
	})

	t.Run("retrieve leaderboard for first proper round number must return expected leaderboard", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		lastUpdated := standingsRounds[2].CreatedAt // standingsRound[2].UpdatedAt is empty
		expectedLeaderBoard := &domain.LeaderBoard{
			RoundNumber: 2,
			Rankings: []domain.LeaderBoardRanking{
				// total 125, max 125, current 125
				generateTestLeaderBoardRanking(1, 0, jamieEntry.ID.String(), jamieScores.max, jamieScores.max, jamieScores.max),
				// total 124, max 124, current 124
				generateTestLeaderBoardRanking(2, 0, frankEntry.ID.String(), frankScores.mid, frankScores.mid, frankScores.mid),
				// total 122, max 122, current 122
				generateTestLeaderBoardRanking(3, 0, harryEntry.ID.String(), harryScores.min, harryScores.min, harryScores.min),
			},
			LastUpdated: &lastUpdated,
		}

		actualLeaderBoard, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, 2)
		if err != nil {
			t.Fatal(err)
		}

		if !gocmp.Equal(actualLeaderBoard, expectedLeaderBoard) {
			t.Fatal(gocmp.Diff(expectedLeaderBoard, actualLeaderBoard))
		}
	})

	t.Run("retrieve leaderboard for second proper round number must return expected leaderboard", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedLeaderBoard := &domain.LeaderBoard{
			RoundNumber: 3,
			Rankings: []domain.LeaderBoardRanking{
				// total 249, max 125, current 125, movement +1 on previous round
				generateTestLeaderBoardRanking(1, 1, frankEntry.ID.String(), frankScores.max, frankScores.max, frankScores.mid+frankScores.max),
				// total 246, max 125, current 121, movement -1 on previous round
				generateTestLeaderBoardRanking(2, -1, jamieEntry.ID.String(), jamieScores.min, jamieScores.max, jamieScores.max+jamieScores.min),
				// total 246, max 124, current 124, no movement on previous round
				generateTestLeaderBoardRanking(3, 0, harryEntry.ID.String(), harryScores.mid, harryScores.mid, harryScores.min+harryScores.mid),
			},
			LastUpdated: standingsRounds[3].UpdatedAt,
		}

		actualLeaderBoard, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, 3)
		if err != nil {
			t.Fatal(err)
		}

		if !gocmp.Equal(actualLeaderBoard, expectedLeaderBoard) {
			t.Fatal(gocmp.Diff(expectedLeaderBoard, actualLeaderBoard))
		}
	})

	t.Run("retrieve leaderboard for third proper round number must return expected leaderboard", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedLeaderBoard := &domain.LeaderBoard{
			RoundNumber: 4,
			Rankings: []domain.LeaderBoardRanking{
				// total 372, max 126, current 126, movement +2 on previous round
				generateTestLeaderBoardRanking(1, 2, harryEntry.ID.String(), harryScores.max, harryScores.max, harryScores.min+harryScores.mid+harryScores.max),
				// total 369, max 125, current 123, no movement on previous round
				generateTestLeaderBoardRanking(2, 0, jamieEntry.ID.String(), jamieScores.mid, jamieScores.max, jamieScores.max+jamieScores.min+jamieScores.mid),
				// total 368, max 125, current 119, movement -2 on previous round
				generateTestLeaderBoardRanking(3, -2, frankEntry.ID.String(), frankScores.min, frankScores.max, frankScores.mid+frankScores.max+frankScores.min),
			},
			LastUpdated: standingsRounds[4].UpdatedAt,
		}

		actualLeaderBoard, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, 4)
		if err != nil {
			t.Fatal(err)
		}

		if !gocmp.Equal(actualLeaderBoard, expectedLeaderBoard) {
			t.Fatal(gocmp.Diff(expectedLeaderBoard, actualLeaderBoard))
		}
	})

	t.Run("retrieve leaderboard for non-existent round number must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, 5)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve leaderboard for non-existent season ID must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := lbAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, "not_a_real_season_id", 2)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

// generateTestLeaderBoardRanking provides a helper function for generating a leaderboard ranking based on the provided values
func generateTestLeaderBoardRanking(position, movement int, entryID string, score, maxScore, totalScore int) domain.LeaderBoardRanking {
	return domain.LeaderBoardRanking{
		RankingWithScore: domain.RankingWithScore{
			Ranking: domain.Ranking{
				ID:       entryID,
				Position: position,
			},
			Score: score,
		},
		MaxScore:   maxScore,
		TotalScore: totalScore,
		Movement:   movement,
	}
}
