package domain_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

// TODO - tests for Season.GetState

// TODO - tests for Season.IsCompletedByStandings

func TestSeason_GetPredictionWindowBeginsWithin(t *testing.T) {
	var now = time.Now()
	var twoNanosecondsAgo = now.Add(-2 * time.Nanosecond)
	var fiveNanosecondsAgo = now.Add(-5 * time.Nanosecond)
	var sevenNanosecondsAgo = now.Add(-7 * time.Nanosecond)

	window1 := domain.TimeFrame{
		From:  sevenNanosecondsAgo,
		Until: fiveNanosecondsAgo,
	}

	window2 := domain.TimeFrame{
		From:  twoNanosecondsAgo,
		Until: now,
	}

	season := domain.Season{
		PredictionsAccepted: []domain.TimeFrame{window1, window2},
	}

	t.Run("timeframe that ends before first prediction window begins must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowBeginsWithin(domain.TimeFrame{
			From:  window1.From.Add(-2 * time.Nanosecond),
			Until: window1.From.Add(-time.Nanosecond),
		}); err != domain.ErrNoMatchingPredictionWindow {
			expectedGot(t, domain.ErrNoMatchingPredictionWindow, err)
		}
	})

	t.Run("timeframes that first prediction window begins within must return first prediction window", func(t *testing.T) {
		testCases := []domain.TimeFrame{
			// straddles window1 begin
			{
				From:  window1.From.Add(-time.Nanosecond),
				Until: window1.From.Add(time.Nanosecond),
			},
			// start matches window1 begin
			{
				From:  window1.From,
				Until: window1.From.Add(time.Nanosecond),
			},
			// end matches window1 begin
			{
				From:  window1.From.Add(-time.Nanosecond),
				Until: window1.From,
			},
		}

		expected := domain.SequencedTimeFrame{
			Count:   1,
			Total:   2,
			Current: &window1,
			Next:    &window2,
		}

		for _, tc := range testCases {
			actual, err := season.GetPredictionWindowBeginsWithin(tc)
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(expected, actual)
			if diff != "" {
				expectedGot(t, "empty diff", diff)
			}
		}
	})

	t.Run("timeframes that second prediction window begins within must return second prediction window", func(t *testing.T) {
		testCases := []domain.TimeFrame{
			// straddles window2 begin
			{
				From:  window2.From.Add(-time.Nanosecond),
				Until: window2.From.Add(time.Nanosecond),
			},
			// start matches window2 begin
			{
				From:  window2.From,
				Until: window2.From.Add(time.Nanosecond),
			},
			// end matches window2 begin
			{
				From:  window2.From.Add(-time.Nanosecond),
				Until: window2.From,
			},
		}

		expected := domain.SequencedTimeFrame{
			Count:   2,
			Total:   2,
			Current: &window2,
			Next:    nil,
		}

		for _, tc := range testCases {
			actual, err := season.GetPredictionWindowBeginsWithin(tc)
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(expected, actual)
			if diff != "" {
				expectedGot(t, "empty diff", diff)
			}
		}
	})

	t.Run("timeframe that begins and ends between either prediction window must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowBeginsWithin(domain.TimeFrame{
			From:  window1.Until.Add(time.Nanosecond),
			Until: window2.From.Add(-time.Nanosecond),
		}); err != domain.ErrNoMatchingPredictionWindow {
			expectedGot(t, domain.ErrNoMatchingPredictionWindow, err)
		}
	})

	t.Run("timeframe that begins after second prediction window begins must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowBeginsWithin(domain.TimeFrame{
			From:  window2.From.Add(time.Nanosecond),
			Until: window2.From.Add(2 * time.Nanosecond),
		}); err != domain.ErrNoMatchingPredictionWindow {
			expectedGot(t, domain.ErrNoMatchingPredictionWindow, err)
		}
	})
}

