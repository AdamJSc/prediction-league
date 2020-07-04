package domain_test

import (
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"testing"
	"time"
)

func TestLeaderBoardAgent_RetrieveLeaderBoardBySeasonAndRoundNumber(t *testing.T) {
	defer truncate(t)

	// <-- seed standings rounds -->

	// start at round 2, so that we can check round 1 produces an empty leaderboard
	var standingsRounds = make(map[int]models.Standings)
	for i := 2; i <= 4; i++ {
		s := generateTestStandings(t)
		s.RoundNumber = i
		standingsRounds[i] = insertStandings(t, s)
	}

	// <-- seed entries that should appear within the leaderboard -->

	harryEntry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))
	jamieEntry := insertEntry(t, generateTestEntry(t,
		"Jamie Redknapp",
		"MrJamieR",
		"jamie.redknapp@football.net",
	))
	frankEntry := insertEntry(t, generateTestEntry(t,
		"Frank Lampard",
		"FrankieLamps",
		"frank.lampard@football.net",
	))
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
		min: 123,
		mid: 124,
		max: 125,
	}
	jamieScores := tieredScores{
		min: 122,
		mid: 125,
		max: 128,
	}
	frankScores := tieredScores{
		min: 121,
		mid: 123,
		max: 125,
	}

	// <-- seed scored entry predictions -->

	var harryScoredEntryPredictions = make(map[int]models.ScoredEntryPrediction)
	var harryScoreSequence = []int{harryScores.min, harryScores.mid, harryScores.max}
	for i := 2; i <= 4; i++ {
		idx := i - 2
		sep := generateTestScoredEntryPrediction(t, harryEntryPrediction.ID, standingsRounds[i].ID)
		sep.Score = harryScoreSequence[idx]
		sep.UpdatedAt.Valid = true
		sep.UpdatedAt.Time = time.Now().Add(time.Duration(i) * 24 * time.Hour)
		harryScoredEntryPredictions[i] = insertScoredEntryPrediction(t, sep)
	}
	var jamieScoredEntryPredictions = make(map[int]models.ScoredEntryPrediction)
	var jamieScoreSequence = []int{jamieScores.max, jamieScores.min, jamieScores.mid}
	for i := 2; i <= 4; i++ {
		idx := i - 2
		sep := generateTestScoredEntryPrediction(t, jamieEntryPrediction.ID, standingsRounds[i].ID)
		sep.Score = jamieScoreSequence[idx]
		sep.UpdatedAt.Valid = true
		sep.UpdatedAt.Time = time.Now().Add(time.Duration(i) * 24 * time.Hour)
		jamieScoredEntryPredictions[i] = insertScoredEntryPrediction(t, sep)
	}
	var frankScoredEntryPredictions = make(map[int]models.ScoredEntryPrediction)
	var frankScoreSequence = []int{frankScores.mid, frankScores.max, frankScores.min}
	for i := 2; i <= 4; i++ {
		idx := i - 2
		sep := generateTestScoredEntryPrediction(t, frankEntryPrediction.ID, standingsRounds[i].ID)
		sep.Score = frankScoreSequence[idx]
		sep.UpdatedAt.Valid = true
		sep.UpdatedAt.Time = time.Now().Add(time.Duration(i) * 24 * time.Hour)
		frankScoredEntryPredictions[i] = insertScoredEntryPrediction(t, sep)
	}

	// <-- seed some scored entry predictions that should never appear within the leaderboard
	// because their `updated at` date occurs before the prediction we already have in our map
	// for the given round number (leaderboard should include the most recent one for each round) -->

	// harry
	ep := insertEntryPrediction(t, generateTestEntryPrediction(t, harryEntry.ID))
	sep := generateTestScoredEntryPrediction(t, ep.ID, standingsRounds[2].ID)
	sep.Score = 100000 // something ludicrous
	sep.UpdatedAt.Valid = true
	sep.UpdatedAt.Time = harryScoredEntryPredictions[2].UpdatedAt.Time.Add(-time.Hour) // occurs BEFORE harry's existing round 2 prediction
	insertScoredEntryPrediction(t, sep)
	// jamie
	ep = insertEntryPrediction(t, generateTestEntryPrediction(t, jamieEntry.ID))
	sep = generateTestScoredEntryPrediction(t, ep.ID, standingsRounds[3].ID)
	sep.Score = 123456789 // something ludicrous
	sep.UpdatedAt.Valid = true
	sep.UpdatedAt.Time = jamieScoredEntryPredictions[3].UpdatedAt.Time.Add(-time.Hour) // occurs BEFORE jamie's existing round 3 prediction
	insertScoredEntryPrediction(t, sep)
	// frank
	ep = insertEntryPrediction(t, generateTestEntryPrediction(t, frankEntry.ID))
	sep = generateTestScoredEntryPrediction(t, ep.ID, standingsRounds[4].ID)
	sep.Score = 55378008 // something ludicrous
	sep.UpdatedAt.Valid = true
	sep.UpdatedAt.Time = frankScoredEntryPredictions[4].UpdatedAt.Time.Add(-time.Hour) // occurs BEFORE frank's existing round 4 prediction
	insertScoredEntryPrediction(t, sep)

	// <-- seed entries that definitely should never appear within the leaderboard -->

	robbieEntry := generateTestEntry(t,
		"Robbie Savage",
		"MrRobbieS",
		"robbie.savage@football.net",
	)
	robbieEntry.SeasonID = "NotSameID" // different season ID to the others
	robbieEntry = insertEntry(t, robbieEntry)
	robbieEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, robbieEntry.ID))
	insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, robbieEntryPrediction.ID, standingsRounds[2].ID))
	joeyEntry := generateTestEntry(t,
		"Joey Barton",
		"MrJoeyB",
		"joey.barton@football.net",
	)
	joeyEntry.RealmName = "NotSameRealm" // different realm name to the others
	joeyEntry = insertEntry(t, joeyEntry)
	joeyEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, joeyEntry.ID))
	insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, joeyEntryPrediction.ID, standingsRounds[2].ID))

	// store season ID arbitrarily from one of the valid entries
	seasonID := harryEntry.SeasonID

	agent := domain.LeaderBoardAgent{
		LeaderBoardAgentInjector: injector{db: db},
	}

	t.Run("retrieve leaderboard for a round number that pre-dates the rounds we have must return empty leaderboard", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// empty leaderboard should be sorted by entrants' nicknames
		var harryRanking, jamieRanking, frankRanking models.LeaderBoardRanking
		frankRanking.Position = 1
		frankRanking.ID = frankEntry.ID.String()
		harryRanking.Position = 2
		harryRanking.ID = harryEntry.ID.String()
		jamieRanking.Position = 3
		jamieRanking.ID = jamieEntry.ID.String()

		expectedLeaderBoard := &models.LeaderBoard{
			RoundNumber: 1,
			Rankings:    []models.LeaderBoardRanking{frankRanking, harryRanking, jamieRanking},
		}

		actualLeaderboard, err := agent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, 1)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(actualLeaderboard, expectedLeaderBoard) {
			t.Fatal(cmp.Diff(expectedLeaderBoard, actualLeaderboard))
		}
	})
}
