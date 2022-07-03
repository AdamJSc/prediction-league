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
	for _, tableName := range []string{"mw_submission", "entry"} {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)); err != nil {
			log.Fatal(err)
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

func TestMatchWeekSubmissionRepo_GetByLegacyIDAndMatchWeekNumber(t *testing.T) {
	t.Cleanup(truncate)

	ctx := context.Background()

	seedID := parseUUID(t, `11111111-1111-1111-1111-111111111111`)
	seed := seedMatchWeekSubmission(t, generateMatchWeekSubmission(t, seedID, testDate))

	repo, err := mysqldb.NewMatchWeekSubmissionRepo(db, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("match week submission that exists must be returned successfully", func(t *testing.T) {
		want := seed
		got, err := repo.GetByLegacyIDAndMatchWeekNumber(ctx, seed.LegacyEntryPredictionID, seed.MatchWeekNumber)
		if err != nil {
			t.Fatal(err)
		}

		cmpDiff(t, "match week submission", want, got)
	})

	t.Run("match week submission that does not exist by legacy id must return the expected error", func(t *testing.T) {
		nonExistentID := parseUUID(t, `22222222-2222-2222-2222-222222222222`)
		_, err := repo.GetByLegacyIDAndMatchWeekNumber(ctx, nonExistentID, seed.MatchWeekNumber)
		if !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want missing db record error, got %+v (%T)", err, err)
		}
	})

	t.Run("match week submission that does not exist by match week number must return the expected error", func(t *testing.T) {
		nonExistentMWNumber := uint16(5678)
		_, err := repo.GetByLegacyIDAndMatchWeekNumber(ctx, seed.LegacyEntryPredictionID, nonExistentMWNumber)
		if !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want missing db record error, got %+v (%T)", err, err)
		}
	})
}

func TestMatchWeekSubmissionRepo_Insert(t *testing.T) {
	t.Skip()
	// TODO: feat - write tests
}

func TestMatchWeekSubmissionRepo_Update(t *testing.T) {
	t.Skip()
	// TODO: feat - write tests
}

func cmpDiff(t *testing.T, description string, want, got interface{}) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch %s (-want, +got): %s", description, diff)
	}
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

func generateEntry() *domain.Entry {
	return &domain.Entry{
		ID: uuid.New(),
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
