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
	var rankings = models.NewRankingCollection(ids)

	t.Run("creating a new ranking collection from ids only must successfully populate the expected positions", func(t *testing.T) {
		for idx, ranking := range rankings {
			if ranking.ID != ids[idx] {
				expectedGot(t, ids[idx], ranking.ID)
			}
			if ranking.Position != idx+1 {
				expectedGot(t, idx+1, ranking.Position)
			}
		}
	})

	t.Run("marshaling a ranking collection should retain the ids only", func(t *testing.T) {
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
			rankings = models.NewRankingCollection(ids)
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
