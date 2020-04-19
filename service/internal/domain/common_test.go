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
	"prediction-league/service/internal/domain"
	"reflect"
	"testing"
)

var (
	db        *coresql.DB
	truncator sqltest.Truncator
)

type injector struct {
	db coresql.Agent
}

func (i injector) MySQL() coresql.Agent { return i.db }

func TestMain(m *testing.M) {
	// setup env
	env.LoadTest(m, "infra/test.env")
	config := struct {
		MySQLURL      string `envconfig:"MYSQL_URL" required:"true"`
		MigrationsURL string `envconfig:"MIGRATIONS_URL" required:"true"`
	}{}
	if err := envconfig.Process("", &config); err != nil {
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

	domain.RegisterCustomValidators()

	truncator = sqltest.NewTruncator("cockroach", db)

	// run test
	os.Exit(m.Run())
}

func truncate(t *testing.T) {
	t.Helper()
	truncator.MustTruncateTables(t, "season", "entry")
}

func expectedGot(t *testing.T, expectedValue interface{}, gotValue interface{}) {
	t.Helper()
	t.Fatalf("expected %+v, got %+v", expectedValue, gotValue)
}

func expectedTypeOfGot(t *testing.T, expectedValue interface{}, gotValue interface{}) {
	t.Helper()
	t.Fatalf("expected type %+v, got type %+v", reflect.TypeOf(expectedValue), reflect.TypeOf(gotValue))
}

func expectedEmpty(t *testing.T, expectedType string, gotValue interface{}) {
	t.Helper()
	t.Fatalf("expected empty %s, got %+v", expectedType, gotValue)
}

func expectedNonEmpty(t *testing.T, fieldName string) {
	t.Helper()
	t.Fatalf("expected non-empty field %s, got an empty value", fieldName)
}
