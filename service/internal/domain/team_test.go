package domain_test

import (
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestTeamCollection_GetByID(t *testing.T) {
	collection := domain.TeamCollection{
		"team_1": domain.Team{ID: "team_1"},
		"team_2": domain.Team{ID: "team_2"},
		"team_3": domain.Team{ID: "team_3"},
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

func TestTeam_CheckValidation(t *testing.T) {
	t.Run("validate teams", func(t *testing.T) {
		for id, team := range domain.TeamsDataStore {
			if id != team.ID {
				t.Fatal(fmt.Errorf("mismatched team id: %s != %s", id, team.ID))
			}

			if err := domain.ValidateTeam(team); err != nil {
				t.Fatal(fmt.Errorf("invalid team id %s: %s", id, err.Error()))
			}
		}
	})
}

func TestTeam_FilterTeamsByIDs(t *testing.T) {
	collection := domain.TeamCollection{
		"ID123": domain.Team{ID: "ID123"},
		"ID456": domain.Team{ID: "ID456"},
		"ID789": domain.Team{ID: "ID789"},
	}

	t.Run("filter teams by valid ids must succeed", func(t *testing.T) {
		ids := []string{
			"ID123",
			"ID789",
		}

		expectedResult := []domain.Team{
			{ID: "ID123"},
			{ID: "ID789"},
		}

		actualCollection, err := domain.FilterTeamsByIDs(ids, collection)
		if err != nil {
			t.Fatal(err)
		}

		if len(expectedResult) != len(actualCollection) {
			expectedGot(t, len(expectedResult), len(actualCollection))
		}
		for idx := range expectedResult {
			if expectedResult[idx].ID != actualCollection[idx].ID {
				expectedGot(t, expectedResult[idx].ID, actualCollection[idx].ID)
			}
		}
	})

	t.Run("filter teams by an invalid id must fail", func(t *testing.T) {
		ids := []string{
			"ID123",
			"not_a_valid_team_id",
			"ID789",
		}

		expectedError := errors.New("missing team id: not_a_valid_team_id")
		if _, err := domain.FilterTeamsByIDs(ids, collection); err == nil || err.Error() != expectedError.Error() {
			expectedGot(t, expectedError, err)
		}
	})

	t.Run("filter teams by duplicate ids must fail", func(t *testing.T) {
		ids := []string{
			"ID123",
			"ID123",
			"ID789",
		}

		expectedError := errors.New("team id exists multiple times: ID123")
		if _, err := domain.FilterTeamsByIDs(ids, collection); err == nil || err.Error() != expectedError.Error() {
			expectedGot(t, expectedError, err)
		}
	})
}
