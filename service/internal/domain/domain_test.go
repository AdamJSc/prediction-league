package domain_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/adapters/mysqldb/sqltypes"
	"prediction-league/service/internal/domain"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	cfg        *domain.Config
	db         *sql.DB
	epr        domain.EntryPredictionRepository
	er         domain.EntryRepository
	rlm        *domain.Realm
	sepr       domain.ScoredEntryPredictionRepository
	sr         domain.StandingsRepository
	testSeason domain.Season
	tpl        *domain.Templates
	tr         domain.TokenRepository
	utc        *time.Location
)

const (
	testRealmName = "TEST_REALM"
	testRealmPIN  = "1234"
)

// TestMain provides a testing bootstrap
func TestMain(m *testing.M) {
	var err error

	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// setup env
	loadTestEnvFromPaths("infra/test.env")

	// load config
	var config struct {
		MySQLURL      string `envconfig:"MYSQL_URL" required:"true"`
		MigrationsURL string `envconfig:"MIGRATIONS_URL" required:"true"`
	}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	utc, err = time.LoadLocation("UTC")
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.NewLogger(os.Stdout, &logger.RealClock{})
	if err != nil {
		log.Fatal(err)
	}

	// setup db connection
	db, err = mysqldb.ConnectAndMigrate(config.MySQLURL, config.MigrationsURL, l)
	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			log.Println("database migration: no changes")
		default:
			log.Fatalf("failed to connect and migrate database: %s", err.Error())
		}
	}

	er, err = mysqldb.NewEntryRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate new entry repo: %s", err.Error())
	}

	epr, err = mysqldb.NewEntryPredictionRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate new entry prediction repo: %s", err.Error())
	}

	sepr, err = mysqldb.NewScoredEntryPredictionRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate new scored entry prediction repo: %s", err.Error())
	}

	sr, err = mysqldb.NewStandingsRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate new standings repo: %s", err.Error())
	}

	tr, err = mysqldb.NewTokenRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate new token repo: %s", err.Error())
	}

	domain.MustInflate()

	// find service path and load templates
	// everything before the last occurrence of "service" within the current working directory path
	dirOfServicePath := wd[:strings.LastIndex(wd, "service")]
	tpl = domain.MustParseTemplates(fmt.Sprintf("%s/service/views", dirOfServicePath))

	rlm = newTestRealm()
	cfg = newTestConfig(*rlm)

	// set testSeason to the first entry within our domain.SeasonsDataStore slice
	keys := reflect.ValueOf(domain.SeasonsDataStore).MapKeys()
	if len(keys) < 1 {
		log.Fatal("need more than one domain.Season")
	}

	testSeason = domain.SeasonsDataStore[keys[0].String()]

	// run test
	os.Exit(m.Run())
}

// loadTestEnvFromPaths tries to load given env files, leaving current environment variables intact.
func loadTestEnvFromPaths(paths ...string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot get working directory: %s", err.Error())
	}
	dir, err := backtrack(wd, "go.mod")
	if err != nil {
		log.Fatalf("cannot find module path: %s", err.Error())
	}
	for _, fpath := range paths {
		fullpath := filepath.Join(dir, fpath)
		if err := godotenv.Load(fullpath); err != nil {
			log.Printf("could not load environment file: %s: skipping...", fullpath)
		}
	}
}

func backtrack(dir, fname string) (string, error) {
	var search = func(dir, fname string) (bool, error) {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return false, err
		}
		for _, f := range files {
			if f.IsDir() || f.Name() != fname {
				continue
			}
			return true, nil
		}
		return false, nil
	}

	if !filepath.IsAbs(dir) {
		return dir, fmt.Errorf("dir must be an absolute file path")
	}
	for {
		found, err := search(dir, fname)
		if err != nil {
			return dir, err
		}
		if !found {
			if dir == "/" {
				return dir, fmt.Errorf("cannot find the file %q in any subsequent directories", fname)
			}
			dir = filepath.Dir(dir)
			continue
		}
		return dir, nil
	}
}

// truncate clears our test tables of all previous data between tests
func truncate(t *testing.T) {
	t.Helper()
	for _, tableName := range []string{"token", "scored_entry_prediction", "entry_prediction", "standings", "entry"} {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)); err != nil {
			t.Fatal(err)
		}
	}
}

// expectedGot is a test failure helper method for two concrete values
func expectedGot(t *testing.T, expectedValue interface{}, gotValue interface{}) {
	t.Helper()
	t.Fatalf("expected %+v, got %+v", expectedValue, gotValue)
}

// expectedTypeOfGot is a test failure helper method for two value types
func expectedTypeOfGot(t *testing.T, expectedValue interface{}, gotValue interface{}) {
	t.Helper()
	t.Fatalf("expected type %+v, got type %+v", reflect.TypeOf(expectedValue), reflect.TypeOf(gotValue))
}

// expectedEmpty is a test failure helper method for an expected empty value
func expectedEmpty(t *testing.T, ref string, gotValue interface{}) {
	t.Helper()
	t.Fatalf("expected empty %s, got %+v", ref, gotValue)
}

// expectedNonEmpty is a test failure helper method for an expected non-empty value
func expectedNonEmpty(t *testing.T, ref string) {
	t.Helper()
	t.Fatalf("expected non-empty %s, got an empty value", ref)
}

// testContextDefault returns a new domain context with default timeout and cancel function for testing purposes
func testContextDefault(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	return testContextWithGuardAttempt(t, testRealmPIN)
}

