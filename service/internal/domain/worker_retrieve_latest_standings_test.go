package domain_test

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestNewRetrieveLatestStandingsWorker(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		tc := make(map[string]domain.Team)
		cl := &mockClock{}
		l := &mockLogger{}
		ea := &domain.EntryAgent{}
		sa := &domain.StandingsAgent{}
		sepa := &domain.ScoredEntryPredictionAgent{}
		ca := &domain.CommunicationsAgent{}
		fds := &domain.NoopFootballDataSource{}

		tt := []struct {
			tc      domain.TeamCollection
			cl      domain.Clock
			l       domain.Logger
			ea      *domain.EntryAgent
			sa      *domain.StandingsAgent
			sepa    *domain.ScoredEntryPredictionAgent
			ca      *domain.CommunicationsAgent
			fds     domain.FootballDataSource
			wantErr bool
		}{
			{nil, cl, l, ea, sa, sepa, ca, fds, true},
			{tc, nil, l, ea, sa, sepa, ca, fds, true},
			{tc, cl, nil, ea, sa, sepa, ca, fds, true},
			{tc, cl, l, nil, sa, sepa, ca, fds, true},
			{tc, cl, l, ea, nil, sepa, ca, fds, true},
			{tc, cl, l, ea, sa, nil, ca, fds, true},
			{tc, cl, l, ea, sa, sepa, nil, fds, true},
			{tc, cl, l, ea, sa, sepa, ca, nil, true},
			{tc, cl, l, ea, sa, sepa, ca, fds, false},
		}
		for idx, tc := range tt {
			w, gotErr := domain.NewRetrieveLatestStandingsWorker(domain.Season{}, tc.tc, tc.cl, tc.l, tc.ea, tc.sa, tc.sepa, tc.ca, tc.fds)
			if tc.wantErr && !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("tc #%d: want ErrIsNil, got %s (%T)", idx, gotErr, gotErr)
			}
			if !tc.wantErr && w == nil {
				t.Fatalf("tc #%d: want non-empty worker, got nil", idx)
			}
		}
	})
}

func TestRetrieveLatestStandingsWorker_ProcessExistingStandings(t *testing.T) {
	t.Run("happy path must produce the expected results", func(t *testing.T) {
		t.Cleanup(truncate)

		ctx := context.Background()

		now := time.Now().Truncate(time.Second)

		existStnd := domain.Standings{
			ID:          uuid.New(),
			SeasonID:    "season_id",
			RoundNumber: 123,
			Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 1234}},
			},
			CreatedAt: now,
		}

		if err := sr.Insert(ctx, &existStnd); err != nil {
			t.Fatal(err)
		}

		clientStnd := domain.Standings{Rankings: []domain.RankingWithMeta{
			{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5678}},
		}}

		wantStnd := existStnd
		wantStnd.Rankings = clientStnd.Rankings
		wantStnd.UpdatedAt = &now

		wantSelect := []domain.Standings{wantStnd}

		sa, err := domain.NewStandingsAgent(sr)
		if err != nil {
			t.Fatal(sa)
		}

		r := domain.NewTestRetrieveLatestStandingsWorker(domain.Season{}, nil, nil, nil, nil, sa, nil, nil, nil)

		gotStnd, err := r.ProcessExistingStandings(ctx, existStnd, clientStnd)
		if err != nil {
			t.Fatal(err)
		}

		diff := cmp.Diff(wantStnd, gotStnd)
		if diff != "" {
			t.Fatalf("want %+v, got %+v, diff: %s", wantStnd, gotStnd, diff)
		}

		gotSelect, err := sr.Select(ctx, map[string]interface{}{"id": existStnd.ID.String()}, false)
		if err != nil {
			t.Fatal(err)
		}

		diff = cmp.Diff(wantSelect, gotSelect)
		if diff != "" {
			t.Fatalf("want %+v, got %+v, diff: %s", wantSelect, gotSelect, cmp.Diff(wantSelect, gotSelect))
		}
	})
}