func TestSeason_GetPredictionWindowEndsWithin(t *testing.T) {
	var now = time.Now()
	var twoNanosecondsAgo = now.Add(-2 * time.Nanosecond)
	var fiveNanosecondsAgo = now.Add(-5 * time.Nanosecond)
	var sevenNanosecondsAgo = now.Add(-7 * time.Nanosecond)

	window1 := domain.TimeFrame{
		From:  sevenNanosecondsAgo,
		Until: fiveNanosecondsAgo,
	}

	window2 := domain.TimeFrame{
		From:  twoNanosecondsAgo,
		Until: now,
	}

	season := domain.Season{
		PredictionsAccepted: []domain.TimeFrame{window1, window2},
	}

	t.Run("timeframe that ends before first prediction window ends must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowEndsWithin(domain.TimeFrame{
			From:  window1.Until.Add(-2 * time.Nanosecond),
			Until: window1.Until.Add(-time.Nanosecond),
		}); err != domain.ErrNoMatchingPredictionWindow {
			expectedGot(t, domain.ErrNoMatchingPredictionWindow, err)
		}
	})

	t.Run("timeframes that first prediction window ends within must return first prediction window", func(t *testing.T) {
		testCases := []domain.TimeFrame{
			// straddles window1 end
			{
				From:  window1.Until.Add(-time.Nanosecond),
				Until: window1.Until.Add(time.Nanosecond),
			},
			// start matches window1 end
			{
				From:  window1.Until,
				Until: window1.Until.Add(time.Nanosecond),
			},
			// end matches window1 end
			{
				From:  window1.Until.Add(-time.Nanosecond),
				Until: window1.Until,
			},
		}

		expected := domain.SequencedTimeFrame{
			Count:   1,
			Total:   2,
			Current: &window1,
			Next:    &window2,
		}

		for _, tc := range testCases {
			actual, err := season.GetPredictionWindowEndsWithin(tc)
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(expected, actual)
			if diff != "" {
				expectedGot(t, "empty diff", diff)
			}
		}
	})

	t.Run("timeframes that second prediction window ends within must return second prediction window", func(t *testing.T) {
		testCases := []domain.TimeFrame{
			// straddles window2 end
			{
				From:  window2.Until.Add(-time.Nanosecond),
				Until: window2.Until.Add(time.Nanosecond),
			},
			// start matches window2 end
			{
				From:  window2.Until,
				Until: window2.Until.Add(time.Nanosecond),
			},
			// end matches window2 end
			{
				From:  window2.Until.Add(-time.Nanosecond),
				Until: window2.Until,
			},
		}

		expected := domain.SequencedTimeFrame{
			Count:   2,
			Total:   2,
			Current: &window2,
			Next:    nil,
		}

		for _, tc := range testCases {
			actual, err := season.GetPredictionWindowEndsWithin(tc)
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(expected, actual)
			if diff != "" {
				expectedGot(t, "empty diff", diff)
			}
		}
	})

	t.Run("timeframe that begins and ends between either prediction window must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowEndsWithin(domain.TimeFrame{
			From:  window1.Until.Add(time.Nanosecond),
			Until: window2.From.Add(-time.Nanosecond),
		}); err != domain.ErrNoMatchingPredictionWindow {
			expectedGot(t, domain.ErrNoMatchingPredictionWindow, err)
		}
	})

	t.Run("timeframe that begins after second prediction window bendsegins must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowEndsWithin(domain.TimeFrame{
			From:  window2.Until.Add(time.Nanosecond),
			Until: window2.Until.Add(2 * time.Nanosecond),
		}); err != domain.ErrNoMatchingPredictionWindow {
			expectedGot(t, domain.ErrNoMatchingPredictionWindow, err)
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

func TestSeason_CheckValidation(t *testing.T) {
	t.Run("validate seasons", func(t *testing.T) {
		for id, season := range domain.SeasonsDataStore {
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

	activeTimeframe := domain.TimeFrame{
		From:  now.Add(-7 * day), // 7 days ago
		Until: now.Add(-2 * day), // 2 days ago
	}

	entriesAcceptedTimeframe := domain.TimeFrame{
		From:  now.Add(-9 * day), // 9 days ago
		Until: now.Add(-7 * day), // 7 days ago
	}

	predictionsAcceptedTimeframes := []domain.TimeFrame{
		{
			From:  now.Add(-9 * day), // 9 days ago
			Until: now.Add(-7 * day), // 7 days ago
		},
		{
			From:  now.Add(-5 * day), // 5 days ago
			Until: now.Add(-3 * day), // 3 days ago
		},
	}

	season := domain.Season{
		Active:              activeTimeframe,
		EntriesAccepted:     entriesAcceptedTimeframe,
		PredictionsAccepted: predictionsAcceptedTimeframes,
	}

	t.Run("on a date prior to active from, season status must be pending", func(t *testing.T) {
		ts := activeTimeframe.From.Add(-day)
		state := season.GetState(ts)
		if state.Status != domain.SeasonStatusPending {
			expectedGot(t, domain.SeasonStatusPending, state.Status)
		}
	})

	t.Run("on active from date, season status must be active", func(t *testing.T) {
		ts := activeTimeframe.From
		state := season.GetState(ts)
		if state.Status != domain.SeasonStatusActive {
			expectedGot(t, domain.SeasonStatusActive, state.Status)
		}
	})

	t.Run("on a date between active from date and active until date, season status must be active", func(t *testing.T) {
		ts := activeTimeframe.From.Add(day)
		state := season.GetState(ts)
		if state.Status != domain.SeasonStatusActive {
			expectedGot(t, domain.SeasonStatusActive, state.Status)
		}
	})

	t.Run("on active until date, season status must be elapsed", func(t *testing.T) {
		ts := activeTimeframe.Until
		state := season.GetState(ts)
		if state.Status != domain.SeasonStatusElapsed {
			expectedGot(t, domain.SeasonStatusElapsed, state.Status)
		}
	})

	t.Run("on a date after active until date, season status must be elapsed", func(t *testing.T) {
		ts := activeTimeframe.Until.Add(day)
		state := season.GetState(ts)
		if state.Status != domain.SeasonStatusElapsed {
			expectedGot(t, domain.SeasonStatusElapsed, state.Status)
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
