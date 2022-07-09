package domain_test

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/domain"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRetrieveLatestStandingsWorker(t *testing.T) {
	tColl := make(map[string]domain.Team)
	cl := &mockClock{}
	l := &mockLogger{}
	ea := &domain.EntryAgent{}
	sa := &domain.StandingsAgent{}
	sepa := &domain.ScoredEntryPredictionAgent{}
	ca := &domain.CommunicationsAgent{}
	fds := &domain.NoopFootballDataSource{}

	tt := []struct {
		name        string
		tColl       domain.TeamCollection
		cl          domain.Clock
		l           domain.Logger
		ea          *domain.EntryAgent
		sa          *domain.StandingsAgent
		sepa        *domain.ScoredEntryPredictionAgent
		emailIssuer domain.RoundCompleteEmailIssuer
		fds         domain.FootballDataSource
		wantErr     bool
	}{
		{"missing team collection", nil, cl, l, ea, sa, sepa, ca, fds, true},
		{"missing clock", tColl, nil, l, ea, sa, sepa, ca, fds, true},
		{"missing logger", tColl, cl, nil, ea, sa, sepa, ca, fds, true},
		{"missing entry agent", tColl, cl, l, nil, sa, sepa, ca, fds, true},
		{"missing standings agent", tColl, cl, l, ea, nil, sepa, ca, fds, true},
		{"missing scored entry predictions agent", tColl, cl, l, ea, sa, nil, ca, fds, true},
		{"missing communications agent", tColl, cl, l, ea, sa, sepa, nil, fds, true},
		{"missing football client", tColl, cl, l, ea, sa, sepa, ca, nil, true},
		{"no missing dependencies", tColl, cl, l, ea, sa, sepa, ca, fds, false},
	}
	for idx, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			params := domain.RetrieveLatestStandingsWorkerParams{
				TeamCollection:             tc.tColl,
				Clock:                      tc.cl,
				Logger:                     tc.l,
				EntryAgent:                 tc.ea,
				StandingsAgent:             tc.sa,
				ScoredEntryPredictionAgent: tc.sepa,
				EmailIssuer:                tc.emailIssuer,
				FootballClient:             tc.fds,
			}

			w, gotErr := domain.NewRetrieveLatestStandingsWorker(params)
			if tc.wantErr && !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("tc #%d: want ErrIsNil, got %s (%T)", idx, gotErr, gotErr)
			}
			if !tc.wantErr && w == nil {
				t.Fatalf("tc #%d: want non-empty worker, got nil", idx)
			}
		})
	}
}

