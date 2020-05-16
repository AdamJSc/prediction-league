package domain_test

import (
	"github.com/LUSHDigital/core-mage/env"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
	"prediction-league/service/internal/datastore"
	"reflect"
	"testing"
	"time"
)

var (
	db        *coresql.DB
	truncator sqltest.Truncator
	utc       *time.Location
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

	truncator = sqltest.NewTruncator("cockroach", db)

	// run test
	os.Exit(m.Run())
}

// truncate clears our test tables of all previous data between tests
func truncate(t *testing.T) {
	t.Helper()
	truncator.MustTruncateTables(t, "entry")
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
