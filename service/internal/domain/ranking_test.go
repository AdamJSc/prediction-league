package domain_test

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"math/rand"
	"prediction-league/service/internal/domain"
	"strings"
	"testing"
	"time"
)

var rankingIDs = []string{"Pitman", "Wilson", "Hayter", "Pugh", "King"}

func TestRankingCollection_GetIDs(t *testing.T) {
	var rc domain.RankingCollection

	for idx, id := range rankingIDs {
		rc = append(rc, domain.Ranking{ID: id, Position: idx + 1})
	}

	t.Run("getting ids from ranking collection must match original slice of ids", func(t *testing.T) {
		ids := rc.GetIDs()
		if !cmp.Equal(ids, rankingIDs) {
			t.Fatal(cmp.Diff(ids, rankingIDs))
		}
	})
}

func TestRankingCollection_GetByID(t *testing.T) {
	var rc domain.RankingCollection

	for idx, id := range rankingIDs {
		rc = append(rc, domain.Ranking{ID: id, Position: idx + 1})
	}

	t.Run("get ranking by existing id must succeed", func(t *testing.T) {
		idToFind := "Hayter"

		ranking, err := rc.GetByID(idToFind)
		if err != nil {
			t.Fatal(err)
		}

		if ranking.ID != idToFind {
			expectedGot(t, idToFind, ranking.ID)
		}

		if ranking.Position != 3 {
			expectedGot(t, 3, ranking.Position)
		}
	})

	t.Run("get ranking by non-existent id must fail", func(t *testing.T) {
		idToFind := "non_existent_id"
		expectedError := fmt.Errorf("ranking id %s: not found", idToFind)

		_, err := rc.GetByID(idToFind)
		if !cmp.Equal(expectedError.Error(), err.Error()) {
			expectedGot(t, expectedError.Error(), err.Error())
		}
	})
}

func TestNewRankingCollectionFromIDs(t *testing.T) {
	var rankings = domain.NewRankingCollectionFromIDs(rankingIDs)

	t.Run("creating a new ranking collection from ids must successfully populate the expected positions", func(t *testing.T) {
		for idx, ranking := range rankings {
			if ranking.ID != rankingIDs[idx] {
				expectedGot(t, rankingIDs[idx], ranking.ID)
			}
			if ranking.Position != idx+1 {
				expectedGot(t, idx+1, ranking.Position)
			}
		}
	})

	t.Run("marshaling and unmarshaling a ranking collection should retain the ids only", func(t *testing.T) {
		var joined = fmt.Sprintf(`["%s"]`, strings.Join(rankingIDs, `","`))

		marshaled, err := json.Marshal(&rankings)
		if err != nil {
			t.Fatal(err)
		}

		if string(marshaled) != joined {
			expectedGot(t, joined, string(marshaled))
		}

		var unmarshaled domain.RankingCollection
		err = json.Unmarshal(marshaled, &unmarshaled)

		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(rankings, unmarshaled) {
			expectedGot(t, rankings, unmarshaled)
		}
	})
}

func TestNewRankingCollectionFromRankingWithMetas(t *testing.T) {
	rwms := []domain.RankingWithMeta{
		generateRankingWithMeta("id_1", 1, 123),
		generateRankingWithMeta("id_2", 3, 456),
		generateRankingWithMeta("id_3", 2, 789),
	}

	rankingsFromRWM := domain.NewRankingCollectionFromRankingWithMetas(rwms)

	t.Run("creating a new ranking collection from RankingWithMetas must successfully retain the expected positions", func(t *testing.T) {
		expected := domain.RankingCollection{
			{
				ID:       "id_1",
				Position: 1,
			},
			{
				ID:       "id_2",
				Position: 3,
			},
			{
				ID:       "id_3",
				Position: 2,
			},
		}

		for i := 0; i < len(rwms); i++ {
			if !cmp.Equal(rankingsFromRWM[i], expected[i]) {
				t.Fatal(cmp.Diff(rankingsFromRWM[i], expected[i]))
			}
		}
	})
}

func TestNewRankingWithScoreCollectionFromIDs(t *testing.T) {
	var rankings = domain.NewRankingWithScoreCollectionFromIDs(rankingIDs)

	t.Run("creating a new ranking with score collection from ids must successfully populate the expected positions", func(t *testing.T) {
		for idx, ranking := range rankings {
			if ranking.ID != rankingIDs[idx] {
				expectedGot(t, rankingIDs[idx], ranking.ID)
			}
			if ranking.Position != idx+1 {
				expectedGot(t, idx+1, ranking.Position)
			}
			if ranking.Score != 0 {
				expectedGot(t, 0, ranking.Score)
			}
		}
	})

	t.Run("ranking with score collection must provide the expected total score", func(t *testing.T) {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

		var expectedTotal int
		for idx := range rankings {
			r := &rankings[idx]

			score := rnd.Intn(500)
			expectedTotal = expectedTotal + score

			r.Score = score
		}

		actualTotal := rankings.GetTotal()
		if actualTotal != expectedTotal {
			expectedGot(t, expectedTotal, actualTotal)
		}
	})
}