func TestRetrieveLatestStandingsWorker_ProcessExistingStandings(t *testing.T) {
	t.Run("happy path must produce the expected results", func(t *testing.T) {
		t.Cleanup(truncate)

		ctx := context.Background()

		now := time.Now().Truncate(time.Second)

		existingStandings := domain.Standings{
			ID:          uuid.New(),
			SeasonID:    "season_id",
			RoundNumber: 123,
			Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 1234}},
			},
			CreatedAt: now,
		}

		if err := sr.Insert(ctx, &existingStandings); err != nil {
			t.Fatal(err)
		}

		clientStandings := domain.Standings{Rankings: []domain.RankingWithMeta{
			{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5678}},
		}}

		wantStandings := existingStandings
		wantStandings.Rankings = clientStandings.Rankings
		wantStandings.UpdatedAt = &now

		wantRetrieved := []domain.Standings{wantStandings}

		sa, err := domain.NewStandingsAgent(sr)
		if err != nil {
			t.Fatal(sa)
		}

		params := domain.RetrieveLatestStandingsWorkerParams{
			Season:         domain.Season{},
			StandingsAgent: sa,
		}
		worker := newTestRetrieveLatestStandingsWorker(t, params)

		gotStandings, err := worker.ProcessExistingStandings(ctx, existingStandings, clientStandings)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "processed standings", wantStandings, gotStandings)

		gotRetrieved, err := sr.Select(ctx, map[string]interface{}{"id": existingStandings.ID.String()}, false)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "retrieved standings", wantRetrieved, gotRetrieved)
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

	params := domain.RetrieveLatestStandingsWorkerParams{
		Season:         domain.Season{},
		StandingsAgent: sa,
	}
	worker := newTestRetrieveLatestStandingsWorker(t, params)

	id1 := uuid.New()
	id2 := uuid.New()

	tt := []struct {
		seededStandings    *domain.Standings // standings seeded prior to test case run
		standingsToProcess domain.Standings  // standings to process / compare against seeded standings
		wantStandings      domain.Standings  // expected result
	}{
		{
			standingsToProcess: domain.Standings{
				RoundNumber: 1, // process first round that does not yet exist (no seed), so expect this to be created straight away
				CreatedAt:   now,
			},
			wantStandings: domain.Standings{
				RoundNumber: 1, // expect process Standings to be created and returned
				CreatedAt:   now,
			},
		},
		{
			seededStandings: &domain.Standings{
				ID:          id1,
				RoundNumber: 2,
				CreatedAt:   now,
			},
			standingsToProcess: domain.Standings{
				ID:          id1,
				RoundNumber: 3, // attempt to process next round when previous round has not been finalised
				CreatedAt:   now,
			},
			wantStandings: domain.Standings{
				ID:          id1,
				RoundNumber: 2, // expected seeded Standings to be updated (finalised) and returned
				Finalised:   true,
				CreatedAt:   now,
				UpdatedAt:   &now,
			},
		},
		{
			seededStandings: &domain.Standings{
				ID:          id2,
				RoundNumber: 3,
				Finalised:   true,
				CreatedAt:   now,
			},
			standingsToProcess: domain.Standings{
				RoundNumber: 4, // attempt to process next round when previous round HAS been finalised
				CreatedAt:   now,
			},
			wantStandings: domain.Standings{
				RoundNumber: 4, // expect process Standings to be created and returned
				CreatedAt:   now,
			},
		},
	}

	for _, tc := range tt {
		if tc.seededStandings != nil {
			if err := sr.Insert(ctx, tc.seededStandings); err != nil {
				t.Fatal(err)
			}
		}

		gotStandings, err := worker.ProcessNewStandings(ctx, tc.standingsToProcess)
		if err != nil {
			t.Fatal(err)
		}

		// if the wanted Standings object doesn't specify an ID, we're expecting it to be created by our test method
		// so populate the ID that we received in the return object
		if tc.wantStandings.ID == *new(uuid.UUID) {
			tc.wantStandings.ID = gotStandings.ID
		}

		cmpDiff(t, "processed standings", tc.wantStandings, gotStandings)

		wantRetrieved := []domain.Standings{tc.wantStandings}
		gotRetrieved, err := sr.Select(ctx, map[string]interface{}{"id": gotStandings.ID.String()}, false)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "retrieved standings", wantRetrieved, gotRetrieved)
	}
}

func TestRetrieveLatestStandingsWorker_HasFinalisedLastRound(t *testing.T) {
	season := domain.Season{ID: "season_id", MaxRounds: 5}

	tt := []struct {
		standings domain.Standings
		want      bool
	}{
		{
			standings: domain.Standings{SeasonID: "alt_season_id"},
			want:      false,
		},
		{
			standings: domain.Standings{SeasonID: "season_id", Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 4}}, // one short of max rounds
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
			}},
			want: false,
		},
		{
			standings: domain.Standings{SeasonID: "season_id", Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
			}},
			want: false,
		},
		{
			standings: domain.Standings{SeasonID: "season_id", Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 5}},
			}, Finalised: true},
			want: true, // season is completed by standings and standings are finalised
		},
	}

	for _, tc := range tt {
		params := domain.RetrieveLatestStandingsWorkerParams{Season: season}
		worker := newTestRetrieveLatestStandingsWorker(t, params)
		if worker.HasFinalisedLastRound(tc.standings) != tc.want {
			t.Fatalf("want finalised last round is %t, got %t", tc.want, !tc.want)
		}
	}
}

