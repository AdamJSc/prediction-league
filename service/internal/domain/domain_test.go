package domain_test

import (
	"context"
	"fmt"
	"github.com/LUSHDigital/core-mage/env"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltest"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/repositories/repofac"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	db         *coresql.DB
	truncator  sqltest.Truncator
	utc        *time.Location
	templates  *domain.Templates
	testSeason domain.Season
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
	env.LoadTest(m, "infra/test.env")

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

	// setup db connection
	db = coresql.MustOpen("mysql", config.MySQLURL)
	driver, _ := mysql.WithInstance(db.DB, &mysql.Config{})
	mig, _ := migrate.NewWithDatabaseInstance(
		config.MigrationsURL,
		"mysql",
		driver,
	)
	coresql.MustMigrateUp(mig)

	domain.MustInflate()

	// find service path and load templates
	// everything before the last occurrence of "service" within the current working directory path
	dirOfServicePath := wd[:strings.LastIndex(wd, "service")]
	templates = domain.MustParseTemplates(fmt.Sprintf("%s/service/views", dirOfServicePath))

	// set testSeason to the first entry within our domain.SeasonsDataStore slice
	keys := reflect.ValueOf(domain.SeasonsDataStore).MapKeys()
	if len(keys) < 1 {
		log.Fatal("need more than one domain.Season")
	}

	testSeason = domain.SeasonsDataStore[keys[0].String()]

	// run test
	os.Exit(m.Run())
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
	standingsID, err := uuid.NewV4()
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

	if err := repofac.NewStandingsDatabaseRepository(db).Insert(ctx, &standings); err != nil {
		t.Fatal(err)
	}

	return standings
}

// updateStandings updates a generated Standings entity in the DB for use within the testsuite
func updateStandings(t *testing.T, standings domain.Standings) domain.Standings {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	if err := repofac.NewStandingsDatabaseRepository(db).Update(ctx, &standings); err != nil {
		t.Fatal(err)
	}

	return standings
}

// generateTestEntry generates a new Entry entity for use within the testsuite
func generateTestEntry(t *testing.T, entrantName, entrantNickname, entrantEmail string) domain.Entry {
	t.Helper()

	ctx, cancel := testContextDefault(t)
	defer cancel()

	id, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	entryRepo := repofac.NewEntryDatabaseRepository(db)
	shortCode, err := entryRepo.GenerateUniqueShortCode(ctx)
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

	if err := repofac.NewEntryDatabaseRepository(db).Insert(ctx, &entry); err != nil {
		t.Fatal(err)
	}

	return entry
}

// generateTestEntryPrediction generates a new EntryPrediction entity for use within the testsuite
func generateTestEntryPrediction(t *testing.T, entryID uuid.UUID) domain.EntryPrediction {
	id, err := uuid.NewV4()
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

	if err := repofac.NewEntryPredictionDatabaseRepository(db).Insert(ctx, &entryPrediction); err != nil {
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

	if err := repofac.NewScoredEntryPredictionDatabaseRepository(db).Insert(ctx, &scoredEntryPrediction); err != nil {
		t.Fatal(err)
	}

	return scoredEntryPrediction
}

type testInjector struct {
	config    domain.Config
	queue     chan domain.Email
	templates *domain.Templates
	er        domain.EntryRepository
	epr       domain.EntryPredictionRepository
	sr        domain.StandingsRepository
	sepr      domain.ScoredEntryPredictionRepository
	tr        domain.TokenRepository
}

func (t *testInjector) Config() domain.Config             { return t.config }
func (t *testInjector) EntryRepo() domain.EntryRepository { return t.er }
func (t *testInjector) EntryPredictionRepo() domain.EntryPredictionRepository {
	return t.epr
}
func (t *testInjector) StandingsRepo() domain.StandingsRepository { return t.sr }
func (t *testInjector) ScoredEntryPredictionRepo() domain.ScoredEntryPredictionRepository {
	return t.sepr
}
func (t *testInjector) TokenRepo() domain.TokenRepository {
	return t.tr
}
func (t *testInjector) EmailQueue() chan domain.Email { return t.queue }
func (t *testInjector) Template() *domain.Templates   { return t.templates }

func newTestInjector(t *testing.T, r domain.Realm, tpl *domain.Templates, db coresql.Agent) *testInjector {
	return &testInjector{
		config:    newTestConfig(t, r),
		queue:     make(chan domain.Email, 1),
		templates: tpl,
		er:        repofac.NewEntryDatabaseRepository(db),
		epr:       repofac.NewEntryPredictionDatabaseRepository(db),
		sr:        repofac.NewStandingsDatabaseRepository(db),
		sepr:      repofac.NewScoredEntryPredictionDatabaseRepository(db),
		tr:        repofac.NewTokenDatabaseRepository(db),
	}
}

func newTestConfig(t *testing.T, r domain.Realm) domain.Config {
	t.Helper()

	return domain.Config{
		Realms: map[string]domain.Realm{
			r.Name: r,
		},
	}
}

func newTestRealm(t *testing.T) domain.Realm {
	t.Helper()

	realm := domain.Realm{
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
	}

	realm.Contact.Name = "Harry R and the PL Team"
	realm.Contact.EmailProper = "hello@world.net"
	realm.Contact.EmailSanitised = "hello (at) world (dot) net"
	realm.Contact.EmailDoNotReply = "do_not_reply@world.net"

	return realm
}
