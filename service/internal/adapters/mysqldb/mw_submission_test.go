package mysqldb_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
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
	db *sql.DB

	teamRankings = []domain.TeamRanking{
		{Position: 1, TeamID: pooleTownTeamID},
		{Position: 2, TeamID: wimborneTownTeamID},
		{Position: 3, TeamID: dorchesterTownTeamID},
		{Position: 4, TeamID: hamworthyUnitedTeamID},
		{Position: 5, TeamID: bournemouthPoppiesTeamID},
		{Position: 6, TeamID: stJohnsRangersTeamID},
		{Position: 7, TeamID: branksomeUnitedTeamID},
	}

	testDate time.Time
	utc      *time.Location
)

func TestMain(m *testing.M) {
	var err error

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatal(err)
	}

	testDate = time.Date(2018, 5, 26, 14, 0, 0, 0, loc)

	projectRootDir := "../../../.."

	// setup env
	mustLoadTestEnvFromPaths(projectRootDir + "/infra/test.env")

	// load config
	var config struct {
		MySQLURL       string `envconfig:"MYSQL_URL" required:"true"`
		MigrationsPath string `envconfig:"MIGRATIONS_PATH" required:"true"`
	}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	utc, err = time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.NewLogger("DEBUG", os.Stdout, &domain.RealClock{})
	if err != nil {
		log.Fatal(err)
	}

	// setup db connection
	migrationsURL := fmt.Sprintf("file://%s/%s", projectRootDir, config.MigrationsPath)
	db, err = mysqldb.ConnectAndMigrate(config.MySQLURL, migrationsURL, l)
	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			log.Println("database migration: no changes")
		default:
			log.Fatalf("failed to connect and migrate database: %s", err.Error())
		}
	}

	// run tests
	os.Exit(m.Run())
}

// truncate clears our test tables of all previous data between tests
func truncate() {
	for _, tableName := range []string{"mw_result_modifier", "mw_result", "mw_submission", "token", "scored_entry_prediction", "entry_prediction", "standings", "entry"} {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)); err != nil {
			log.Fatalf("cannot truncate table '%s': %s", tableName, err.Error())
		}
	}
}

// mustLoadTestEnvFromPaths tries to load given env files, leaving current environment variables intact.
func mustLoadTestEnvFromPaths(paths ...string) {
	for _, p := range paths {
		if err := godotenv.Load(p); err != nil {
			log.Printf("could not load environment file: %s: skipping...", p)
		}
	}
}

