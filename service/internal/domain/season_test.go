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

	predictionsAcceptedTimeframes := []models.TimeFrame{
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
		Active:              activeTimeframe,
		EntriesAccepted:     entriesAcceptedTimeframe,
		PredictionsAccepted: predictionsAcceptedTimeframes,
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

	t.Run("on a date prior to first predictions accepted from date, is_accepting_predictions must be false and next_predictions_window must be first timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[0].From.Add(-day)
		state := season.GetState(ts)
		if state.IsAcceptingPredictions {
			t.Fatalf("expected season to not be accepting predictions, but it was, state: %+v", state)
		}
		if !cmp.Equal(*state.NextPredictionsWindow, predictionsAcceptedTimeframes[0]) {
			t.Fatal(cmp.Diff(*state.NextPredictionsWindow, predictionsAcceptedTimeframes[0]))
		}
	})

	t.Run("on first predictions accepted from date, is_accepting_predictions must be true and next_predictions_window must be first timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[0].From
		state := season.GetState(ts)
		if !state.IsAcceptingPredictions {
			t.Fatalf("expected season to be accepting predictions, but it wasn't, state: %+v", state)
		}
		if !state.NextPredictionsWindow.From.Equal(predictionsAcceptedTimeframes[0].From) {
			expectedGot(t, predictionsAcceptedTimeframes[0].From, state.NextPredictionsWindow.From)
		}
		if !state.NextPredictionsWindow.Until.Equal(predictionsAcceptedTimeframes[0].Until) {
			expectedGot(t, predictionsAcceptedTimeframes[0].Until, state.NextPredictionsWindow.Until)
		}
	})

	t.Run("on a date between first predictions accepted from date and first predictions accepted until date, is_accepting_predictions must be true and next_predictions_window must be first timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[0].From.Add(day)
		state := season.GetState(ts)
		if !state.IsAcceptingPredictions {
			t.Fatalf("expected season to be accepting predictions, but it wasn't, state: %+v", state)
		}
		if !state.NextPredictionsWindow.From.Equal(predictionsAcceptedTimeframes[0].From) {
			expectedGot(t, predictionsAcceptedTimeframes[0].From, state.NextPredictionsWindow.From)
		}
		if !state.NextPredictionsWindow.Until.Equal(predictionsAcceptedTimeframes[0].Until) {
			expectedGot(t, predictionsAcceptedTimeframes[0].Until, state.NextPredictionsWindow.Until)
		}
	})

	t.Run("on first predictions accepted until date, is_accepting_predictions must be false and next_predictions_window must be second timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[0].Until
		state := season.GetState(ts)
		if state.IsAcceptingPredictions {
			t.Fatalf("expected season to not be accepting predictions, but it was, state: %+v", state)
		}
		if !cmp.Equal(*state.NextPredictionsWindow, predictionsAcceptedTimeframes[1]) {
			t.Fatal(cmp.Diff(*state.NextPredictionsWindow, predictionsAcceptedTimeframes[1]))
		}
	})

	t.Run("on a date between first predictions accepted until date and second predictions accepted from date, is_accepting_predictions must be false and next_predictions_window must be second timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[1].From.Add(-day)
		state := season.GetState(ts)
		if state.IsAcceptingPredictions {
			t.Fatalf("expected season to not be accepting predictions, but it was, state: %+v", state)
		}
		if !cmp.Equal(*state.NextPredictionsWindow, predictionsAcceptedTimeframes[1]) {
			t.Fatal(cmp.Diff(*state.NextPredictionsWindow, predictionsAcceptedTimeframes[1]))
		}
	})

	t.Run("on second predictions accepted from date, is_accepting_predictions must be true and next_predictions_window must be second timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[1].From
		state := season.GetState(ts)
		if !state.IsAcceptingPredictions {
			t.Fatalf("expected season to be accepting predictions, but it wasn't, state: %+v", state)
		}
		if !state.NextPredictionsWindow.From.Equal(predictionsAcceptedTimeframes[1].From) {
			expectedGot(t, predictionsAcceptedTimeframes[1].From, state.NextPredictionsWindow.From)
		}
		if !state.NextPredictionsWindow.Until.Equal(predictionsAcceptedTimeframes[1].Until) {
			expectedGot(t, predictionsAcceptedTimeframes[1].Until, state.NextPredictionsWindow.Until)
		}
	})

	t.Run("on a date between second predictions accepted from date and second predictions accepted until date, is_accepting_predictions must be true and predictions_next_accepted must be second timeframe", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[1].From.Add(day)
		state := season.GetState(ts)
		if !state.IsAcceptingPredictions {
			t.Fatalf("expected season to be accepting predictions, but it wasn't, state: %+v", state)
		}
		if !state.NextPredictionsWindow.From.Equal(predictionsAcceptedTimeframes[1].From) {
			expectedGot(t, predictionsAcceptedTimeframes[1].From, state.NextPredictionsWindow.From)
		}
		if !state.NextPredictionsWindow.Until.Equal(predictionsAcceptedTimeframes[1].Until) {
			expectedGot(t, predictionsAcceptedTimeframes[1].Until, state.NextPredictionsWindow.Until)
		}
	})

	t.Run("on second predictions accepted until date, is_accepting_predictions must be false and predictions_next_accepted must be empty", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[1].Until
		state := season.GetState(ts)
		if state.IsAcceptingPredictions {
			t.Fatalf("expected season to not be accepting predictions, but it was, state: %+v", state)
		}
		if state.NextPredictionsWindow != nil {
			expectedGot(t, nil, state.NextPredictionsWindow)
		}
	})

	t.Run("on a date after second predictions accepted until date, is_accepting_predictions must be false and predictions_next_accepted must be empty", func(t *testing.T) {
		ts := predictionsAcceptedTimeframes[1].Until.Add(day)
		state := season.GetState(ts)
		if state.IsAcceptingPredictions {
			t.Fatalf("expected season to not be accepting predictions, but it was, state: %+v", state)
		}
		if state.NextPredictionsWindow != nil {
			expectedGot(t, nil, state.NextPredictionsWindow)
		}
	})
}
