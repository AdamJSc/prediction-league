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
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"reflect"
	"testing"
	"time"
)

var (
	db         *coresql.DB
	truncator  sqltest.Truncator
	utc        *time.Location
	testSeason models.Season
)

const (
	testRealmName = "TEST_REALM"
	testRealmPIN  = "1234"
)

// injector can be passed to Agents as our AgentInjectors for testing
type injector struct {
	db coresql.Agent
}

func (i injector) MySQL() coresql.Agent { return i.db }

// TestMain provides a testing bootstrap
func TestMain(m *testing.M) {
	// setup env
	env.LoadTest(m, "infra/test.env")

	// load config
	config := struct {
		MySQLURL      string `envconfig:"MYSQL_URL" required:"true"`
		MigrationsURL string `envconfig:"MIGRATIONS_URL" required:"true"`
	}{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	var err error
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

	datastore.MustInflate()

	// set testSeason to the first entry within our datastore.Seasons slice
	keys := reflect.ValueOf(datastore.Seasons).MapKeys()
	if len(keys) < 1 {
		log.Fatal("need more than one datastore.Season")
	}

	testSeason = datastore.Seasons[keys[0].String()]

	// run test
	os.Exit(m.Run())
}

// truncate clears our test tables of all previous data between tests
func truncate(t *testing.T) {
	t.Helper()
	for _, tableName := range []string{"scored_entry_selection", "entry_selection", "standings", "entry"} {
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

// testContext returns a new context with default timeout and cancel function for testing purposes
func testContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	return context.WithTimeout(context.Background(), 5*time.Second)
}

// testDomainContext returns a new domain context with default timeout and cancel function for testing purposes
func testDomainContext(t *testing.T) (domain.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := testContext(t)

	var domainCtx = domain.Context{Context: ctx}
	domainCtx.Realm.Name = testRealmName
	domainCtx.Realm.PIN = testRealmPIN
	domainCtx.Guard.SetAttempt(testRealmPIN)

	return domainCtx, cancel
}

// generateTestStandings generates a new Standings entity for use within the testsuite
func generateTestStandings(t *testing.T) models.Standings {
	t.Helper()

	// get first season
	key := reflect.ValueOf(datastore.Seasons).MapKeys()[0].String()
	season := datastore.Seasons[key]

	id, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	var rankings = []models.RankingWithMeta{{
		Ranking: models.Ranking{ID: "hello"},
	}, {
		Ranking: models.Ranking{ID: "world"},
	}}

	return models.Standings{
		ID:          id,
		SeasonID:    season.ID,
		RoundNumber: 1,
		Rankings:    rankings,
	}
}

// insertStandings inserts a generated Standings entity into the DB for use within the testsuite
func insertStandings(t *testing.T, standings models.Standings) models.Standings {
	t.Helper()

	ctx, cancel := testContext(t)
	defer cancel()

	if err := repositories.NewStandingsDatabaseRepository(db).Insert(ctx, &standings); err != nil {
		t.Fatal(err)
	}

	return standings
}

// generateTestEntry generates a new Entry entity for use within the testsuite
func generateTestEntry(t *testing.T, entrantName, entrantNickname, entrantEmail string) models.Entry {
	t.Helper()

	ctx, cancel := testContext(t)
	defer cancel()

	id, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	shortCode, err := domain.GenerateUniqueShortCode(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	paymentMethod := models.EntryPaymentMethodOther
	paymentRef := "my_payment_ref"

	return models.Entry{
		ID:              id,
		ShortCode:       shortCode,
		SeasonID:        testSeason.ID,
		RealmName:       testRealmName,
		EntrantName:     entrantName,
		EntrantNickname: entrantNickname,
		EntrantEmail:    entrantEmail,
		Status:          models.EntryStatusPending,
		PaymentMethod:   sqltypes.ToNullString(&paymentMethod),
		PaymentRef:      sqltypes.ToNullString(&paymentRef),
		EntrySelections: nil,
	}
}

// insertEntry insert a generated Entry entity into the DB for use within the testsuite
func insertEntry(t *testing.T, entry models.Entry) models.Entry {
	t.Helper()

	ctx, cancel := testContext(t)
	defer cancel()

	if err := repositories.NewEntryDatabaseRepository(db).Insert(ctx, &entry); err != nil {
		t.Fatal(err)
	}

	return entry
}

// generateTestEntrySelection generates a new EntrySelection entity for use within the testsuite
func generateTestEntrySelection(t *testing.T, entryID uuid.UUID) models.EntrySelection {
	id, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	return models.EntrySelection{
		ID:       id,
		EntryID:  entryID,
		Rankings: models.NewRankingCollectionFromIDs(testSeason.TeamIDs),
	}
}

// insertEntrySelection inserts a generated EntrySelection entity into the DB for use within the testsuite
func insertEntrySelection(t *testing.T, entrySelection models.EntrySelection) models.EntrySelection {
	t.Helper()

	ctx, cancel := testContext(t)
	defer cancel()

	if err := repositories.NewEntrySelectionDatabaseRepository(db).Insert(ctx, &entrySelection); err != nil {
		t.Fatal(err)
	}

	return entrySelection
}

// generateTestScoredEntrySelection generates a new ScoredEntrySelection entity for use within the testsuite
func generateTestScoredEntrySelection(t *testing.T, entrySelectionID, standingsID uuid.UUID) models.ScoredEntrySelection {
	return models.ScoredEntrySelection{
		EntrySelectionID: entrySelectionID,
		StandingsID:      standingsID,
		Rankings:         models.NewRankingWithScoreCollectionFromIDs(testSeason.TeamIDs),
		Score:            123,
	}
}

// insertScoredEntrySelection inserts a generated ScoredEntrySelection entity into the DB for use within the testsuite
func insertScoredEntrySelection(t *testing.T, scoredEntrySelection models.ScoredEntrySelection) models.ScoredEntrySelection {
	t.Helper()

	ctx, cancel := testContext(t)
	defer cancel()

	if err := repositories.NewScoredEntrySelectionDatabaseRepository(db).Insert(ctx, &scoredEntrySelection); err != nil {
		t.Fatal(err)
	}

	return scoredEntrySelection
}
