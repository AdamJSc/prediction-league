package domain_test

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	emptyCommunicationsAgent        = &domain.CommunicationsAgent{}
	emptyEntryAgent                 = &domain.EntryAgent{}
	emptyMatchWeekResultAgent       = &domain.MatchWeekResultAgent{}
	emptyMatchWeekSubmissionAgent   = &domain.MatchWeekSubmissionAgent{}
	emptyScoredEntryPredictionAgent = &domain.ScoredEntryPredictionAgent{}
	emptyStandingsAgent             = &domain.StandingsAgent{}
	noopFootballDataClient          = &domain.NoopFootballDataSource{}
)

func TestNewRetrieveLatestStandingsWorker(t *testing.T) {
	tColl := make(map[string]domain.Team)
	cl := &mockClock{}
	l := &mockLogger{}
	ea := emptyEntryAgent
	sa := emptyStandingsAgent
	sepa := emptyScoredEntryPredictionAgent
	mwsa := emptyMatchWeekSubmissionAgent
	mwra := emptyMatchWeekResultAgent
	ca := emptyCommunicationsAgent
	fcl := noopFootballDataClient

	tt := []struct {
		name        string
		tColl       domain.TeamCollection
		cl          domain.Clock
		l           domain.Logger
		ea          *domain.EntryAgent
		sa          *domain.StandingsAgent
		sepa        *domain.ScoredEntryPredictionAgent
		mwsa        *domain.MatchWeekSubmissionAgent
		mwra        *domain.MatchWeekResultAgent
		emailIssuer domain.RoundCompleteEmailIssuer
		fcl         domain.FootballDataSource
		wantErr     bool
	}{
		{"missing team collection", nil, cl, l, ea, sa, sepa, mwsa, mwra, ca, fcl, true},
		{"missing clock", tColl, nil, l, ea, sa, sepa, mwsa, mwra, ca, fcl, true},
		{"missing logger", tColl, cl, nil, ea, sa, sepa, mwsa, mwra, ca, fcl, true},
		{"missing entry agent", tColl, cl, l, nil, sa, sepa, mwsa, mwra, ca, fcl, true},
		{"missing standings agent", tColl, cl, l, ea, nil, sepa, mwsa, mwra, ca, fcl, true},
		{"missing scored entry predictions agent", tColl, cl, l, ea, sa, nil, mwsa, mwra, ca, fcl, true},
		{"missing match week submission agent", tColl, cl, l, ea, sa, sepa, nil, mwra, ca, fcl, true},
		{"missing match week result agent", tColl, cl, l, ea, sa, sepa, mwsa, nil, ca, fcl, true},
		{"missing communications agent", tColl, cl, l, ea, sa, sepa, mwsa, mwra, nil, fcl, true},
		{"missing football client", tColl, cl, l, ea, sa, sepa, mwsa, mwra, ca, nil, true},
		{"no missing dependencies", tColl, cl, l, ea, sa, sepa, mwsa, mwra, ca, fcl, false},
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
				MatchWeekSubmissionAgent:   tc.mwsa,
				MatchWeekResultAgent:       tc.mwra,
				EmailIssuer:                tc.emailIssuer,
				FootballClient:             tc.fcl,
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

		worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
			Season:         domain.Season{},
			StandingsAgent: sa,
		})

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

	worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
		Season:         domain.Season{},
		StandingsAgent: sa,
	})

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
		worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
			Season: season,
		})
		if worker.HasFinalisedLastRound(tc.standings) != tc.want {
			t.Fatalf("want finalised last round is %t, got %t", tc.want, !tc.want)
		}
	}
}

