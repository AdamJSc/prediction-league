package domain_test

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestSeason_IsCompletedByStandings(t *testing.T) {
	s := domain.Season{ID: "season_id", MaxRounds: 5}

	tt := []struct {
		stnd domain.Standings
		want bool
	}{
		{
			stnd: domain.Standings{SeasonID: "alt_season_id"},
			want: false,
		},
		{
			stnd: domain.Standings{SeasonID: "season_id", Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 4}}, // one short of max rounds
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
			}},
			want: false,
		},
		{
			stnd: domain.Standings{SeasonID: "season_id", Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
			}},
			want: true,
		},
	}

	for _, tc := range tt {
		if s.IsCompletedByStandings(tc.stnd) != tc.want {
			t.Fatalf("want %t, got %t", tc.want, !tc.want)
		}
	}
}

func TestSeason_GetPredictionWindowBeginsWithin(t *testing.T) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	dt := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)

	predWindow := domain.TimeFrame{
		From:  dt.Add(-7 * time.Nanosecond), // 7 nanoseconds before test date
		Until: dt.Add(-5 * time.Nanosecond), // 5 nanoseconds before test date
	}

	season := domain.Season{
		PredictionsAccepted: predWindow,
	}

	t.Run("timeframe that prediction window begins within must return prediction window", func(t *testing.T) {
		tt := []struct {
			name string
			tf   domain.TimeFrame
		}{
			{
				name: "straddles predictions begin",
				tf: domain.TimeFrame{
					From:  predWindow.From.Add(-time.Nanosecond),
					Until: predWindow.From.Add(time.Nanosecond),
				},
			},
			{
				name: "start is same as predictions begin",
				tf: domain.TimeFrame{
					From:  predWindow.From,
					Until: predWindow.From.Add(time.Nanosecond),
				},
			},
			{
				name: "end is same as predictions begin",
				tf: domain.TimeFrame{
					From:  predWindow.From.Add(-time.Nanosecond),
					Until: predWindow.From,
				},
			},
		}

		wantSeqTF := domain.SequencedTimeFrame{
			Count:   1,
			Total:   1,
			Current: &predWindow,
		}

		for _, tc := range tt {
			gotSeqTF, err := season.GetPredictionWindowBeginsWithin(tc.tf)
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(wantSeqTF, gotSeqTF)
			if diff != "" {
				t.Fatalf("tc '%s': want sequenced tf %+v, got %+v, diff: %s", tc.name, wantSeqTF, gotSeqTF, diff)
			}
		}
	})

	t.Run("timeframe that does not fall within prediction window must return the expected error", func(t *testing.T) {
		tt := []struct {
			name string
			tf   domain.TimeFrame
		}{
			{
				name: "ends before predictions begin",
				tf: domain.TimeFrame{
					From:  predWindow.From.Add(-2 * time.Nanosecond),
					Until: predWindow.From.Add(-time.Nanosecond),
				},
			},
			{
				name: "begins after predictions begin",
				tf: domain.TimeFrame{
					From:  predWindow.From.Add(time.Nanosecond),
					Until: predWindow.From.Add(2 * time.Nanosecond),
				},
			},
		}

		for _, tc := range tt {
			if _, gotErr := season.GetPredictionWindowBeginsWithin(tc.tf); !errors.Is(gotErr, domain.ErrNoMatchingPredictionWindow) {
				t.Fatalf("tc '%s': want no matching prediction window error, got %s (%T)", tc.name, gotErr, gotErr)
			}
		}
	})
}

