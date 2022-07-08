package mysqldb_test

import (
	"context"
	"errors"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	resultTeamRankings = []domain.ResultTeamRanking{
		{TeamRanking: domain.TeamRanking{Position: 1, TeamID: pooleTownTeamID}, StandingsPos: 7, Hit: 6},
		{TeamRanking: domain.TeamRanking{Position: 2, TeamID: wimborneTownTeamID}, StandingsPos: 6, Hit: 4},
		{TeamRanking: domain.TeamRanking{Position: 3, TeamID: dorchesterTownTeamID}, StandingsPos: 5, Hit: 2},
		{TeamRanking: domain.TeamRanking{Position: 4, TeamID: hamworthyUnitedTeamID}, StandingsPos: 4, Hit: 0},
		{TeamRanking: domain.TeamRanking{Position: 5, TeamID: bournemouthPoppiesTeamID}, StandingsPos: 3, Hit: 2},
		{TeamRanking: domain.TeamRanking{Position: 6, TeamID: stJohnsRangersTeamID}, StandingsPos: 2, Hit: 4},
		{TeamRanking: domain.TeamRanking{Position: 7, TeamID: branksomeUnitedTeamID}, StandingsPos: 1, Hit: 6},
	}

	modifierSummaries = []domain.ModifierSummary{
		{Code: domain.BaseScoreModifierCode, Value: 100},
		{Code: domain.TeamRankingsHitModifierCode, Value: 88},
	}
)

func TestNewMatchWeekResultRepo(t *testing.T) {
	t.Run("passing non-nil db must succeed", func(t *testing.T) {
		if _, err := mysqldb.NewMatchWeekResultRepo(db, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("passing nil db must produce the expected error", func(t *testing.T) {
		if _, err := mysqldb.NewMatchWeekResultRepo(nil, nil); !errors.Is(err, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekResultRepo_GetBySubmissionID(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

	seedID := parseUUID(t, uuidAll1s)
	seed := seedMatchWeekResult(t, generateMatchWeekResult(t, seedID, modifierSummaries, testDate))

	repo, err := mysqldb.NewMatchWeekResultRepo(db, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("match week result that exists must be returned successfully", func(t *testing.T) {
		want := seed
		got, err := repo.GetBySubmissionID(ctx, seed.MatchWeekSubmissionID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", want, got)
	})

	t.Run("match week result that does not exist by submission id must return the expected error", func(t *testing.T) {
		nonExistentID := parseUUID(t, uuidAll2s)
		_, err := repo.GetBySubmissionID(ctx, nonExistentID)
		if !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want missing db record error, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekResultRepo_Insert(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

	t.Run("passing nil match week result must generate no error", func(t *testing.T) {
		repo, err := mysqldb.NewMatchWeekResultRepo(db, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := repo.Insert(ctx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("valid match week result must be inserted successfully", func(t *testing.T) {
		mwResult := generateMatchWeekResult(t, parseUUID(t, uuidAll1s), modifierSummaries, time.Time{}) // empty createdAt timestamp
		initialMWResult := cloneMatchWeekResult(mwResult)                                               // capture state before insert

		createdAt := testDate
		repo, err := mysqldb.NewMatchWeekResultRepo(db, newTimeFunc(createdAt))
		if err != nil {
			t.Fatal(err)
		}

		if err := repo.Insert(ctx, mwResult); err != nil {
			t.Fatal(err)
		}

		want := initialMWResult
		want.CreatedAt = createdAt // should be overridden on insert

		got, err := repo.GetBySubmissionID(ctx, initialMWResult.MatchWeekSubmissionID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week result", want, got)

		// inserting same mw result again must return the expected error
		wantErrType := domain.DuplicateDBRecordError{}
		gotErr := repo.Insert(ctx, got)
		if !errors.As(gotErr, &wantErrType) {
			t.Fatalf("want error type %T, got %T", wantErrType, gotErr)
		}
	})
}

func TestMatchWeekResultRepo_Update(t *testing.T) {
	t.Skip()
	// TODO: feat - write repo method tests
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
