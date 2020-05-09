package domain_test

import (
	"fmt"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
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

	season := models.Season{
		EntriesFrom: now.Add(-7 * day), // 7 days ago
		StartDate:   now.Add(-5 * day), // 5 days ago
		EndDate:     now.Add(-3 * day), // 3 days ago
	}

	t.Run("on a date prior to entriesfrom, season status must be forthcoming", func(t *testing.T) {
		ts := now.Add(-8 * day) // 8 days ago
		status := season.GetStatus(ts)
		if status != models.SeasonStatusForthcoming {
			expectedGot(t, models.SeasonStatusForthcoming, status)
		}
	})

	t.Run("on a date between entriesfrom and startdate, season status must be accepting entries", func(t *testing.T) {
		ts := now.Add(-6 * day) // 6 days ago
		status := season.GetStatus(ts)
		if status != models.SeasonStatusAcceptingEntries {
			expectedGot(t, models.SeasonStatusAcceptingEntries, status)
		}
	})

	t.Run("on a date between startdate and enddate, season status must be active", func(t *testing.T) {
		ts := now.Add(-4 * day) // 4 days ago
		status := season.GetStatus(ts)
		if status != models.SeasonStatusActive {
			expectedGot(t, models.SeasonStatusActive, status)
		}
	})

	t.Run("on a date after enddate, season status must be elapsed", func(t *testing.T) {
		ts := now.Add(-2 * day) // 2 days ago
		status := season.GetStatus(ts)
		if status != models.SeasonStatusElapsed {
			expectedGot(t, models.SeasonStatusElapsed, status)
		}
	})
}