// testContextWithGuardAttempt provides a wrapper for setting a guard attempt value on a new testContextDefault
func testContextWithGuardAttempt(t *testing.T, attempt string) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := domain.NewContext()

	realm := domain.RealmFromContext(ctx)
	realm.Name = testRealmName
	realm.PIN = testRealmPIN

	domain.GuardFromContext(ctx).SetAttempt(attempt)

	return ctx, cancel
}

// generateTestStandings generates a new Standings entity for use within the testsuite
func generateTestStandings(t *testing.T) domain.Standings {
	t.Helper()

	// generate random ID
	standingsID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	var rankings []domain.RankingWithMeta
	for _, teamID := range testSeason.TeamIDs {
		rwm := domain.NewRankingWithMeta()
		rwm.Ranking = domain.Ranking{ID: teamID}
		rankings = append(rankings, rwm)
	}

	return domain.Standings{
		ID:          standingsID,
		SeasonID:    testSeason.ID,
		RoundNumber: 1,
		Rankings:    rankings,
		Finalised:   true,
		CreatedAt:   time.Now().Truncate(time.Second),
	}
}

// insertStandings inserts a generated Standings entity into the DB for use within the testsuite
func insertStandings(t *testing.T, standings domain.Standings) domain.Standings {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	if err := sr.Insert(ctx, &standings); err != nil {
		t.Fatal(err)
	}

	return standings
}

// updateStandings updates a generated Standings entity in the DB for use within the testsuite
func updateStandings(t *testing.T, standings domain.Standings) domain.Standings {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	if err := sr.Update(ctx, &standings); err != nil {
		t.Fatal(err)
	}

	return standings
}

// generateTestEntry generates a new Entry entity for use within the testsuite
func generateTestEntry(t *testing.T, entrantName, entrantNickname, entrantEmail string) domain.Entry {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	shortCode, err := er.GenerateUniqueShortCode(ctx)
	if err != nil {
		t.Fatal(err)
	}

	paymentMethod := domain.EntryPaymentMethodOther
	paymentRef := "my_payment_ref"

	return domain.Entry{
		ID:               id,
		ShortCode:        shortCode,
		SeasonID:         testSeason.ID,
		RealmName:        testRealmName,
		EntrantName:      entrantName,
		EntrantNickname:  entrantNickname,
		EntrantEmail:     entrantEmail,
		Status:           domain.EntryStatusPending,
		PaymentMethod:    sqltypes.ToNullString(&paymentMethod),
		PaymentRef:       sqltypes.ToNullString(&paymentRef),
		EntryPredictions: nil,
	}
}

// insertEntry insert a generated Entry entity into the DB for use within the testsuite
func insertEntry(t *testing.T, entry domain.Entry) domain.Entry {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	er, err := mysqldb.NewEntryRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	if err := er.Insert(ctx, &entry); err != nil {
		t.Fatal(err)
	}

	return entry
}

// generateTestEntryPrediction generates a new EntryPrediction entity for use within the testsuite
func generateTestEntryPrediction(t *testing.T, entryID uuid.UUID) domain.EntryPrediction {
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	return domain.EntryPrediction{
		ID:       id,
		EntryID:  entryID,
		Rankings: domain.NewRankingCollectionFromIDs(testSeason.TeamIDs),
	}
}

// insertEntryPrediction inserts a generated EntryPrediction entity into the DB for use within the testsuite
func insertEntryPrediction(t *testing.T, entryPrediction domain.EntryPrediction) domain.EntryPrediction {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	if err := epr.Insert(ctx, &entryPrediction); err != nil {
		t.Fatal(err)
	}

	return entryPrediction
}

// generateTestScoredEntryPrediction generates a new ScoredEntryPrediction entity for use within the testsuite
func generateTestScoredEntryPrediction(t *testing.T, entryPredictionID, standingsID uuid.UUID) domain.ScoredEntryPrediction {
	t.Helper()

	return domain.ScoredEntryPrediction{
		EntryPredictionID: entryPredictionID,
		StandingsID:       standingsID,
		Rankings:          domain.NewRankingWithScoreCollectionFromIDs(testSeason.TeamIDs),
		Score:             123,
		CreatedAt:         time.Now().Truncate(time.Second),
	}
}

// insertScoredEntryPrediction inserts a generated ScoredEntryPrediction entity into the DB for use within the testsuite
func insertScoredEntryPrediction(t *testing.T, scoredEntryPrediction domain.ScoredEntryPrediction) domain.ScoredEntryPrediction {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	if err := sepr.Insert(ctx, &scoredEntryPrediction); err != nil {
		t.Fatal(err)
	}

	return scoredEntryPrediction
}

func newTestConfig(r domain.Realm) *domain.Config {
	return &domain.Config{
		Realms: map[string]domain.Realm{
			r.Name: r,
		},
	}
}

func newTestRealm() *domain.Realm {
	return &domain.Realm{
		Name:     "TEST_REALM",
		Origin:   "http://test_realm.org",
		PIN:      "12345",
		SeasonID: testSeason.ID,
		EntryFee: domain.RealmEntryFee{
			Amount: 12.34,
			Label:  "£12.34",
			Breakdown: []string{
				"£12.33 charge",
				"£0.01 processing fee",
			},
		},
		Contact: domain.RealmContact{
			Name:            "Harry R and the PL Team",
			EmailProper:     "hello@world.net",
			EmailSanitised:  "hello (at) world (dot) net",
			EmailDoNotReply: "do_not_reply@world.net",
		},
	}
}
