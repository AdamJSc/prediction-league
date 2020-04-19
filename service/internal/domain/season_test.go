package domain_test

import (
	"fmt"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestSeasons_ValidateFormat(t *testing.T) {
	for id, season := range domain.Seasons() {
		if id != season.ID {
			t.Fatal(fmt.Errorf("mismatched season id: %s != %s", id, season.ID))
		}

		if err := domain.ValidateSeason(season); err != nil {
			t.Fatal(fmt.Errorf("invalid season id: %s %+v", id, err))
		}
	}
}
