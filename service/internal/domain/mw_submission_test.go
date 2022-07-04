package domain_test

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	teamRankings = []domain.TeamRanking{
		{Position: 1, TeamID: pooleTownTeamID},
		{Position: 2, TeamID: wimborneTownTeamID},
		{Position: 3, TeamID: dorchesterTownTeamID},
		{Position: 4, TeamID: hamworthyUnitedTeamID},
		{Position: 5, TeamID: bournemouthPoppiesTeamID},
		{Position: 6, TeamID: stJohnsRangersTeamID},
		{Position: 7, TeamID: branksomeUnitedTeamID},
	}
)

func TestNewMatchWeekSubmissionAgent(t *testing.T) {
	t.Run("passing non-nil repo must succeed", func(t *testing.T) {
		repo := newMatchWeekSubmissionRepo(t, uuid.UUID{}, time.Time{})
		if _, err := domain.NewMatchWeekSubmissionAgent(repo); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("passing nil repo must produce the expected error", func(t *testing.T) {
		if _, err := domain.NewMatchWeekSubmissionAgent(nil); !errors.Is(err, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekSubmissionAgent_UpsertByLegacy(t *testing.T) {
	t.Cleanup(truncate)

	seedID := parseUUID(t, uuidAll1s)
	seedCreatedAt := testDate.Add(-24 * time.Hour)
	seed := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, seedID, seedCreatedAt))

	ctx := context.Background()

	t.Run("upsert submission that does not exist by legacy id should be inserted", func(t *testing.T) {
		repoID := uuid.New() // id to insert new submission with
		repoDate := testDate // createdAt date to insert new submission with
		repo := newMatchWeekSubmissionRepo(t, repoID, repoDate)

		agent, err := domain.NewMatchWeekSubmissionAgent(repo)
		if err != nil {
			t.Fatal(err)
		}

		toUpsert := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{})
		toUpsert.MatchWeekNumber = seed.MatchWeekNumber
		toUpsert.LegacyEntryPredictionID = uuid.New() // will not be found by legacy id, so should insert a new entry

		wantUpserted := cloneMatchWeekSubmission(toUpsert) // capture state prior to upsert
		wantUpserted.ID = repoID                           // should be overridden on insert
		wantUpserted.CreatedAt = repoDate                  // should be overridden on insert

		if err := agent.UpsertByLegacy(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that seed still exists
		wantSeed := cloneMatchWeekSubmission(seed)
		gotSeed, err := repo.GetByID(ctx, wantSeed.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "seeded match week submission", wantSeed, gotSeed)

		// ensure that submission was inserted
		gotUpserted, err := repo.GetByID(ctx, wantUpserted.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("upsert submission that does not exist by match week number should be inserted", func(t *testing.T) {
		repoID := uuid.New() // id to insert new submission with
		repoDate := testDate // createdAt date to insert new submission with
		repo := newMatchWeekSubmissionRepo(t, repoID, repoDate)

		agent, err := domain.NewMatchWeekSubmissionAgent(repo)
		if err != nil {
			t.Fatal(err)
		}

		toUpsert := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{})
		toUpsert.MatchWeekNumber = 9999 // will not be found by match week number, so should insert a new entry
		toUpsert.LegacyEntryPredictionID = seed.LegacyEntryPredictionID

		wantUpserted := cloneMatchWeekSubmission(toUpsert) // capture state prior to upsert
		wantUpserted.ID = repoID                           // should be overridden on insert
		wantUpserted.CreatedAt = repoDate                  // should be overridden on insert

		if err := agent.UpsertByLegacy(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that seed still exists
		wantSeed := cloneMatchWeekSubmission(seed)
		gotSeed, err := repo.GetByID(ctx, wantSeed.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "seeded match week submission", wantSeed, gotSeed)

		// ensure that submission was inserted
		gotUpserted, err := repo.GetByID(ctx, wantUpserted.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("upsert submission that exists by legacy id and match week number should be updated", func(t *testing.T) {
		t.Skip()
		// TODO: feat - write test
	})

	t.Run("failure when matching by legacy id and match week number must return the expected error", func(t *testing.T) {
		t.Skip()
		// TODO: feat - write test
	})

	t.Run("insert failure must return the expected error", func(t *testing.T) {
		t.Skip()
		// TODO: feat - write test
	})

	t.Run("update failure must return the expected error", func(t *testing.T) {
		t.Skip()
		// TODO: feat - write test
	})
}

func newMatchWeekSubmissionRepo(t *testing.T, id uuid.UUID, ts time.Time) *mysqldb.MatchWeekSubmissionRepo {
	t.Helper()

	repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, newUUIDFunc(id), newTimeFunc(ts))
	if err != nil {
		t.Fatal(err)
	}

	return repo
}

func generateMatchWeekSubmission(t *testing.T, id uuid.UUID, createdAt time.Time) *domain.MatchWeekSubmission {
	t.Helper()

	entry := seedEntry(t, generateEntry())

	seedLegacyEntryPredictionID, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	return &domain.MatchWeekSubmission{
		ID:                      id,
		EntryID:                 entry.ID,
		MatchWeekNumber:         1234,
		TeamRankings:            teamRankings,
		LegacyEntryPredictionID: seedLegacyEntryPredictionID,
		CreatedAt:               createdAt,
	}
}

func seedMatchWeekSubmission(t *testing.T, seed *domain.MatchWeekSubmission) *domain.MatchWeekSubmission {
	t.Helper()

	repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, newUUIDFunc(seed.ID), newTimeFunc(seed.CreatedAt))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := repo.Insert(ctx, seed); err != nil {
		t.Fatal(err)
	}

	return seed
}

func cloneMatchWeekSubmission(original *domain.MatchWeekSubmission) *domain.MatchWeekSubmission {
	clone := *original
	return &clone
}

func generateEntry() *domain.Entry {
	id := uuid.New()

	return &domain.Entry{
		ID:              id,
		EntrantNickname: fmt.Sprintf("%s_nickname", id),
		EntrantEmail:    fmt.Sprintf("%s@seeder.com", id),
	}
}

func seedEntry(t *testing.T, seed *domain.Entry) *domain.Entry {
	t.Helper()

	repo, err := mysqldb.NewEntryRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := repo.Insert(ctx, seed); err != nil {
		t.Fatal(err)
	}

	return seed
}

func parseUUID(t *testing.T, id string) uuid.UUID {
	t.Helper()

	val, err := uuid.Parse(id)
	if err != nil {
		t.Fatal(err)
	}

	return val
}

func newUUIDFunc(id uuid.UUID) func() (uuid.UUID, error) {
	return func() (uuid.UUID, error) {
		return id, nil
	}
}

func newTimeFunc(ts time.Time) func() time.Time {
	return func() time.Time {
		return ts
	}
}