func TestSeason_GetPredictionWindowEndsWithin(t *testing.T) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	dt := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)

	predWindow := domain.TimeFrame{
		From:  dt.Add(-7 * time.Nanosecond), // 7 nanoseconds before test date
		Until: dt.Add(-5 * time.Nanosecond), // 5 nanoseconds before test date
	}

	season := domain.Season{
		PredictionsAccepted: predWindow,
	}

	t.Run("timeframe that prediction window ends within must return prediction window", func(t *testing.T) {
		tt := []struct {
			name string
			tf   domain.TimeFrame
		}{
			{
				name: "straddles predictions end",
				tf: domain.TimeFrame{
					From:  predWindow.Until.Add(-time.Nanosecond),
					Until: predWindow.Until.Add(time.Nanosecond),
				},
			},
			{
				name: "start is same as predictions end",
				tf: domain.TimeFrame{
					From:  predWindow.Until,
					Until: predWindow.Until.Add(time.Nanosecond),
				},
			},
			{
				name: "end is same as predictions end",
				tf: domain.TimeFrame{
					From:  predWindow.Until.Add(-time.Nanosecond),
					Until: predWindow.Until,
				},
			},
		}

		wantSeqTF := domain.SequencedTimeFrame{
			Count:   1,
			Total:   1,
			Current: &predWindow,
		}

		for _, tc := range tt {
			gotSeqTF, err := season.GetPredictionWindowEndsWithin(tc.tf)
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(wantSeqTF, gotSeqTF)
			if diff != "" {
				t.Fatalf("tc '%s': want sequenced tf %+v, got %+v, diff: %s", tc.name, wantSeqTF, gotSeqTF, diff)
			}
		}
	})

	t.Run("timeframe that does not fall within prediction window must return the expected error", func(t *testing.T) {
		tt := []struct {
			name string
			tf   domain.TimeFrame
		}{
			{
				name: "ends before predictions end",
				tf: domain.TimeFrame{
					From:  predWindow.Until.Add(-2 * time.Nanosecond),
					Until: predWindow.Until.Add(-time.Nanosecond),
				},
			},
			{
				name: "begins after predictions end",
				tf: domain.TimeFrame{
					From:  predWindow.Until.Add(time.Nanosecond),
					Until: predWindow.Until.Add(2 * time.Nanosecond),
				},
			},
		}

		for _, tc := range tt {
			if _, gotErr := season.GetPredictionWindowEndsWithin(tc.tf); !errors.Is(gotErr, domain.ErrNoMatchingPredictionWindow) {
				t.Fatalf("tc '%s': want no matching prediction window error, got %s (%T)", tc.name, gotErr, gotErr)
			}
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

func TestValidateSeason(t *testing.T) {
	t.Run("failed season validation must return the expected error", func(t *testing.T) {
		tt := []struct {
			name    string
			s       domain.Season
			tc      domain.TeamCollection
			wantErr error
		}{
			{
				name: "fake season is skipped",
				s:    domain.Season{ID: "FakeSeason"},
			},
		}

		for _, tc := range tt {
			gotErr := domain.ValidateSeason(tc.s, tc.tc)
			switch {
			case tc.wantErr == nil && gotErr != nil:
				t.Fatalf("tc '%s': want no error, got %s (%T)", tc.name, gotErr, gotErr)
			case tc.wantErr != nil && gotErr == nil:
				t.Fatalf("tc '%s': want err %s (%T), got nil", tc.name, tc.wantErr, tc.wantErr)
			case tc.wantErr != nil && gotErr != nil && tc.wantErr.Error() != gotErr.Error():
				t.Fatalf("tc '%s': want err %s (%T), got err %s (%T)", tc.name, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
		}
	})
}

func TestSeason_CheckValidation(t *testing.T) {
	t.Run("run validation on seasons", func(t *testing.T) {
		seasons, err := domain.GetSeasonCollection()
		if err != nil {
			t.Fatal(err)
		}
		for id, season := range seasons {
			if id != season.ID {
				t.Fatal(fmt.Errorf("mismatched season id: %s != %s", id, season.ID))
			}

			if err := domain.ValidateSeason(season, domain.GetTeamCollection()); err != nil {
				t.Fatal(fmt.Errorf("invalid season: id %s: %s", id, err.Error()))
			}
		}
	})
}

func TestSeason_GetState(t *testing.T) {
	elapsingGracePeriod := 6 * time.Hour

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	dt := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)

	liveTimeframe := domain.TimeFrame{
		From:  dt.Add(-6 * time.Second),
		Until: dt.Add(-5 * time.Second),
	}

	entriesAcceptedTimeframe := domain.TimeFrame{
		From:  dt.Add(-4 * time.Second),
		Until: dt.Add(-3 * time.Second),
	}

	predictionsAcceptedTimeframe := domain.TimeFrame{
		From:  dt.Add(-2 * time.Second),
		Until: dt.Add(-time.Second),
	}

	season := domain.Season{
		Live:                liveTimeframe,
		EntriesAccepted:     entriesAcceptedTimeframe,
		PredictionsAccepted: predictionsAcceptedTimeframe,
	}

	t.Run("at a timestamp prior to live from, live status must be pending", func(t *testing.T) {
		ts := liveTimeframe.From.Add(-time.Nanosecond)
		state := season.GetState(ts)
		if state.LiveStatus != domain.SeasonStatePending {
			expectedGot(t, domain.SeasonStatePending, state.LiveStatus)
		}
	})

	t.Run("on live from date, live status must be active", func(t *testing.T) {
		ts := liveTimeframe.From
		state := season.GetState(ts)
		if state.LiveStatus != domain.SeasonStateActive {
			expectedGot(t, domain.SeasonStateActive, state.LiveStatus)
		}
	})

	t.Run("at a timestamp between live from date and live until date, live status must be active", func(t *testing.T) {
		ts := liveTimeframe.From.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.LiveStatus != domain.SeasonStateActive {
			expectedGot(t, domain.SeasonStateActive, state.LiveStatus)
		}
	})

	t.Run("on live until date, live status must be elapsed", func(t *testing.T) {
		ts := liveTimeframe.Until
		state := season.GetState(ts)
		if state.LiveStatus != domain.SeasonStateElapsed {
			expectedGot(t, domain.SeasonStateElapsed, state.LiveStatus)
		}
	})

	t.Run("at a timestamp after live until date, live status must be elapsed", func(t *testing.T) {
		ts := liveTimeframe.Until.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.LiveStatus != domain.SeasonStateElapsed {
			expectedGot(t, domain.SeasonStateElapsed, state.LiveStatus)
		}
	})

	t.Run("at a timestamp prior to entries accepted from, entries status must be pending", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.From.Add(-time.Nanosecond)
		state := season.GetState(ts)
		if state.EntriesStatus != domain.SeasonStatePending {
			expectedGot(t, domain.SeasonStatePending, state.EntriesStatus)
		}
	})

	t.Run("on entries accepted from date, entries status must be active", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.From
		state := season.GetState(ts)
		if state.EntriesStatus != domain.SeasonStateActive {
			expectedGot(t, domain.SeasonStateActive, state.EntriesStatus)
		}
	})

	t.Run("at a timestamp between entries accepted from date and entries accepted until date, entries status must be active", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.From.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.EntriesStatus != domain.SeasonStateActive {
			expectedGot(t, domain.SeasonStateActive, state.EntriesStatus)
		}
	})

	t.Run("on entries accepted until date, entries status must be elapsed", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.Until
		state := season.GetState(ts)
		if state.EntriesStatus != domain.SeasonStateElapsed {
			expectedGot(t, domain.SeasonStateElapsed, state.EntriesStatus)
		}
	})

	t.Run("at a timestamp after entries accepted until date, entries status must be elapsed", func(t *testing.T) {
		ts := entriesAcceptedTimeframe.Until.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.EntriesStatus != domain.SeasonStateElapsed {
			expectedGot(t, domain.SeasonStateElapsed, state.EntriesStatus)
		}
	})

	t.Run("at a timestamp prior to predictions accepted from, predictions status must be pending", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.From.Add(-time.Nanosecond)
		state := season.GetState(ts)
		if state.PredictionsStatus != domain.SeasonStatePending {
			expectedGot(t, domain.SeasonStatePending, state.PredictionsStatus)
		}
	})

	t.Run("on predictions accepted from date, predictions status must be active", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.From
		state := season.GetState(ts)
		if state.PredictionsStatus != domain.SeasonStateActive {
			expectedGot(t, domain.SeasonStateActive, state.PredictionsStatus)
		}
	})

	t.Run("at a timestamp between predictions accepted from date and predictions accepted until date, predictions status must be active", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.From.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.PredictionsStatus != domain.SeasonStateActive {
			expectedGot(t, domain.SeasonStateActive, state.PredictionsStatus)
		}
	})

	t.Run("on predictions accepted until date, predictions status must be elapsed", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.Until
		state := season.GetState(ts)
		if state.PredictionsStatus != domain.SeasonStateElapsed {
			expectedGot(t, domain.SeasonStateElapsed, state.PredictionsStatus)
		}
	})

	t.Run("at a timestamp after predictions accepted until date, predictions status must be elapsed", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.Until.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.PredictionsStatus != domain.SeasonStateElapsed {
			expectedGot(t, domain.SeasonStateElapsed, state.PredictionsStatus)
		}
	})

	t.Run("at a timestamp prior to predictions accepted until minus grace period, predictions closing must be false", func(t *testing.T) {
		offset := elapsingGracePeriod + time.Nanosecond
		ts := predictionsAcceptedTimeframe.Until.Add(-offset)
		state := season.GetState(ts)
		if state.PredictionsClosing {
			t.Fatal("want predictions closing false, but got true")
		}
	})

	t.Run("at a timestamp that is exactly predictions accepted until minus grace period, predictions closing must be true", func(t *testing.T) {
		offset := elapsingGracePeriod
		ts := predictionsAcceptedTimeframe.Until.Add(-offset)
		state := season.GetState(ts)
		if !state.PredictionsClosing {
			t.Fatal("want predictions closing true, but got false")
		}
	})

	t.Run("at a timestamp after predictions accepted until minus grace period, but before predictions accepted until, predictions closing must be true", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.Until.Add(-time.Nanosecond)
		state := season.GetState(ts)
		if !state.PredictionsClosing {
			t.Fatal("want predictions closing true, but got false")
		}
	})

	t.Run("on predictions accepted until date, predictions closing must be false", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.Until
		state := season.GetState(ts)
		if state.PredictionsClosing {
			t.Fatal("want predictions closing false, but got true")
		}
	})

	t.Run("at a timestamp after predictions accepted until date, predictions closing must be false", func(t *testing.T) {
		ts := predictionsAcceptedTimeframe.Until.Add(time.Nanosecond)
		state := season.GetState(ts)
		if state.PredictionsClosing {
			t.Fatal("want predictions closing false, but got true")
		}
	})
}
