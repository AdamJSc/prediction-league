package domain_test

import (
	"fmt"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestTeam_CheckValidation(t *testing.T) {
	t.Run("validate teams", func(t *testing.T) {
		for id, team := range datastore.Teams {
			if id != team.ID {
				t.Fatal(fmt.Errorf("mismatched team id: %s != %s", id, team.ID))
			}

			if err := domain.ValidateTeam(team); err != nil {
				t.Fatal(fmt.Errorf("invalid team id: %s %+v", id, err))
			}
		}
	})
}
