package models_test

import (
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/models"
	"testing"
	"time"
)

// TODO - tests for Season.GetState

// TODO - tests for Season.IsCompletedByStandings

func TestSeason_GetPredictionWindowBeginsWithin(t *testing.T) {
	var now = time.Now()
	var twoNanosecondsAgo = now.Add(-2 * time.Nanosecond)
	var fourNanosecondsAgo = now.Add(-4 * time.Nanosecond)
	var sixNanosecondsAgo = now.Add(-6 * time.Nanosecond)

	window1 := models.TimeFrame{
		From:  sixNanosecondsAgo,
		Until: fourNanosecondsAgo,
	}

	window2 := models.TimeFrame{
		From:  twoNanosecondsAgo,
		Until: now,
	}

	season := models.Season{
		PredictionsAccepted: []models.TimeFrame{window1, window2},
	}

	t.Run("timeframe that ends before first prediction window begins must return error", func(t *testing.T) {
		if _, err := season.GetPredictionWindowBeginsWithin(models.TimeFrame{
			From:  window1.From.Add(-2 * time.Nanosecond),
			Until: window1.From.Add(-time.Nanosecond),
		}); err != models.ErrNoMatchingPredictionWindow {
			expectedGot(t, models.ErrNoMatchingPredictionWindow, err)
		}
	})

	t.Run("timeframes that begin within first prediction window must return first prediction window", func(t *testing.T) {
		testCases := []models.TimeFrame{
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

		expected := models.SequencedTimeFrame{
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

	t.Run("timeframes that begin within second prediction window must return second prediction window", func(t *testing.T) {
		testCases := []models.TimeFrame{
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

		expected := models.SequencedTimeFrame{
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
}

func TestSeasonCollection_GetByID(t *testing.T) {
	collection := models.SeasonCollection{
		"season_1": models.Season{ID: "season_1"},
		"season_2": models.Season{ID: "season_2"},
		"season_3": models.Season{ID: "season_3"},
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
