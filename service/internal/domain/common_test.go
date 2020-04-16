package domain_test

import (
	"github.com/LUSHDigital/core-mage/env"
	coresql "github.com/LUSHDigital/core-sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
	"prediction-league/service/internal/domain"
	"testing"
)

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
	db := coresql.MustOpen("mysql", config.MySQLURL)
	driver, _ := mysql.WithInstance(db.DB, &mysql.Config{})
	mig, _ := migrate.NewWithDatabaseInstance(
		config.MigrationsURL,
		"mysql",
		driver,
	)
	coresql.MustMigrateUp(mig)

	domain.RegisterCustomValidators()

	// run test
	os.Exit(m.Run())
}

func TestItDoesSomething(t *testing.T) {
	t.Logf("MYSQL_URL = %s", os.Getenv("MYSQL_URL"))
}
