package main

import (
	"context"
	"fmt"
	"github.com/LUSHDigital/core"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core/workers/httpsrv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/gorilla/mux"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/app/httph/handlers"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/scheduler"
	"prediction-league/service/internal/seeder"
	"prediction-league/service/internal/views"
)

func main() {
	// setup env and config
	config := domain.MustLoadConfigFromEnvPaths(".env", "infra/app.env")

	// setup db connection
	db := coresql.MustOpen("mysql", config.MySQLURL)
	driver, _ := mysql.WithInstance(db.DB, &mysql.Config{})
	mig, _ := migrate.NewWithDatabaseInstance(
		config.MigrationsURL,
		"mysql",
		driver,
	)
	coresql.MustMigrateUp(mig)
	seeder.MustSeed(db)

	datastore.MustInflate()

	// setup server
	httpAppContainer := httph.NewHTTPAppContainer(dependencies{
		config:    config,
		mysql:     db,
		router:    mux.NewRouter(),
		templates: domain.ParseTemplates(),
	})
	handlers.RegisterRoutes(httpAppContainer)

	httpServer := httpsrv.New(&http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServicePort),
		Handler: httpAppContainer.Router(),
	})

	jobScheduler := scheduler.Cron{
		Jobs: scheduler.MustGenerateCronJobs(config),
	}

	// run service
	svc := &core.Service{
		Name: "prediction-league",
		Type: "service",
	}
	svc.MustRun(
		context.Background(),
		httpServer,
		jobScheduler,
	)
}

type dependencies struct {
	config    domain.Config
	mysql     coresql.Agent
	router    *mux.Router
	templates *views.Templates
}

func (d dependencies) Config() domain.Config      { return d.config }
func (d dependencies) MySQL() coresql.Agent       { return d.mysql }
func (d dependencies) Router() *mux.Router        { return d.router }
func (d dependencies) Template() *views.Templates { return d.templates }