func TestRetrieveLatestStandingsWorker_GenerateScoredEntryPrediction(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

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

	// wantResultRankings defines the expected outcome of generating the scored entry prediction
	// from both sets of rankings above
	wantResultRankings := []domain.ResultTeamRanking{
		{TeamRanking: domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID}, StandingsPos: 6, Hit: 5},
		{TeamRanking: domain.TeamRanking{Position: 2, TeamID: wimborneTownTeamID}, StandingsPos: 5, Hit: 3},
		{TeamRanking: domain.TeamRanking{Position: 3, TeamID: dorchesterTownTeamID}, StandingsPos: 7, Hit: 4},
		{TeamRanking: domain.TeamRanking{Position: 4, TeamID: hamworthyUnitedTeamID}, StandingsPos: 4, Hit: 0},
		{TeamRanking: domain.TeamRanking{Position: 5, TeamID: bournemouthPoppiesTeamID}, StandingsPos: 3, Hit: 2},
		{TeamRanking: domain.TeamRanking{Position: 6, TeamID: stJohnsRangersTeamID}, StandingsPos: 2, Hit: 4},
		{TeamRanking: domain.TeamRanking{Position: 7, TeamID: branksomeUnitedTeamID}, StandingsPos: 1, Hit: 6},
	}

	t.Run("valid entry prediction and standings must generate the expected scored entry prediction", func(t *testing.T) {
		submissionRepoID := parseUUID(t, uuidAll1s)
		submissionRepoDate := testDate.Add(-2 * time.Hour)
		mwSubmissionRepo := newMatchWeekSubmissionRepo(t, submissionRepoID, submissionRepoDate)
		mwSubmissionAgent := newMatchWeekSubmissionAgent(t, mwSubmissionRepo)

		resultRepoDate := testDate.Add(-time.Hour)
		mwResultRepo := newMatchWeekResultRepo(t, resultRepoDate)
		mwResultAgent := newMatchWeekResultAgent(t, mwResultRepo)

		worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
			MatchWeekSubmissionAgent: mwSubmissionAgent,
			MatchWeekResultAgent:     mwResultAgent,
		})

		seededEntry := seedEntry(t, generateEntry())

		entryPrediction := domain.EntryPrediction{
			ID:       uuid.New(),
			EntryID:  seededEntry.ID,
			Rankings: okEntryPredictionRankings,
		}

		standings := domain.Standings{
			ID:       uuid.New(),
			Rankings: okStandingsRankings,
		}

		wantScoredEntryPrediction := &domain.ScoredEntryPrediction{
			EntryPredictionID: entryPrediction.ID,
			StandingsID:       standings.ID,
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

		gotScoredEntryPrediction, err := worker.GenerateScoredEntryPrediction(ctx, entryPrediction, standings)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "scored entry prediction", wantScoredEntryPrediction, gotScoredEntryPrediction)

		// ensure that the expected match week submission was inserted
		wantMWSubmission := &domain.MatchWeekSubmission{
			ID:                      submissionRepoID,
			EntryID:                 seededEntry.ID,
			MatchWeekNumber:         uint16(standings.RoundNumber),
			TeamRankings:            newSubmissionRankings(entryPrediction.Rankings),
			LegacyEntryPredictionID: entryPrediction.ID,
			CreatedAt:               submissionRepoDate,
		}

		gotMWSubmission, err := mwSubmissionRepo.GetByID(ctx, submissionRepoID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "inserted match week submission", wantMWSubmission, gotMWSubmission)

		// ensure that the expected match week result was inserted
		wantMWResult := &domain.MatchWeekResult{
			MatchWeekSubmissionID: submissionRepoID,
			TeamRankings:          wantResultRankings,
			Score:                 int64(gotScoredEntryPrediction.Score),
			Modifiers: []domain.ModifierSummary{
				{Code: "BASE_SCORE", Value: 100},
				{Code: "RANKINGS_HIT", Value: -24},
			},
			CreatedAt: resultRepoDate,
		}

		gotMWResult, err := mwResultRepo.GetBySubmissionID(ctx, submissionRepoID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "inserted match week result", wantMWResult, gotMWResult)
	})

	// TODO: feat - add test case for generating scored entry prediction that already exists as match week submission + result

	t.Run("failure to upsert match week submission must return the expected error", func(t *testing.T) {
		badMWSubmissionRepo, err := mysqldb.NewMatchWeekSubmissionRepo(badDB, newUUIDFunc(parseUUID(t, uuidAll2s)), nil)
		if err != nil {
			t.Fatal(err)
		}
		badMWSubmissionAgent := newMatchWeekSubmissionAgent(t, badMWSubmissionRepo)

		worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
			MatchWeekSubmissionAgent: badMWSubmissionAgent,
		})

		seededEntry := seedEntry(t, generateEntry())

		entryPrediction := domain.EntryPrediction{
			ID:       uuid.New(),
			EntryID:  seededEntry.ID,
			Rankings: okEntryPredictionRankings,
		}

		standings := domain.Standings{
			ID:       uuid.New(),
			Rankings: okStandingsRankings,
		}

		wantErrMsg := "cannot get submission by legacy id: default addr for network 'connectionString' unknown"
		_, gotErr := worker.GenerateScoredEntryPrediction(ctx, entryPrediction, standings)
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})

	t.Run("failure to upsert match week result must return the expected error", func(t *testing.T) {
		mwSubmissionRepo := newMatchWeekSubmissionRepo(t, parseUUID(t, uuidAll3s), testDate)
		mwSubmissionAgent := newMatchWeekSubmissionAgent(t, mwSubmissionRepo)

		badMWResultRepo, err := mysqldb.NewMatchWeekResultRepo(badDB, nil)
		if err != nil {
			t.Fatal(err)
		}
		badMWResultAgent := newMatchWeekResultAgent(t, badMWResultRepo)

		worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
			MatchWeekSubmissionAgent: mwSubmissionAgent,
			MatchWeekResultAgent:     badMWResultAgent,
		})

		seededEntry := seedEntry(t, generateEntry())

		entryPrediction := domain.EntryPrediction{
			ID:       uuid.New(),
			EntryID:  seededEntry.ID,
			Rankings: okEntryPredictionRankings,
		}

		standings := domain.Standings{
			ID:       uuid.New(),
			Rankings: okStandingsRankings,
		}

		wantErrMsg := "cannot get match week result by submission id: default addr for network 'connectionString' unknown"
		_, gotErr := worker.GenerateScoredEntryPrediction(ctx, entryPrediction, standings)
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})

	t.Run("mismatch rankings count must produce expected error", func(t *testing.T) {
		worker := &domain.RetrieveLatestStandingsWorker{}

		entryPrediction := domain.EntryPrediction{
			Rankings: okEntryPredictionRankings,
		}

		wantErrMsg := "rankings count mismatch: submission 7: standings 0"
		_, gotErr := worker.GenerateScoredEntryPrediction(ctx, entryPrediction, domain.Standings{})
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})
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

			worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
				Season:      tc.season,
				EmailIssuer: emailIssuer,
			})

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

		worker := newTestRetrieveLatestStandingsWorker(t, domain.RetrieveLatestStandingsWorkerParams{
			Season:      season,
			EmailIssuer: emailIssuer,
		})

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
		params.EntryAgent = emptyEntryAgent
	}
	if params.StandingsAgent == nil {
		params.StandingsAgent = emptyStandingsAgent
	}
	if params.ScoredEntryPredictionAgent == nil {
		params.ScoredEntryPredictionAgent = emptyScoredEntryPredictionAgent
	}
	if params.MatchWeekSubmissionAgent == nil {
		params.MatchWeekSubmissionAgent = emptyMatchWeekSubmissionAgent
	}
	if params.MatchWeekResultAgent == nil {
		params.MatchWeekResultAgent = emptyMatchWeekResultAgent
	}
	if params.EmailIssuer == nil {
		params.EmailIssuer = emptyCommunicationsAgent
	}
	if params.FootballClient == nil {
		params.FootballClient = noopFootballDataClient
	}

	worker, err := domain.NewRetrieveLatestStandingsWorker(params)
	if err != nil {
		t.Fatal(err)
	}

	return worker
}

func newSubmissionRankings(rc domain.RankingCollection) []domain.TeamRanking {
	rankings := make([]domain.TeamRanking, 0)

	for _, r := range rc {
		rankings = append(rankings, domain.TeamRanking{
			Position: uint16(r.Position),
			TeamID:   r.ID,
		})
	}

	return rankings
}
