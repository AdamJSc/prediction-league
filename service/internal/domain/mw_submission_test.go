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

	seedID := newUUID(t)
	seedCreatedAt := testDate.Add(-24 * time.Hour)
	seed := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, seedID, seedCreatedAt))

	noSubsSeededEntry := seedEntry(t, generateEntry()) // seed entry that has no submissions attached so that we can reference an existent entry that doesn't have any submissions

	ctx := context.Background()

	t.Run("upsert submission that does not exist by entry id should be inserted", func(t *testing.T) {
		repoID := uuid.New() // id to insert new submission with
		repoDate := testDate // createdAt date to insert new submission with
		repo := newMatchWeekSubmissionRepo(t, repoID, repoDate)
		agent := newMatchWeekSubmissionAgent(t, repo)

		toUpsert := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{})
		toUpsert.MatchWeekNumber = seed.MatchWeekNumber
		toUpsert.EntryID = noSubsSeededEntry.ID // will not be found by entry id, so should be inserted as a new submission

		wantUpserted := cloneMatchWeekSubmission(toUpsert) // capture state prior to upsert
		wantUpserted.ID = repoID                           // should be overridden on insert
		wantUpserted.CreatedAt = repoDate                  // should be overridden on insert

		if err := agent.UpsertByLegacy(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that seed still exists
		wantSeed := cloneMatchWeekSubmission(seed)
		gotSeed, err := repo.GetByID(ctx, seed.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "seeded match week submission", wantSeed, gotSeed)

		// ensure that submission was inserted
		gotUpserted, err := repo.GetByID(ctx, toUpsert.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("upsert submission that does not exist by match week number should be inserted", func(t *testing.T) {
		repoID := uuid.New() // id to insert new submission with
		repoDate := testDate // createdAt date to insert new submission with
		repo := newMatchWeekSubmissionRepo(t, repoID, repoDate)
		agent := newMatchWeekSubmissionAgent(t, repo)

		toUpsert := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{})
		toUpsert.MatchWeekNumber = 9999 // will not be found by match week number, so should be inserted as a new submission
		toUpsert.EntryID = seed.EntryID

		wantUpserted := cloneMatchWeekSubmission(toUpsert) // capture state prior to upsert
		wantUpserted.ID = repoID                           // should be overridden on insert
		wantUpserted.CreatedAt = repoDate                  // should be overridden on insert

		if err := agent.UpsertByLegacy(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that seed still exists
		wantSeed := cloneMatchWeekSubmission(seed)
		gotSeed, err := repo.GetByID(ctx, seed.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "seeded match week submission", wantSeed, gotSeed)

		// ensure that submission was inserted
		gotUpserted, err := repo.GetByID(ctx, toUpsert.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("upsert submission that exists by entry id and match week number should be updated", func(t *testing.T) {
		repoDate := testDate // updatedAt date to update existing submission with
		repo := newMatchWeekSubmissionRepo(t, uuid.UUID{}, repoDate)
		agent := newMatchWeekSubmissionAgent(t, repo)

		toUpsert := cloneMatchWeekSubmission(seed) // only change team rankings so will be found by legacy id and match week number
		toUpsert.TeamRankings = []domain.TeamRanking{
			{Position: 1, TeamID: pooleTownTeamID},
		}

		wantUpserted := cloneMatchWeekSubmission(toUpsert) // capture state prior to upsert
		wantUpserted.UpdatedAt = &repoDate                 // should be overridden on update

		if err := agent.UpsertByLegacy(ctx, toUpsert); err != nil {
			t.Fatal(err)
		}

		// ensure that submission was updated
		gotUpserted, err := repo.GetByID(ctx, seed.ID)
		if err != nil {
			t.Fatal(err)
		}
		cmpDiff(t, "upserted match week submission", wantUpserted, gotUpserted)
	})

	t.Run("insert failure must return the expected error", func(t *testing.T) {
		idFn := func() (uuid.UUID, error) {
			return uuid.UUID{}, errors.New("sad times :'(")
		}
		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, idFn, nil)
		if err != nil {
			t.Fatal(err)
		}
		agent := newMatchWeekSubmissionAgent(t, repo)

		submission := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{})

		// new submission will be inserted but uuid function will return error
		wantErrMsg := "cannot insert submission: cannot get uuid: sad times :'("
		gotErr := agent.UpsertByLegacy(ctx, submission)
		cmpErrorMsg(t, wantErrMsg, gotErr)
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

func newMatchWeekSubmissionAgent(t *testing.T, repo domain.MatchWeekSubmissionRepository) *domain.MatchWeekSubmissionAgent {
	t.Helper()

	agent, err := domain.NewMatchWeekSubmissionAgent(repo)
	if err != nil {
		t.Fatal(err)
	}

	return agent
}

func generateMatchWeekSubmission(t *testing.T, id uuid.UUID, createdAt time.Time) *domain.MatchWeekSubmission {
	t.Helper()

	entry := seedEntry(t, generateEntry()) // entry id has foreign key restraint

	seedLegacyEntryPredictionID, err := uuid.NewUUID() // no key restraint, generate new value
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

func newUUID(t *testing.T) uuid.UUID {
	t.Helper()

	val, err := uuid.NewUUID()
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
