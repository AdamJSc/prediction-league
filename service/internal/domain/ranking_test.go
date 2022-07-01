package domain_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"prediction-league/service/internal/domain"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

func TestGetChangedRankingIDs(t *testing.T) {
	pughID, fletchID, pitmanID := "marc", "steve", "brett"

	tt := []struct {
		name    string
		x       domain.RankingCollection
		y       domain.RankingCollection
		wantIDs []string
	}{
		{
			name: "identical collections, no ids",
			x: domain.RankingCollection{
				{pughID, 1},
				{fletchID, 2},
				{pitmanID, 3},
			},
			y: domain.RankingCollection{
				{pughID, 1},
				{fletchID, 2},
				{pitmanID, 3},
			},
			wantIDs: make([]string, 0),
		},
		{
			name: "identical collection in alternative order, no ids",
			x: domain.RankingCollection{
				{pughID, 1},
				{fletchID, 2},
				{pitmanID, 3},
			},
			y: domain.RankingCollection{
				{pitmanID, 3},
				{fletchID, 2},
				{pughID, 1},
			},
			wantIDs: make([]string, 0),
		},
		{
			name: "two items swap one position, expect two ids",
			x: domain.RankingCollection{
				{pughID, 1},
				{fletchID, 2},
				{pitmanID, 3},
			},
			y: domain.RankingCollection{
				{pughID, 1},
				{pitmanID, 2}, // formerly position 3
				{fletchID, 3}, // formerly position 2
			},
			wantIDs: []string{fletchID, pitmanID},
		},
		{
			name: "identical ids but one is one position different, expect one id",
			x: domain.RankingCollection{
				{pughID, 1},
				{fletchID, 2},
				{pitmanID, 3},
			},
			y: domain.RankingCollection{
				{pughID, 1},
				{fletchID, 2},
				{pitmanID, 4}, // formerly position 3
			},
			wantIDs: []string{pitmanID},
		},
	}

	for _, tc := range tt {
		gotIDs := domain.GetChangedRankingIDs(tc.x, tc.y)
		if diff := cmp.Diff(tc.wantIDs, gotIDs); diff != "" {
			t.Fatalf("want ids %+v, got %+v, diff: %s", tc.wantIDs, gotIDs, diff)
		}
	}
}
