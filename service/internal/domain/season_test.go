package domain_test

import (
	"fmt"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestSeason_CheckValidation(t *testing.T) {
	t.Run("validate seasons", func(t *testing.T) {
		for id, season := range domain.Seasons() {
			if id != season.ID {
				t.Fatal(fmt.Errorf("mismatched season id: %s != %s", id, season.ID))
			}

			if err := domain.ValidateSeason(season); err != nil {
				t.Fatal(fmt.Errorf("invalid season id: %s %+v", id, err))
			}
		}
	})
}

func TestSeason_GetStatus(t *testing.T) {
	now := time.Now()
	day := 24 * time.Hour

	season := domain.Season{
		EntriesFrom: now.Add(-7 * day), // 7 days ago
		StartDate:   now.Add(-5 * day), // 5 days ago
		EndDate:     now.Add(-3 * day), // 3 days ago
	}

	t.Run("on a date prior to entriesfrom, season status must be forthcoming", func(t *testing.T) {
		ts := now.Add(-8 * day) // 8 days ago
		status := season.GetStatus(ts)
		if status != domain.SeasonStatusForthcoming {
			expectedGot(t, domain.SeasonStatusForthcoming, status)
		}
	})

	t.Run("on a date between entriesfrom and startdate, season status must be accepting entries", func(t *testing.T) {
		ts := now.Add(-6 * day) // 6 days ago
		status := season.GetStatus(ts)
		if status != domain.SeasonStatusAcceptingEntries {
			expectedGot(t, domain.SeasonStatusAcceptingEntries, status)
		}
	})

	t.Run("on a date between startdate and enddate, season status must be active", func(t *testing.T) {
		ts := now.Add(-4 * day) // 4 days ago
		status := season.GetStatus(ts)
		if status != domain.SeasonStatusActive {
			expectedGot(t, domain.SeasonStatusActive, status)
		}
	})

	t.Run("on a date after enddate, season status must be elapsed", func(t *testing.T) {
		ts := now.Add(-2 * day) // 2 days ago
		status := season.GetStatus(ts)
		if status != domain.SeasonStatusElapsed {
			expectedGot(t, domain.SeasonStatusElapsed, status)
		}
	})
}

func TestSeasonCollection_GetByID(t *testing.T) {
	collection := domain.SeasonCollection{
		"season_1": domain.Season{ID: "season_1"},
		"season_2": domain.Season{ID: "season_2"},
		"season_3": domain.Season{ID: "season_3"},
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