func TestNewMatchWeekSubmissionRepo(t *testing.T) {
	t.Run("passing non-nil db must succeed", func(t *testing.T) {
		if _, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("passing nil db must produce the expected error", func(t *testing.T) {
		if _, err := mysqldb.NewMatchWeekSubmissionRepo(nil, nil, nil); !errors.Is(err, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekSubmissionRepo_GetByID(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

	seedID := newUUID(t)
	seed := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, seedID, testDate))

	repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("match week submission that exists must be returned successfully", func(t *testing.T) {
		want := seed
		got, err := repo.GetByID(ctx, seed.ID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week submission", want, got)
	})

	t.Run("match week submission that does not exist must return the expected error", func(t *testing.T) {
		nonExistentID := newUUID(t)
		_, err := repo.GetByID(ctx, nonExistentID)
		if !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want missing db record error, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekSubmissionRepo_GetByEntryIDAndMatchWeekNumber(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

	seed := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, newUUID(t), testDate))

	repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("match week submission that exists must be returned successfully", func(t *testing.T) {
		want := seed
		got, err := repo.GetByEntryIDAndMatchWeekNumber(ctx, seed.EntryID, seed.MatchWeekNumber)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week submission", want, got)
	})

	t.Run("match week submission that does not exist by legacy id must return the expected error", func(t *testing.T) {
		nonExistentID := newUUID(t)
		_, err := repo.GetByEntryIDAndMatchWeekNumber(ctx, nonExistentID, seed.MatchWeekNumber)
		if !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want missing db record error, got %+v (%T)", err, err)
		}
	})

	t.Run("match week submission that does not exist by match week number must return the expected error", func(t *testing.T) {
		nonExistentMWNumber := uint16(5678)
		_, err := repo.GetByEntryIDAndMatchWeekNumber(ctx, seed.EntryID, nonExistentMWNumber)
		if !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want missing db record error, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekSubmissionRepo_Insert(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()
	insertID := newUUID(t)
	createdAt := testDate

	t.Run("passing nil match week submission must generate no error", func(t *testing.T) {
		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := repo.Insert(ctx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("valid match week submission must be inserted successfully", func(t *testing.T) {
		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, newUUIDFunc(insertID), newTimeFunc(createdAt))
		if err != nil {
			t.Fatal(err)
		}

		submission := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{}) // empty id and createdAt timestamp

		want := cloneMatchWeekSubmission(submission) // capture state before insert
		want.ID = insertID                           // should be overridden on insert
		want.CreatedAt = createdAt                   // should be overridden on insert

		if err := repo.Insert(ctx, submission); err != nil {
			t.Fatal(err)
		}

		got, err := repo.GetByID(ctx, insertID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week submission", want, got)
		cmpDiff(t, "id on entity", want.ID, submission.ID)
		cmpDiff(t, "created date on entity", want.CreatedAt, submission.CreatedAt)

		// inserting same submission again must return the expected error
		wantErrType := domain.DuplicateDBRecordError{}
		gotErr := repo.Insert(ctx, got)
		if !errors.As(gotErr, &wantErrType) {
			t.Fatalf("want error type %T, got %T", wantErrType, gotErr)
		}
	})

	t.Run("error generating uuid must return the expected error", func(t *testing.T) {
		uuidFn := func() (uuid.UUID, error) {
			return uuid.UUID{}, errors.New("sad times :'(")
		}

		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, uuidFn, newTimeFunc(testDate))
		if err != nil {
			t.Fatal(err)
		}

		submission := generateMatchWeekSubmission(t, uuid.UUID{}, time.Time{})

		wantErrMsg := "cannot get uuid: sad times :'("
		gotErr := repo.Insert(ctx, submission)
		cmpErrorMsg(t, wantErrMsg, gotErr)
	})
}

func TestMatchWeekSubmissionRepo_Update(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

	t.Run("passing nil match week submission must generate no error", func(t *testing.T) {
		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := repo.Update(ctx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("updating match week submission that exists must be successful", func(t *testing.T) {
		createdAt := testDate

		updatedAt := createdAt.Add(time.Second)
		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, newTimeFunc(updatedAt))
		if err != nil {
			t.Fatal(err)
		}

		seed1 := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, newUUID(t), createdAt))
		seed2 := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, newUUID(t), createdAt))

		changedSeed := &domain.MatchWeekSubmission{
			ID:                      seed1.ID,
			EntryID:                 seed2.EntryID,
			MatchWeekNumber:         9999,
			TeamRankings:            []domain.TeamRanking{{Position: 9999}},
			LegacyEntryPredictionID: seed2.LegacyEntryPredictionID,
		}

		want := cloneMatchWeekSubmission(changedSeed) // capture state before update
		want.CreatedAt = seed1.CreatedAt              // should not be overridden on update
		want.UpdatedAt = &updatedAt                   // should be overridden on update

		if err := repo.Update(ctx, changedSeed); err != nil {
			t.Fatal(err)
		}

		got, err := repo.GetByID(ctx, seed1.ID)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week submission", want, got)
		cmpDiff(t, "updated date on entity", want.UpdatedAt, changedSeed.UpdatedAt)
	})

	t.Run("updating match week submission that does not exist must return the expected error", func(t *testing.T) {
		repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		submission := &domain.MatchWeekSubmission{ID: newUUID(t)}

		wantErrType := domain.MissingDBRecordError{}
		gotErr := repo.Update(ctx, submission)
		if !errors.As(gotErr, &wantErrType) {
			t.Fatalf("want error type %T, got %T", wantErrType, gotErr)
		}
	})
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
