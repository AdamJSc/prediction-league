package domain_test

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

const (
	bournemouthPoppiesTeamID = "BFC"
	branksomeUnitedTeamID    = "BUFC"
	dorchesterTownTeamID     = "DTFC"
	hamworthyUnitedTeamID    = "HUFC"
	pooleTownTeamID          = "PTFC"
	stJohnsRangersTeamID     = "SJRFC"
	wimborneTownTeamID       = "WTFC"
)

var (
	altModifierSummaries = []domain.ModifierSummary{
		{Code: domain.TeamRankingsHitModifierCode, Value: 9999},
	}

	modifierSummaries = []domain.ModifierSummary{
		{Code: domain.BaseScoreModifierCode, Value: 100},
		{Code: domain.TeamRankingsHitModifierCode, Value: 88},
	}

	randomiser = rand.New(rand.NewSource(time.Now().UnixNano()))

	resultTeamRankings = []domain.ResultTeamRanking{
		{TeamRanking: domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID}, StandingsPos: 7, Hit: 6},
		{TeamRanking: domain.TeamRanking{Position: 2, TeamID: wimborneTownTeamID}, StandingsPos: 6, Hit: 4},
		{TeamRanking: domain.TeamRanking{Position: 3, TeamID: dorchesterTownTeamID}, StandingsPos: 5, Hit: 2},
		{TeamRanking: domain.TeamRanking{Position: 4, TeamID: hamworthyUnitedTeamID}, StandingsPos: 4, Hit: 0},
		{TeamRanking: domain.TeamRanking{Position: 5, TeamID: bournemouthPoppiesTeamID}, StandingsPos: 3, Hit: 2},
		{TeamRanking: domain.TeamRanking{Position: 6, TeamID: stJohnsRangersTeamID}, StandingsPos: 2, Hit: 4},
		{TeamRanking: domain.TeamRanking{Position: 7, TeamID: branksomeUnitedTeamID}, StandingsPos: 1, Hit: 6},
	}
)

func TestNewMatchWeekResult(t *testing.T) {
	id := mustGetUUIDFromString(t, `12345678-1234-1234-1234-123456789012`)

	t.Run("no modifiers must populate match week result as expected", func(t *testing.T) {
		wantMWResult := &domain.MatchWeekResult{
			MatchWeekSubmissionID: id,
		}

		gotMWResult, err := domain.NewMatchWeekResult(id)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", wantMWResult, gotMWResult)
	})

	t.Run("one modifier must populate match week result as expected", func(t *testing.T) {
		testMod := func(result *domain.MatchWeekResult) error {
			result.Score = 100
			return nil
		}

		wantMWResult := &domain.MatchWeekResult{
			MatchWeekSubmissionID: id,
			Score:                 100,
		}

		gotMWResult, err := domain.NewMatchWeekResult(id, testMod)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", wantMWResult, gotMWResult)
	})

	t.Run("multiple modifiers must populate match week result as expected", func(t *testing.T) {
		testMod1 := func(result *domain.MatchWeekResult) error {
			result.Score = 100
			return nil
		}

		testModNoEffect := func(result *domain.MatchWeekResult) error {
			return nil
		}

		testMod2 := func(result *domain.MatchWeekResult) error {
			result.Score = result.Score * 3
			return nil
		}

		wantMWResult := &domain.MatchWeekResult{
			MatchWeekSubmissionID: id,
			Score:                 300,
		}

		gotMWResult, err := domain.NewMatchWeekResult(id, testMod1, testModNoEffect, testMod2)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", wantMWResult, gotMWResult)
	})

	t.Run("modifier error must fail", func(t *testing.T) {
		testMod1 := func(result *domain.MatchWeekResult) error {
			result.Score = 100
			return nil
		}

		testModError := func(result *domain.MatchWeekResult) error {
			return errors.New("sad times :'(")
		}

		testMod2 := func(result *domain.MatchWeekResult) error {
			result.Score = result.Score * 3
			return nil
		}

		_, gotErr := domain.NewMatchWeekResult(id, testMod1, testModError, testMod2)
		cmpErrorMsg(t, "sad times :'(", gotErr)
	})
}