func TestRetrieveLatestStandingsWorker_IssueEmails(t *testing.T) {
	t.Run("happy path must issue the expected number of emails", func(t *testing.T) {
		tt := []struct {
			season                 domain.Season
			standings              domain.Standings
			scoredEntryPredictions []domain.ScoredEntryPrediction // email recipients (via entry prediction id -> entry id)
			wantCount              int
			wantFinalRound         bool
		}{
			{
				season: domain.Season{
					ID:        "season_id",
					MaxRounds: 2,
				},
				standings: domain.Standings{
					SeasonID: "season_id",
					Rankings: []domain.RankingWithMeta{
						{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings finalise the season
					},
				},
				scoredEntryPredictions: []domain.ScoredEntryPrediction{
					{EntryPredictionID: uuid.New(), Score: 123},
					{EntryPredictionID: uuid.New(), Score: 456},
					{EntryPredictionID: uuid.New(), Score: 789},
				},
				wantCount:      3,    // want 3 emails issued
				wantFinalRound: true, // email represents final round
			},
			{
				season: domain.Season{
					ID:        "season_id",
					MaxRounds: 123,
				},
				standings: domain.Standings{
					SeasonID:  "season_id",
					Finalised: true,
					Rankings: []domain.RankingWithMeta{
						{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings are finalised but do not finalise season
					},
				},
				scoredEntryPredictions: []domain.ScoredEntryPrediction{
					{EntryPredictionID: uuid.New(), Score: 123},
					{EntryPredictionID: uuid.New(), Score: 456},
					{EntryPredictionID: uuid.New(), Score: 789},
				},
				wantCount:      3,     // want 3 emails issued
				wantFinalRound: false, // email does not represent final round
			},
			{
				season: domain.Season{
					ID:        "season_id",
					MaxRounds: 123, // non-finalised season
				},
				standings: domain.Standings{
					SeasonID: "season_id",
					Rankings: []domain.RankingWithMeta{
						{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings are not finalised
					},
				},
				scoredEntryPredictions: []domain.ScoredEntryPrediction{
					{EntryPredictionID: uuid.New(), Score: 123},
					{EntryPredictionID: uuid.New(), Score: 456},
					{EntryPredictionID: uuid.New(), Score: 789},
				},
				// want no emails issued
			},
		}

		for _, tc := range tt {
			emailIssuer := &happyMockRoundCompleteEmailIssuer{
				t:              t,
				mux:            &sync.Mutex{},
				seps:           make(map[string]domain.ScoredEntryPrediction),
				wantFinalRound: tc.wantFinalRound,
			}

			params := domain.RetrieveLatestStandingsWorkerParams{
				Season:      tc.season,
				EmailIssuer: emailIssuer,
			}
			worker := newTestRetrieveLatestStandingsWorker(t, params)

			if err := worker.IssueEmails(context.Background(), tc.standings, tc.scoredEntryPredictions); err != nil {
				t.Fatal(err)
			}

			if len(emailIssuer.seps) != tc.wantCount {
				t.Fatalf("want %d emails issued, got %d", tc.wantCount, len(emailIssuer.seps))
			}

			if tc.wantCount > 0 {
				for _, tcSEP := range tc.scoredEntryPredictions {
					sep, ok := emailIssuer.seps[tcSEP.EntryPredictionID.String()]
					if !ok {
						t.Fatalf("scored entry prediction with score of %d is missing", tcSEP.Score)
					}
					cmpDiff(t, "scored entry prediction", tcSEP, sep)
				}
			}
		}
	})

	t.Run("errors returned by issue emails method invocations must be accumulated as expected", func(t *testing.T) {
		season := domain.Season{
			ID:        "season_id",
			MaxRounds: 2,
		}

		standings := domain.Standings{
			SeasonID:  "season_id",
			Finalised: true,
			Rankings: []domain.RankingWithMeta{
				{MetaData: map[string]int{domain.MetaKeyPlayedGames: 2}}, // standings finalise the season
			},
		}

		scoredEntryPredictions := []domain.ScoredEntryPrediction{
			{EntryPredictionID: uuid.New(), Score: 123},
			{EntryPredictionID: uuid.New(), Score: 456},
			{EntryPredictionID: uuid.New(), Score: 789},
		}

		wantCount := 3 // want 3 errors

		emailIssuer := &errMockRoundCompleteEmailIssuer{
			mux:  &sync.Mutex{},
			errs: make(map[string]error),
		}

		params := domain.RetrieveLatestStandingsWorkerParams{
			Season:      season,
			EmailIssuer: emailIssuer,
		}
		worker := newTestRetrieveLatestStandingsWorker(t, params)

		err := worker.IssueEmails(context.Background(), standings, scoredEntryPredictions)
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

func newTestRetrieveLatestStandingsWorker(
	t *testing.T,
	params domain.RetrieveLatestStandingsWorkerParams,
) *domain.RetrieveLatestStandingsWorker {
	t.Helper()

	if params.TeamCollection == nil {
		params.TeamCollection = make(domain.TeamCollection, 0)
	}
	if params.Clock == nil {
		params.Clock = &mockClock{}
	}
	if params.Logger == nil {
		params.Logger = &mockLogger{}
	}
	if params.EntryAgent == nil {
		params.EntryAgent = &domain.EntryAgent{}
	}
	if params.StandingsAgent == nil {
		params.StandingsAgent = &domain.StandingsAgent{}
	}
	if params.ScoredEntryPredictionAgent == nil {
		params.ScoredEntryPredictionAgent = &domain.ScoredEntryPredictionAgent{}
	}
	if params.EmailIssuer == nil {
		params.EmailIssuer = &domain.CommunicationsAgent{}
	}
	if params.FootballClient == nil {
		params.FootballClient = &footballdataorg.Client{}
	}

	worker, err := domain.NewRetrieveLatestStandingsWorker(params)
	if err != nil {
		t.Fatal(err)
	}

	return worker
}
