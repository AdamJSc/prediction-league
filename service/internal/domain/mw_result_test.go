package domain_test

import (
	"errors"
	"prediction-league/service/internal/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
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