func TestRankingCollection_JSON(t *testing.T) {
	t.Run("creating a new ranking collection must successfully populate the expected positions", func(t *testing.T) {
		var (
			ids      = []string{"Pitman", "Wilson", "Hayter", "Pugh", "King"}
			rankings = domain.NewRankingCollectionFromIDs(ids)
		)

		for idx, ranking := range rankings {
			if ranking.ID != ids[idx] {
				expectedGot(t, ids[idx], ranking.ID)
			}
			if ranking.Position != idx+1 {
				expectedGot(t, idx+1, ranking.Position)
			}
		}
	})
}

func TestCalculateRankingScores(t *testing.T) {
	basePlanets := []string{"Mercury", "Venus", "Earth", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune"}
	baseRC := domain.NewRankingCollectionFromIDs(basePlanets)

	t.Run("calculate ranking scores of equivalent collections must produce the expected score", func(t *testing.T) {
		comparisonPlanets := []string{
			"Neptune", // should score 7
			"Uranus",  // should score 5
			"Saturn",  // should score 3
			"Venus",   // should score 2
			"Jupiter", // should score 0
			"Mars",    // should score 2
			"Mercury", // should score 6
			"Earth",   // should score 5
		}
		comparisonRC := domain.NewRankingCollectionFromIDs(comparisonPlanets)

		rws, err := domain.CalculateRankingsScores(baseRC, comparisonRC)
		if err != nil {
			t.Fatal(err)
		}

		// 7 + 5 + 3 + 2 + 0 + 2 + 6 + 5
		expectedScore := 30
		actualScore := rws.GetTotal()

		if expectedScore != actualScore {
			expectedGot(t, expectedScore, actualScore)
		}
	})

	t.Run("calculate ranking scores of non-equivalent collections must fail", func(t *testing.T) {
		// mismatched length
		comparisonPlanets := append(basePlanets, "extra element")
		comparisonRCMismatchedLength := domain.NewRankingCollectionFromIDs(comparisonPlanets)

		expectedErr := fmt.Errorf("mismatched baseIDs length: base %d, comparison %d", len(basePlanets), len(comparisonPlanets))
		_, err := domain.CalculateRankingsScores(baseRC, comparisonRCMismatchedLength)
		if expectedErr.Error() != err.Error() {
			expectedGot(t, expectedErr.Error(), err.Error())
		}

		// comparison collection has an id that base collection does not
		discrepantID := "Not A Planet"
		comparisonPlanets = basePlanets[:len(basePlanets)-1]        // minus last base element
		comparisonPlanets = append(comparisonPlanets, discrepantID) // add discrepant id to comparison
		comparisonRCWithDiscrepantID := domain.NewRankingCollectionFromIDs(comparisonPlanets)

		expectedErr = fmt.Errorf("base collection does not have id: '%s'", discrepantID)
		_, err = domain.CalculateRankingsScores(baseRC, comparisonRCWithDiscrepantID)
		if expectedErr.Error() != err.Error() {
			expectedGot(t, expectedErr.Error(), err.Error())
		}

		// comparison collection has a duplicate id that base collection does not
		duplicateID := basePlanets[0]
		comparisonPlanets = basePlanets
		comparisonPlanets[len(comparisonPlanets)-1] = duplicateID // replace last slice element with duplicate id
		comparisonRCWithDuplicateID := domain.NewRankingCollectionFromIDs(comparisonPlanets)

		expectedErr = fmt.Errorf("mismatched counts: id '%s' base collection count = 1, collection count = 2", duplicateID)
		_, err = domain.CalculateRankingsScores(baseRC, comparisonRCWithDuplicateID)
		if expectedErr.Error() != err.Error() {
			expectedGot(t, expectedErr.Error(), err.Error())
		}
	})
}

func generateRankingWithMeta(id string, pos int, metaVal int) domain.RankingWithMeta {
	var rwm = domain.NewRankingWithMeta()

	rwm.ID = id
	rwm.Position = pos
	rwm.MetaData["hello"] = metaVal

	return rwm
}