func TestRetrieveLatestStandingsWorker_ProcessNewStandings(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()
	now := time.Now().Truncate(time.Second)

	sa, err := domain.NewStandingsAgent(sr)
	if err != nil {
		t.Fatal(sa)
	}

	r := domain.NewTestRetrieveLatestStandingsWorker(domain.Season{}, nil, nil, nil, nil, sa, nil, nil, nil)

	id1 := uuid.New()
	id2 := uuid.New()

	tt := []struct {
		seedStnd   *domain.Standings
		stndToProc domain.Standings
		wantStnd   domain.Standings
	}{
		{
			stndToProc: domain.Standings{
				RoundNumber: 1, // process first round that does not yet exist (no seed), so expect this to be created straight away
				CreatedAt:   now,
			},
			wantStnd: domain.Standings{
				RoundNumber: 1, // expect process Standings to be created and returned
				CreatedAt:   now,
			},
		},
		{
			seedStnd: &domain.Standings{
				ID:          id1,
				RoundNumber: 2,
				CreatedAt:   now,
			},
			stndToProc: domain.Standings{
				ID:          id1,
				RoundNumber: 3, // attempt to process next round when previous round has not been finalised
				CreatedAt:   now,
			},
			wantStnd: domain.Standings{
				ID:          id1,
				RoundNumber: 2, // expected seeded Standings to be updated (finalised) and returned
				Finalised:   true,
				CreatedAt:   now,
				UpdatedAt:   &now,
			},
		},
		{
			seedStnd: &domain.Standings{
				ID:          id2,
				RoundNumber: 3,
				Finalised:   true,
				CreatedAt:   now,
			},
			stndToProc: domain.Standings{
				RoundNumber: 4, // attempt to process next round when previous round HAS been finalised
				CreatedAt:   now,
			},
			wantStnd: domain.Standings{
				RoundNumber: 4, // expect process Standings to be created and returned
				CreatedAt:   now,
			},
		},
	}

	for _, tc := range tt {
		if tc.seedStnd != nil {
			if err := sr.Insert(ctx, tc.seedStnd); err != nil {
				t.Fatal(err)
			}
		}

		gotStnd, err := r.ProcessNewStandings(ctx, tc.stndToProc)
		if err != nil {
			t.Fatal(err)
		}

		// if the wanted Standings object doesn't specify an ID, we're expecting it to be created by our test method
		// so populate the ID that we received in the return object
		if tc.wantStnd.ID == *new(uuid.UUID) {
			tc.wantStnd.ID = gotStnd.ID
		}

		diff := cmp.Diff(tc.wantStnd, gotStnd)
		if diff != "" {
			t.Fatalf("want %+v, got %+v, diff: %s", tc.wantStnd, gotStnd, diff)
		}

		wantSelect := []domain.Standings{tc.wantStnd}

		gotSelect, err := sr.Select(ctx, map[string]interface{}{"id": gotStnd.ID.String()}, false)
		if err != nil {
			t.Fatal(err)
		}

		diff = cmp.Diff(wantSelect, gotSelect)
		if diff != "" {
			t.Fatalf("want %+v, got %+v, diff: %s", wantSelect, gotSelect, cmp.Diff(wantSelect, gotSelect))
		}
	}
}

func TestRetrieveLatestStandingsWorker_HasFinalisedLastRound(t *testing.T) {
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
			want: false,
		},
		{
			stnd: domain.Standings{SeasonID: "season_id", Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
			}, Finalised: true},
			want: true, // season is completed by standings and standings are finalised
		},
	}

	for _, tc := range tt {
		r := domain.NewTestRetrieveLatestStandingsWorker(s, nil, nil, nil, nil, nil, nil, nil, nil)
		if r.HasFinalisedLastRound(tc.stnd) != tc.want {
			t.Fatalf("want %t, got %t", tc.want, !tc.want)
		}
	}
}

