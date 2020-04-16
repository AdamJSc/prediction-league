package main

import (
	"context"
	"fmt"
	"github.com/LUSHDigital/core"
	"github.com/LUSHDigital/core-mage/env"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core/workers/httpsrv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/app/httph/handlers"
	"prediction-league/service/internal/domain"
)

func main() {
	// setup env
	env.Load("infra/app.env")
	config := struct {
		ServicePort   string `envconfig:"SERVICE_PORT" required:"true"`
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

	// setup server
	httpAppContainer := httph.NewHTTPAppContainer(dependencies{
		mysql:  db,
		router: mux.NewRouter(),
	})
	handlers.RegisterRoutes(httpAppContainer)

	httpServer := httpsrv.New(&http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServicePort),
		Handler: httpAppContainer.Router(),
	})

	// run service
	svc := &core.Service{
		Name: "prediction-league",
		Type: "service",
	}
	svc.MustRun(
		context.Background(),
		httpServer,
	)
}

type dependencies struct {
	mysql  coresql.Agent
	router *mux.Router
}

func (d dependencies) MySQL() coresql.Agent { return d.mysql }
func (d dependencies) Router() *mux.Router  { return d.router }
