package domain_test

import (
	"errors"
	"math/rand"
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

var randomiser = rand.New(rand.NewSource(time.Now().UnixNano()))

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
	t.Skip()
	// TODO: feat - write agent constructor tests
}

func TestMatchWeekResultAgent_UpsertBySubmissionID(t *testing.T) {
	t.Skip()
	// TODO: feat - write agent method tests
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