func TestBaseScoreModifier(t *testing.T) {
	t.Run("setting base score must produce the expected match week result", func(t *testing.T) {
		modifier := domain.BaseScoreModifier(5678)

		wantMWResult := &domain.MatchWeekResult{
			Score: 5678,
			Modifiers: []domain.ModifierSummary{
				{
					Code:  "BASE_SCORE",
					Value: 5678,
				},
			},
		}

		gotMWResult := &domain.MatchWeekResult{
			Score: 1234, // should override value
		}
		if err := modifier(gotMWResult); err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", wantMWResult, gotMWResult)
	})
}

func TestTeamRankingsHitModifier(t *testing.T) {
	okSubmissionRankings := []domain.TeamRanking{
		{Position: 1, TeamID: pooleTownTeamID},
		{Position: 2, TeamID: wimborneTownTeamID},
		{Position: 3, TeamID: dorchesterTownTeamID},
		{Position: 4, TeamID: hamworthyUnitedTeamID},
		{Position: 5, TeamID: bournemouthPoppiesTeamID},
		{Position: 6, TeamID: stJohnsRangersTeamID},
		{Position: 7, TeamID: branksomeUnitedTeamID},
	}

	okStandingsRankings := []domain.StandingsTeamRanking{
		{TeamRanking: domain.TeamRanking{Position: 1, TeamID: branksomeUnitedTeamID}},    // hit = 6 (submission = 7)
		{TeamRanking: domain.TeamRanking{Position: 2, TeamID: stJohnsRangersTeamID}},     // hit = 4 (submission = 6)
		{TeamRanking: domain.TeamRanking{Position: 3, TeamID: bournemouthPoppiesTeamID}}, // hit = 2 (submission = 5)
		{TeamRanking: domain.TeamRanking{Position: 4, TeamID: hamworthyUnitedTeamID}},    // hit = 0 (submission = 4)
		{TeamRanking: domain.TeamRanking{Position: 5, TeamID: wimborneTownTeamID}},       // hit = 3 (submission = 2)
		{TeamRanking: domain.TeamRanking{Position: 6, TeamID: pooleTownTeamID}},          // hit = 5 (submission = 1)
		{TeamRanking: domain.TeamRanking{Position: 7, TeamID: dorchesterTownTeamID}},     // hit = 4 (submission = 3)
	}

	t.Run("valid submission and standings must produce the expected match week result", func(t *testing.T) {
		submission := &domain.MatchWeekSubmission{
			TeamRankings: randomiseTeamRankings(okSubmissionRankings), // test method must sort these by position ascending
		}

		standings := &domain.MatchWeekStandings{
			TeamRankings: randomiseStandingsTeamRankings(okStandingsRankings), // test method must sort these by position ascending
		}

		wantMWResult := &domain.MatchWeekResult{
			TeamRankings: []domain.ResultTeamRanking{
				{
					TeamRanking:  domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID},
					StandingsPos: 6,
					Hit:          5,
				},
				{
					TeamRanking:  domain.TeamRanking{Position: 2, TeamID: wimborneTownTeamID},
					StandingsPos: 5,
					Hit:          3,
				},
				{
					TeamRanking:  domain.TeamRanking{Position: 3, TeamID: dorchesterTownTeamID},
					StandingsPos: 7,
					Hit:          4,
				},
				{
					TeamRanking:  domain.TeamRanking{Position: 4, TeamID: hamworthyUnitedTeamID},
					StandingsPos: 4,
					Hit:          0,
				},
				{
					TeamRanking:  domain.TeamRanking{Position: 5, TeamID: bournemouthPoppiesTeamID},
					StandingsPos: 3,
					Hit:          2,
				},
				{
					TeamRanking:  domain.TeamRanking{Position: 6, TeamID: stJohnsRangersTeamID},
					StandingsPos: 2,
					Hit:          4,
				},
				{
					TeamRanking:  domain.TeamRanking{Position: 7, TeamID: branksomeUnitedTeamID},
					StandingsPos: 1,
					Hit:          6,
				},
			},
			Score: -24,
			Modifiers: []domain.ModifierSummary{
				{Code: "RANKINGS_HIT", Value: -24},
			},
		}

		modifier := domain.TeamRankingsHitModifier(submission, standings)

		gotMWResult := &domain.MatchWeekResult{}
		if err := modifier(gotMWResult); err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", wantMWResult, gotMWResult)
	})

	t.Run("nil submission must produce empty match week result", func(t *testing.T) {
		modifier := domain.TeamRankingsHitModifier(nil, &domain.MatchWeekStandings{})

		gotMWResult := &domain.MatchWeekResult{}
		if err := modifier(gotMWResult); err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", &domain.MatchWeekResult{}, gotMWResult)
	})

	t.Run("nil standings must produce empty match week result", func(t *testing.T) {
		modifier := domain.TeamRankingsHitModifier(&domain.MatchWeekSubmission{}, nil)

		gotMWResult := &domain.MatchWeekResult{}
		if err := modifier(gotMWResult); err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", &domain.MatchWeekResult{}, gotMWResult)
	})

	t.Run("mismatch rankings count must produce expected error", func(t *testing.T) {
		modifier := domain.TeamRankingsHitModifier(
			&domain.MatchWeekSubmission{TeamRankings: okSubmissionRankings},
			&domain.MatchWeekStandings{},
		)

		wantErrMsg := "rankings count mismatch: submission 7: standings 0"
		gotErr := modifier(&domain.MatchWeekResult{})
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})

	t.Run("duplicate team ids in submission rankings must produce expected error", func(t *testing.T) {
		modifier := domain.TeamRankingsHitModifier(
			&domain.MatchWeekSubmission{TeamRankings: []domain.TeamRanking{
				{Position: 1, TeamID: pooleTownTeamID},
				{Position: 2, TeamID: wimborneTownTeamID},
				{Position: 3, TeamID: hamworthyUnitedTeamID},
				{Position: 4, TeamID: dorchesterTownTeamID},
				{Position: 5, TeamID: pooleTownTeamID},
				{Position: 6, TeamID: wimborneTownTeamID},
				{Position: 7, TeamID: pooleTownTeamID},
			}},
			&domain.MatchWeekStandings{TeamRankings: okStandingsRankings},
		)

		wantErrMsg := "submission team rankings: duplicate team ids found: 'PTFC' (3), 'WTFC' (2)"
		gotErr := modifier(&domain.MatchWeekResult{})
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})

	t.Run("duplicate team ids in standings rankings must produce expected error", func(t *testing.T) {
		modifier := domain.TeamRankingsHitModifier(
			&domain.MatchWeekSubmission{TeamRankings: okSubmissionRankings},
			&domain.MatchWeekStandings{TeamRankings: []domain.StandingsTeamRanking{
				{TeamRanking: domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 2, TeamID: wimborneTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 3, TeamID: hamworthyUnitedTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 4, TeamID: dorchesterTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 5, TeamID: pooleTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 6, TeamID: wimborneTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 7, TeamID: pooleTownTeamID}},
			}},
		)

		wantErrMsg := "standings team rankings: duplicate team ids found: 'PTFC' (3), 'WTFC' (2)"
		gotErr := modifier(&domain.MatchWeekResult{})
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})

	t.Run("missing team ids must produce expected error", func(t *testing.T) {
		modifier := domain.TeamRankingsHitModifier(
			&domain.MatchWeekSubmission{TeamRankings: []domain.TeamRanking{
				{Position: 1, TeamID: pooleTownTeamID},
				{Position: 2, TeamID: wimborneTownTeamID},
				{Position: 3, TeamID: "BOSTON_RED_SOX"},
				{Position: 4, TeamID: hamworthyUnitedTeamID},
				{Position: 5, TeamID: "EDMONTON_OILERS"},
				{Position: 6, TeamID: stJohnsRangersTeamID},
				{Position: 7, TeamID: "DARTFORD_TIDDLYWINKS_MASSIF"},
			}},
			&domain.MatchWeekStandings{TeamRankings: []domain.StandingsTeamRanking{
				{TeamRanking: domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 2, TeamID: wimborneTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 3, TeamID: dorchesterTownTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 4, TeamID: hamworthyUnitedTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 5, TeamID: bournemouthPoppiesTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 6, TeamID: stJohnsRangersTeamID}},
				{TeamRanking: domain.TeamRanking{Position: 7, TeamID: branksomeUnitedTeamID}},
			}},
		)

		wantErrMsg := "team ids missing from standings rankings: 'BOSTON_RED_SOX', 'EDMONTON_OILERS', 'DARTFORD_TIDDLYWINKS_MASSIF'"
		gotErr := modifier(&domain.MatchWeekResult{})
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})
}

