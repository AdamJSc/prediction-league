package domain_test

import (
	"context"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestSeasonAgent_CreateSeason(t *testing.T) {
	defer truncate(t)

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	agent := domain.SeasonAgent{SeasonAgentInjector: injector{db: db}}

	t.Run("creating a valid season must succeed", func(t *testing.T) {
		name := "My Season"
		entriesFrom := time.Date(1992, 7, 1, 0, 0, 0, 0, loc)
		startDate := time.Date(1992, 8, 15, 15, 0, 0, 0, loc)
		endDate := time.Date(1993, 5, 11, 23, 59, 59, 0, loc)

		s := domain.Season{
			Name:        name,
			EntriesFrom: entriesFrom,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		// should succeed
		if err := agent.CreateSeason(context.Background(), &s, 0); err != nil {
			t.Fatal(err)
		}

		// check raw values that shouldn't have changed
		if !cmp.Equal(name, s.Name)().Success() {
			t.Fatalf("expected '%s', got '%s'", name, s.Name)
		}
		if !cmp.Equal(entriesFrom, s.EntriesFrom)().Success() {
			t.Fatalf("expected %+v, got %+v", entriesFrom, s.EntriesFrom)
		}
		if !cmp.Equal(startDate, s.StartDate)().Success() {
			t.Fatalf("expected %+v, got %+v", startDate, s.StartDate)
		}
		if !cmp.Equal(endDate, s.EndDate)().Success() {
			t.Fatalf("expected %+v, got %+v", endDate, s.EndDate)
		}

		// check sanitised values
		expectedID := "199293_1"
		expectedEntriesUntil := sqltypes.ToNullTime(s.StartDate)

		if !cmp.Equal(expectedID, s.ID)().Success() {
			t.Fatalf("expected '%s', got '%s'", expectedID, s.ID)
		}
		if !cmp.Equal(expectedEntriesUntil, s.EntriesUntil)().Success() {
			t.Fatalf("expected %+v, got %+v", expectedEntriesUntil, s.EntriesUntil)
		}
		if cmp.Equal(time.Time{}, s.CreatedAt)().Success() {
			t.Fatal("expected non-empty time, but got an empty one")
		}
		if !cmp.Equal(sqltypes.NullTime{}, s.UpdatedAt)().Success() {
			t.Fatalf("expected empty nulltime, got %+v", s.UpdatedAt)
		}
	})
}
