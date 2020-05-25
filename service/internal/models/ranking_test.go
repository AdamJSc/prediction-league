package models_test

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/models"
	"strings"
	"testing"
)

func TestRankingCollection(t *testing.T) {
	var ids = []string{"Pitman", "Wilson", "Hayter", "Pugh", "King"}

	t.Run("creating a new ranking collection from ids must successfully populate the expected positions", func(t *testing.T) {
		var rankings = models.NewRankingCollectionFromIDs(ids)

		for idx, ranking := range rankings {
			if ranking.ID != ids[idx] {
				expectedGot(t, ids[idx], ranking.ID)
			}
			if ranking.Position != idx+1 {
				expectedGot(t, idx+1, ranking.Position)
			}
		}
	})

	t.Run("creating a new ranking collection from RankingWithMetas must successfully retain the expected positions", func(t *testing.T) {
		rwms := []models.RankingWithMeta{
			generateRankingWithMeta("id_1", 1, 123),
			generateRankingWithMeta("id_2", 3, 456),
			generateRankingWithMeta("id_3", 2, 789),
		}

		rankingsFromRWM := models.NewRankingCollectionFromRankingWithMetas(rwms)

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

	t.Run("creating a new ranking with score collection from ids must successfully populate the expected positions", func(t *testing.T) {
		var rankings = models.NewRankingWithScoreCollectionFromIDs(ids)

		for idx, ranking := range rankings {
			if ranking.ID != ids[idx] {
				expectedGot(t, ids[idx], ranking.ID)
			}
			if ranking.Position != idx+1 {
				expectedGot(t, idx+1, ranking.Position)
			}
			if ranking.Score != 0 {
				expectedGot(t, 0, ranking.Score)
			}
		}
	})

	t.Run("marshaling and unmarshaling a ranking collection should retain the ids only", func(t *testing.T) {
		var rankings = models.NewRankingCollectionFromIDs(ids)

		var joined = fmt.Sprintf(`["%s"]`, strings.Join(ids, `","`))

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