func TestNewMatchWeekResultAgent(t *testing.T) {
	t.Run("passing non-nil repo must succeed", func(t *testing.T) {
		repo := newMatchWeekResultRepo(t, time.Time{})
		if _, err := domain.NewMatchWeekResultAgent(repo); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("passing nil repo must produce the expected error", func(t *testing.T) {
		if _, err := domain.NewMatchWeekResultAgent(nil); !errors.Is(err, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekResultAgent_UpsertBySubmissionID(t *testing.T) {
	t.Cleanup(truncate)

	seedCreatedAt := testDate.Add(-24 * time.Hour)
	seed := seedMatchWeekResult(t, generateMatchWeekResult(t, parseUUID(t, uuidAll1s), modifierSummaries, seedCreatedAt))

	ctx := context.Background()

	t.Run("upsert result that does not exist by submission id should be inserted", func(t *testing.T) {
		repoDate := testDate // createdAt date to insert new submission with
		repo := newMatchWeekResultRepo(t, repoDate)

		agent, err := domain.NewMatchWeekResultAgent(repo)
		if err != nil {
			t.Fatal(err)
		}

		upsertID := parseUUID(t, uuidAll2s)
		toUpsert := generateMatchWeekResult(t, upsertID, altModifierSummaries, time.Time{}) // will not be found by submission id, so should insert a new entry

		wantUpserted := cloneMatchWeekResult(toUpsert) // capture state prior to upsert
		wantUpserted.CreatedAt = repoDate              // should be overridden on insert

		if err := agent.UpsertBySubmissionID(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that seed still exists
		gotSeed, err := repo.GetBySubmissionID(ctx, seed.MatchWeekSubmissionID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "seeded match week submission", seed, gotSeed)

		// ensure that submission was inserted
		gotUpserted, err := repo.GetBySubmissionID(ctx, toUpsert.MatchWeekSubmissionID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("upsert result that exists by submission id should be updated", func(t *testing.T) {
		repoDate := testDate // updatedAt date to update existing submission with
		repo := newMatchWeekResultRepo(t, repoDate)

		agent, err := domain.NewMatchWeekResultAgent(repo)
		if err != nil {
			t.Fatal(err)
		}

		toUpsert := cloneMatchWeekResult(seed) // only change team rankings so will be found by legacy id and match week number
		toUpsert.TeamRankings = []domain.ResultTeamRanking{
			{TeamRanking: domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID}},
		}

		wantUpserted := cloneMatchWeekResult(toUpsert) // capture state prior to upsert
		wantUpserted.UpdatedAt = &repoDate             // should be overridden on update

		if err := agent.UpsertBySubmissionID(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that submission was updated
		gotUpserted, err := repo.GetBySubmissionID(ctx, seed.MatchWeekSubmissionID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("failure to get by submission id must return the expected error", func(t *testing.T) {
		badDB, err := sql.Open("mysql", "connectionString/dbName")
		if err != nil {
			t.Fatal(badDB)
		}

		repo, err := mysqldb.NewMatchWeekResultRepo(badDB, nil)
		if err != nil {
			t.Fatal(err)
		}

		agent, err := domain.NewMatchWeekResultAgent(repo)
		if err != nil {
			t.Fatal(err)
		}

		mwResult := generateMatchWeekResult(t, uuid.UUID{}, modifierSummaries, time.Time{})

		// db will return error on first operation
		wantErrMsg := "cannot get match week result by submission id: default addr for network 'connectionString' unknown"
		gotErr := agent.UpsertBySubmissionID(ctx, mwResult)
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})
}

func newMatchWeekResultRepo(t *testing.T, ts time.Time) *mysqldb.MatchWeekResultRepo {
	t.Helper()

	repo, err := mysqldb.NewMatchWeekResultRepo(db, newTimeFunc(ts))
	if err != nil {
		t.Fatal(err)
	}

	return repo
}

func mustGetUUIDFromString(t *testing.T, input string) uuid.UUID {
	value, err := uuidFromString(input)()
	if err != nil {
		t.Fatal(err)
	}

	return value
}

func uuidFromString(input string) uuidFunc {
	return func() (uuid.UUID, error) {
		return uuid.Parse(input)
	}
}

type uuidFunc func() (uuid.UUID, error)

func randomiseTeamRankings(rankings []domain.TeamRanking) []domain.TeamRanking {
	copied := make([]domain.TeamRanking, 0)
	for _, rank := range rankings {
		copied = append(copied, rank)
	}

	sort.SliceStable(copied, func(i, j int) bool {
		return shouldSwap()
	})

	return copied
}

func randomiseStandingsTeamRankings(rankings []domain.StandingsTeamRanking) []domain.StandingsTeamRanking {
	copied := make([]domain.StandingsTeamRanking, 0)
	for _, rank := range rankings {
		copied = append(copied, rank)
	}

	sort.SliceStable(copied, func(i, j int) bool {
		return shouldSwap()
	})

	return copied
}

func generateMatchWeekResult(t *testing.T, submissionID uuid.UUID, modifiers []domain.ModifierSummary, createdAt time.Time) *domain.MatchWeekResult {
	t.Helper()

	// submission id has foreign key restraint
	submission := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, submissionID, testDate))

	return &domain.MatchWeekResult{
		MatchWeekSubmissionID: submission.ID,
		TeamRankings:          resultTeamRankings,
		Score:                 1234,
		Modifiers:             modifiers,
		CreatedAt:             createdAt,
	}
}

func seedMatchWeekResult(t *testing.T, seed *domain.MatchWeekResult) *domain.MatchWeekResult {
	t.Helper()

	repo, err := mysqldb.NewMatchWeekResultRepo(db, newTimeFunc(seed.CreatedAt))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := repo.Insert(ctx, seed); err != nil {
		t.Fatal(err)
	}

	return seed
}

func cloneMatchWeekResult(original *domain.MatchWeekResult) *domain.MatchWeekResult {
	clone := *original
	return &clone
}

func shouldSwap() bool {
	randNum := randomiser.Intn(2) // either 0 or 1
	return randNum == 1
}

func cmpDiff(t *testing.T, description string, want, got interface{}) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch %s (-want, +got): %s", description, diff)
	}
}

func cmpErrorMsg(t *testing.T, wantMsg string, got error) {
	t.Helper()

	if got == nil {
		t.Fatalf("want error msg '%s', got nil", wantMsg)
	}
	cmpDiff(t, "error msg", wantMsg, got.Error())
}
