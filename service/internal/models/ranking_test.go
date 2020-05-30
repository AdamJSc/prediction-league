package models_test

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/models"
	"strings"
	"testing"
)

var rankingIDs = []string{"Pitman", "Wilson", "Hayter", "Pugh", "King"}

func TestRankingCollection_GetIDs(t *testing.T) {
	var rc models.RankingCollection

	for idx, id := range rankingIDs {
		rc = append(rc, models.Ranking{ID: id, Position: idx + 1})
	}

	t.Run("getting ids from ranking collection must match original slice of ids", func(t *testing.T) {
		ids := rc.GetIDs()
		if !cmp.Equal(ids, rankingIDs) {
			t.Fatal(cmp.Diff(ids, rankingIDs))
		}
	})
}

func TestRankingCollection_GetByID(t *testing.T) {
	var rc models.RankingCollection

	for idx, id := range rankingIDs {
		rc = append(rc, models.Ranking{ID: id, Position: idx + 1})
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
		expectedError := fmt.Errorf("not found ranking with id: %s", idToFind)

		_, err := rc.GetByID(idToFind)
		if !cmp.Equal(expectedError.Error(), err.Error()) {
			expectedGot(t, expectedError.Error(), err.Error())
		}
	})
}

func TestNewRankingCollectionFromIDs(t *testing.T) {
	var rankings = models.NewRankingCollectionFromIDs(rankingIDs)

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

		var unmarshaled models.RankingCollection
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
	rwms := []models.RankingWithMeta{
		generateRankingWithMeta("id_1", 1, 123),
		generateRankingWithMeta("id_2", 3, 456),
		generateRankingWithMeta("id_3", 2, 789),
	}

	rankingsFromRWM := models.NewRankingCollectionFromRankingWithMetas(rwms)

	t.Run("creating a new ranking collection from RankingWithMetas must successfully retain the expected positions", func(t *testing.T) {
		expected := models.RankingCollection{
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
	var rankings = models.NewRankingWithScoreCollectionFromIDs(rankingIDs)

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
}

func TestRankingCollection_JSON(t *testing.T) {
	t.Run("creating a new ranking collection must successfully populate the expected positions", func(t *testing.T) {
		var (
			ids      = []string{"Pitman", "Wilson", "Hayter", "Pugh", "King"}
			rankings = models.NewRankingCollectionFromIDs(ids)
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

func generateRankingWithMeta(id string, pos int, metaVal int) models.RankingWithMeta {
	var rwm = models.NewRankingWithMeta()

	rwm.ID = id
	rwm.Position = pos
	rwm.MetaData["hello"] = metaVal

	return rwm
}