func TestRetrieveLatestStandingsWorker_IssueEmails(t *testing.T) {
	t.Run("happy path must issue the expected number of emails", func(t *testing.T) {
		tt := []struct {
			s              domain.Season
			st             domain.Standings
			seps           []domain.ScoredEntryPrediction
			wantCount      int
			wantFinalRound bool
		}{
			{
				s: domain.Season{
					ID:        "season_id",
					MaxRounds: 2,
				},
				st: domain.Standings{
					SeasonID: "season_id",
					Rankings: []domain.RankingWithMeta{
						{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings finalise the season
					},
				},
				seps: []domain.ScoredEntryPrediction{
					{EntryPredictionID: uuid.New(), Score: 123},
					{EntryPredictionID: uuid.New(), Score: 456},
					{EntryPredictionID: uuid.New(), Score: 789},
				},
				wantCount:      3,    // want 3 emails issued
				wantFinalRound: true, // email represents final round
			},
			{
				s: domain.Season{
					ID:        "season_id",
					MaxRounds: 123,
				},
				st: domain.Standings{
					SeasonID:  "season_id",
					Finalised: true,
					Rankings: []domain.RankingWithMeta{
						{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings are finalised but do not finalise season
					},
				},
				seps: []domain.ScoredEntryPrediction{
					{EntryPredictionID: uuid.New(), Score: 123},
					{EntryPredictionID: uuid.New(), Score: 456},
					{EntryPredictionID: uuid.New(), Score: 789},
				},
				wantCount:      3,     // want 3 emails issued
				wantFinalRound: false, // email does not represent final round
			},
			{
				s: domain.Season{
					ID:        "season_id",
					MaxRounds: 123, // non-finalised season
				},
				st: domain.Standings{
					SeasonID: "season_id",
					Rankings: []domain.RankingWithMeta{
						{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings are not finalised
					},
				},
				seps: []domain.ScoredEntryPrediction{
					{EntryPredictionID: uuid.New(), Score: 123},
					{EntryPredictionID: uuid.New(), Score: 456},
					{EntryPredictionID: uuid.New(), Score: 789},
				},
				// want no emails issued
			},
		}

		for _, tc := range tt {
			mrcei := &happyMockRoundCompleteEmailIssuer{
				t:              t,
				mux:            &sync.Mutex{},
				seps:           make(map[string]domain.ScoredEntryPrediction),
				wantFinalRound: tc.wantFinalRound,
			}

			r := domain.NewTestRetrieveLatestStandingsWorker(tc.s, nil, nil, nil, nil, nil, nil, mrcei, nil)

			if err := r.IssueEmails(context.Background(), tc.st, tc.seps); err != nil {
				t.Fatal(err)
			}

			if len(mrcei.seps) != tc.wantCount {
				t.Fatalf("want %d scored entry predictions, got %d", len(tc.seps), len(mrcei.seps))
			}

			if tc.wantCount > 0 {
				for _, tcsep := range tc.seps {
					sep, ok := mrcei.seps[tcsep.EntryPredictionID.String()]
					if !ok {
						t.Fatalf("scored entry prediction with score of %d is missing", tcsep.Score)
					}
					if !reflect.DeepEqual(tcsep, sep) {
						t.Fatalf("want scored entry prediction %+v, got %+v", tcsep, sep)
					}
				}
			}
		}
	})

	t.Run("errors returned by issue emails method invocations must be accumulated as expected", func(t *testing.T) {
		s := domain.Season{
			ID:        "season_id",
			MaxRounds: 2,
		}

		st := domain.Standings{
			SeasonID:  "season_id",
			Finalised: true,
			Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings finalise the season
			},
		}

		seps := []domain.ScoredEntryPrediction{
			{EntryPredictionID: uuid.New(), Score: 123},
			{EntryPredictionID: uuid.New(), Score: 456},
			{EntryPredictionID: uuid.New(), Score: 789},
		}

		wantCount := 3 // want 3 errors

		mrcei := &errMockRoundCompleteEmailIssuer{
			mux:  &sync.Mutex{},
			errs: make(map[string]error),
		}

		r := domain.NewTestRetrieveLatestStandingsWorker(s, nil, nil, nil, nil, nil, nil, mrcei, nil)

		err := r.IssueEmails(context.Background(), st, seps)
		mErr := domain.MultiError{}
		if !errors.As(err, &mErr) {
			t.Fatalf("want multierror, got %T", err)
		}

		if len(mErr.Errs) != wantCount {
			t.Fatalf("want %d errors, got %+v", wantCount, mErr.Errs)
		}
	})
}

func TestGenerateScoredEntryPrediction(t *testing.T) {
	okEntryPredictionRankings := domain.RankingCollection{
		{Position: 1, ID: pooleTownTeamID},
		{Position: 2, ID: wimborneTownTeamID},
		{Position: 3, ID: dorchesterTownTeamID},
		{Position: 4, ID: hamworthyUnitedTeamID},
		{Position: 5, ID: bournemouthPoppiesTeamID},
		{Position: 6, ID: stJohnsRangersTeamID},
		{Position: 7, ID: branksomeUnitedTeamID},
	}

	okStandingsRankings := []domain.RankingWithMeta{
		{Ranking: domain.Ranking{Position: 1, ID: branksomeUnitedTeamID}},    // score = 6 (prediction = 7)
		{Ranking: domain.Ranking{Position: 2, ID: stJohnsRangersTeamID}},     // score = 4 (prediction = 6)
		{Ranking: domain.Ranking{Position: 3, ID: bournemouthPoppiesTeamID}}, // score = 2 (prediction = 5)
		{Ranking: domain.Ranking{Position: 4, ID: hamworthyUnitedTeamID}},    // score = 0 (prediction = 4)
		{Ranking: domain.Ranking{Position: 5, ID: wimborneTownTeamID}},       // score = 3 (prediction = 2)
		{Ranking: domain.Ranking{Position: 6, ID: pooleTownTeamID}},          // score = 5 (prediction = 1)
		{Ranking: domain.Ranking{Position: 7, ID: dorchesterTownTeamID}},     // score = 4 (prediction = 3)
	}

	t.Run("valid entry prediction and standings must generate the expected scored entry prediction", func(t *testing.T) {
		entryPredictionID := parseUUID(t, `11111111-1111-1111-1111-111111111111`)
		entryPrediction := domain.EntryPrediction{
			ID:       entryPredictionID,
			Rankings: okEntryPredictionRankings,
		}

		standingsID := parseUUID(t, `22222222-2222-2222-2222-222222222222`)
		standings := domain.Standings{
			ID:       standingsID,
			Rankings: okStandingsRankings,
		}

		wantScoredEntryPrediction := &domain.ScoredEntryPrediction{
			EntryPredictionID: entryPredictionID,
			StandingsID:       standingsID,
			Rankings: []domain.RankingWithScore{
				{
					Ranking: domain.Ranking{Position: 1, ID: pooleTownTeamID},
					Score:   5,
				},
				{
					Ranking: domain.Ranking{Position: 2, ID: wimborneTownTeamID},
					Score:   3,
				},
				{
					Ranking: domain.Ranking{Position: 3, ID: dorchesterTownTeamID},
					Score:   4,
				},
				{
					Ranking: domain.Ranking{Position: 4, ID: hamworthyUnitedTeamID},
					Score:   0,
				},
				{
					Ranking: domain.Ranking{Position: 5, ID: bournemouthPoppiesTeamID},
					Score:   2,
				},
				{
					Ranking: domain.Ranking{Position: 6, ID: stJohnsRangersTeamID},
					Score:   4,
				},
				{
					Ranking: domain.Ranking{Position: 7, ID: branksomeUnitedTeamID},
					Score:   6,
				},
			},
			Score: 76, // 100 base points, minus 24 total hits (all of the above ranking scores added together)
		}

		gotScoredEntryPrediction, err := domain.GenerateScoredEntryPrediction(entryPrediction, standings)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "scored entry prediction", wantScoredEntryPrediction, gotScoredEntryPrediction)
	})

	t.Run("mismatch rankings count must produce expected error", func(t *testing.T) {
		entryPrediction := domain.EntryPrediction{
			Rankings: okEntryPredictionRankings,
		}

		wantErrMsg := "rankings count mismatch: submission 7: standings 0"
		_, gotErr := domain.GenerateScoredEntryPrediction(entryPrediction, domain.Standings{})
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})
}

type happyMockRoundCompleteEmailIssuer struct {
	domain.RoundCompleteEmailIssuer
	t              *testing.T
	mux            *sync.Mutex
	seps           map[string]domain.ScoredEntryPrediction
	wantFinalRound bool
}

func (h *happyMockRoundCompleteEmailIssuer) IssueRoundCompleteEmail(ctx context.Context, sep domain.ScoredEntryPrediction, isFinalRound bool) error {
	h.mux.Lock()
	defer h.mux.Unlock()
	h.seps[sep.EntryPredictionID.String()] = sep
	if h.wantFinalRound != isFinalRound {
		h.t.Fatalf("want final round %t, got %t", h.wantFinalRound, isFinalRound)
	}
	return nil
}

type errMockRoundCompleteEmailIssuer struct {
	domain.RoundCompleteEmailIssuer
	mux  *sync.Mutex
	errs map[string]error
}

func (e *errMockRoundCompleteEmailIssuer) IssueRoundCompleteEmail(ctx context.Context, sep domain.ScoredEntryPrediction, isFinalRound bool) error {
	e.mux.Lock()
	defer e.mux.Unlock()
	err := fmt.Errorf("error %s", sep.EntryPredictionID)
	e.errs[sep.EntryPredictionID.String()] = err
	return err
}
