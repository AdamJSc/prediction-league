package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/LUSHDigital/core"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core/workers/httpsrv"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/app/httph/handlers"
	"prediction-league/service/internal/clients"
	"prediction-league/service/internal/clients/sendgrid"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/scheduler"
	"prediction-league/service/internal/seeder"
	"prediction-league/service/internal/views"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/gorilla/mux"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	// permit flag that provides a debug mode by overriding timestamp for time-sensitive operations
	ts := flag.String("ts", "", "override timestamp used by time-sensitive operations, in the format yyyymmddhhmmss")
	flag.Parse()

	// setup server
	httpAppContainer := httph.NewHTTPAppContainer(dependencies{
		config:         config,
		mysql:          db,
		emailClient:    sendgrid.NewClient(config.SendGridAPIKey),
		router:         mux.NewRouter(),
		templates:      domain.MustParseTemplates(),
		debugTimestamp: parseTimeString(ts),
	})
	handlers.RegisterRoutes(httpAppContainer)

	httpServer := httpsrv.New(&http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServicePort),
		Handler: httpAppContainer.Router(),
	})

	// start cron
	scheduler.LoadCron(httpAppContainer).Start()

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
	config         domain.Config
	mysql          coresql.Agent
	emailClient    clients.EmailClient
	router         *mux.Router
	templates      *views.Templates
	debugTimestamp *time.Time
}

func (d dependencies) Config() domain.Config            { return d.config }
func (d dependencies) MySQL() coresql.Agent             { return d.mysql }
func (d dependencies) EmailClient() clients.EmailClient { return d.emailClient }
func (d dependencies) Router() *mux.Router              { return d.router }
func (d dependencies) Template() *views.Templates       { return d.templates }
func (d dependencies) DebugTimestamp() *time.Time       { return d.debugTimestamp }

func parseTimeString(t *string) *time.Time {
	if t == nil {
		return nil
	}

	timeString := *t
	if timeString == "" {
		return nil
	}

	var (
		parsed time.Time
		err    error
	)

	parsed, err = time.Parse("20060102150405", timeString)
	if err != nil {
		parsed, err = time.Parse("20060102", timeString)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &parsed
}
