package models_test

import (
	"prediction-league/service/internal/models"
	"testing"
)

// TODO - tests for Season.GetState

// TODO - tests for Season.IsCompletedByStandings

func TestSeasonCollection_GetByID(t *testing.T) {
	collection := models.SeasonCollection{
		"season_1": models.Season{ID: "season_1"},
		"season_2": models.Season{ID: "season_2"},
		"season_3": models.Season{ID: "season_3"},
	}

	t.Run("retrieving an existing season by id must succeed", func(t *testing.T) {
		id := "season_2"
		s, err := collection.GetByID(id)
		if err != nil {
			t.Fatal(err)
		}

		if s.ID != id {
			expectedGot(t, id, s.ID)
		}
	})

	t.Run("retrieving a non-existing season by id must fail", func(t *testing.T) {
		id := "not_existent_season_id"
		if _, err := collection.GetByID(id); err == nil {
			expectedNonEmpty(t, "season collection getbyid error")
		}
	})
}
