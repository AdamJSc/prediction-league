package domain_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"testing"
	"time"
)

func TestSeason_CheckValidation(t *testing.T) {
	t.Run("validate seasons", func(t *testing.T) {
		for id, season := range datastore.Seasons {
			if id != season.ID {
				t.Fatal(fmt.Errorf("mismatched season id: %s != %s", id, season.ID))
			}

			if err := domain.ValidateSeason(season); err != nil {
				t.Fatal(fmt.Errorf("invalid season id %s: %s", id, err.Error()))
			}
		}
	})
}

func TestSeason_GetState(t *testing.T) {
	now := time.Now()
	day := 24 * time.Hour

	activeTimeframe := models.TimeFrame{
		From:  now.Add(-7 * day), // 7 days ago
		Until: now.Add(-2 * day), // 2 days ago
	}

	entriesAcceptedTimeframe := models.TimeFrame{
		From:  now.Add(-9 * day), // 9 days ago
		Until: now.Add(-7 * day), // 7 days ago
	}

	selectionsAcceptedTimeframes := []models.TimeFrame{
		{
			From:  now.Add(-9 * day), // 9 days ago
			Until: now.Add(-7 * day), // 7 days ago
		},
		{
			From:  now.Add(-5 * day), // 5 days ago
			Until: now.Add(-3 * day), // 3 days ago
		},
	}

	season := models.Season{
		Active:             activeTimeframe,
		EntriesAccepted:    entriesAcceptedTimeframe,
		SelectionsAccepted: selectionsAcceptedTimeframes,
	}

	t.Run("on a date prior to active from, season status must be pending", func(t *testing.T) {
		ts := activeTimeframe.From.Add(-day)
		state := season.GetState(ts)
		if state.Status != models.SeasonStatusPending {
			expectedGot(t, models.SeasonStatusPending, state.Status)
		}
	})

	t.Run("on active from date, season status must be active", func(t *testing.T) {
		ts := activeTimeframe.From
		state := season.GetState(ts)
		if state.Status != models.SeasonStatusActive {
			expectedGot(t, models.SeasonStatusActive, state.Status)
		}
	})

	t.Run("on a date between active from date and active until date, season status must be active", func(t *testing.T) {
		ts := activeTimeframe.From.Add(day)
		state := season.GetState(ts)
		if state.Status != models.SeasonStatusActive {
			expectedGot(t, models.SeasonStatusActive, state.Status)
		}
	})

	t.Run("on active until date, season status must be elapsed", func(t *testing.T) {
		ts := activeTimeframe.Until
		state := season.GetState(ts)
		if state.Status != models.SeasonStatusElapsed {
			expectedGot(t, models.SeasonStatusElapsed, state.Status)
		}
	})

	t.Run("on a date after active until date, season status must be elapsed", func(t *testing.T) {
		ts := activeTimeframe.Until.Add(day)
		state := season.GetState(ts)
		if state.Status != models.SeasonStatusElapsed {
			expectedGot(t, models.SeasonStatusElapsed, state.Status)
		}
	})

	t.Run("on a date prior to entries accepted from date, is_accepting_entries must be false", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.From.Add(-day)
		state := season.GetState(ts)
		if state.IsAcceptingEntries {
			t.Fatalf("expected season to not be accepting entries, but it was, state: %+v", state)
		}
	})

	t.Run("on entries accepted from date, is_accepting_entries must be true", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.From
		state := season.GetState(ts)
		if !state.IsAcceptingEntries {
			t.Fatalf("expected season to be accepting entries, but it wasn't, state: %+v", state)
		}
	})

	t.Run("on a date between entries accepted from date and entries accepted until date, is_accepting_entries must be true", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.From.Add(day)
		state := season.GetState(ts)
		if !state.IsAcceptingEntries {
			t.Fatalf("expected season to be accepting entries, but it wasn't, state: %+v", state)
		}
	})

	t.Run("on entries accepted until date, is_accepting_entries must be false", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.Until
		state := season.GetState(ts)
		if state.IsAcceptingEntries {
			t.Fatalf("expected season to not be accepting entries, but it was, state: %+v", state)
		}
	})

	t.Run("on a date after entries accepted until date, is_accepting_entries must be false", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.Until.Add(day)
		state := season.GetState(ts)
		if state.IsAcceptingEntries {
			t.Fatalf("expected season to not be accepting entries, but it was, state: %+v", state)
		}
	})

	t.Run("on a date prior to first selections accepted from date, is_accepting_selections must be false and selections_next_accepted must be first timeframe", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[0].From.Add(-day)
		state := season.GetState(ts)
		if state.IsAcceptingSelections {
			t.Fatalf("expected season to not be accepting selections, but it was, state: %+v", state)
		}
		if !cmp.Equal(*state.SelectionsNextAccepted, selectionsAcceptedTimeframes[0]) {
			t.Fatal(cmp.Diff(*state.SelectionsNextAccepted, selectionsAcceptedTimeframes[0]))
		}
	})

	t.Run("on first selections accepted from date, is_accepting_selections must be true and selections_next_accepted must be empty", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[0].From
		state := season.GetState(ts)
		if !state.IsAcceptingSelections {
			t.Fatalf("expected season to be accepting selections, but it wasn't, state: %+v", state)
		}
		if state.SelectionsNextAccepted != nil {
			expectedGot(t, nil, state.SelectionsNextAccepted)
		}
	})

	t.Run("on a date between first selections accepted from date and first selections accepted until date, is_accepting_selections must be true and selections_next_accepted must be empty", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[0].From.Add(day)
		state := season.GetState(ts)
		if !state.IsAcceptingSelections {
			t.Fatalf("expected season to be accepting selections, but it wasn't, state: %+v", state)
		}
		if state.SelectionsNextAccepted != nil {
			expectedGot(t, nil, state.SelectionsNextAccepted)
		}
	})

	t.Run("on first selections accepted until date, is_accepting_selections must be false and selections_next_accepted must be second timeframe", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[0].Until
		state := season.GetState(ts)
		if state.IsAcceptingSelections {
			t.Fatalf("expected season to not be accepting selections, but it was, state: %+v", state)
		}
		if !cmp.Equal(*state.SelectionsNextAccepted, selectionsAcceptedTimeframes[1]) {
			t.Fatal(cmp.Diff(*state.SelectionsNextAccepted, selectionsAcceptedTimeframes[1]))
		}
	})

	t.Run("on a date between first selections accepted until date and second selections accepted from date, is_accepting_selections must be false and selections_next_accepted must be second timeframe", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[1].From.Add(-day)
		state := season.GetState(ts)
		if state.IsAcceptingSelections {
			t.Fatalf("expected season to not be accepting selections, but it was, state: %+v", state)
		}
		if !cmp.Equal(*state.SelectionsNextAccepted, selectionsAcceptedTimeframes[1]) {
			t.Fatal(cmp.Diff(*state.SelectionsNextAccepted, selectionsAcceptedTimeframes[1]))
		}
	})

	t.Run("on second selections accepted from date, is_accepting_selections must be true and selections_next_accepted must be empty", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[1].From
		state := season.GetState(ts)
		if !state.IsAcceptingSelections {
			t.Fatalf("expected season to be accepting selections, but it wasn't, state: %+v", state)
		}
		if state.SelectionsNextAccepted != nil {
			expectedGot(t, nil, state.SelectionsNextAccepted)
		}
	})

	t.Run("on a date between second selections accepted from date and second selections accepted until date, is_accepting_selections must be true and selections_next_accepted must be empty", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[1].From.Add(day)
		state := season.GetState(ts)
		if !state.IsAcceptingSelections {
			t.Fatalf("expected season to be accepting selections, but it wasn't, state: %+v", state)
		}
		if state.SelectionsNextAccepted != nil {
			expectedGot(t, nil, state.SelectionsNextAccepted)
		}
	})

	t.Run("on second selections accepted until date, is_accepting_selections must be false and selections_next_accepted must be empty", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[1].Until
		state := season.GetState(ts)
		if state.IsAcceptingSelections {
			t.Fatalf("expected season to not be accepting selections, but it was, state: %+v", state)
		}
		if state.SelectionsNextAccepted != nil {
			expectedGot(t, nil, state.SelectionsNextAccepted)
		}
	})

	t.Run("on a date after second selections accepted until date, is_accepting_selections must be false and selections_next_accepted must be empty", func(t *testing.T) {
		ts := selectionsAcceptedTimeframes[1].Until.Add(day)
		state := season.GetState(ts)
		if state.IsAcceptingSelections {
			t.Fatalf("expected season to not be accepting selections, but it was, state: %+v", state)
		}
		if state.SelectionsNextAccepted != nil {
			expectedGot(t, nil, state.SelectionsNextAccepted)
		}
	})
}
