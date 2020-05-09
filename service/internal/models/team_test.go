package models_test

import (
	"prediction-league/service/internal/models"
	"testing"
)

func TestTeamCollection_GetByID(t *testing.T) {
	collection := models.TeamCollection{
		"team_1": models.Team{ID: "team_1"},
		"team_2": models.Team{ID: "team_2"},
		"team_3": models.Team{ID: "team_3"},
	}

	t.Run("retrieving an existing team by id must succeed", func(t *testing.T) {
		id := "team_2"
		s, err := collection.GetByID(id)
		if err != nil {
			t.Fatal(err)
		}

		if s.ID != id {
			expectedGot(t, id, s.ID)
		}
	})

	t.Run("retrieving a non-existing team by id must fail", func(t *testing.T) {
		id := "not_existent_team_id"
		if _, err := collection.GetByID(id); err == nil {
			expectedNonEmpty(t, "team collection getbyid error")
		}
	})
}
