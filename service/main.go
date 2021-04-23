package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/app/httph/handlers"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/repositories"
	"prediction-league/service/internal/repositories/repofac"
	"prediction-league/service/internal/scheduler"
	"time"

	"github.com/LUSHDigital/core"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core/workers/httpsrv"
	_ "github.com/go-sql-driver/mysql"
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
	mig, err := migrate.NewWithDatabaseInstance(
		config.MigrationsURL,
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot open sql connection: %w", err))
	}
	coresql.MustMigrateUp(mig)

	// permit flag that provides a debug mode by overriding timestamp for time-sensitive operations
	ts := flag.String("ts", "", "override timestamp used by time-sensitive operations, in the format yyyymmddhhmmss")
	flag.Parse()

	// setup server
	httpAppContainer := httph.NewHTTPAppContainer(dependencies{
		config:                    config,
		mysql:                     db,
		emailClient:               mailgun.NewClient(config.MailgunAPIKey),
		emailQueue:                make(chan domain.Email),
		router:                    mux.NewRouter(),
		templates:                 domain.MustParseTemplates("./service/views"),
		debugTimestamp:            parseTimeString(ts),
		standingsRepo:             repofac.NewStandingsDatabaseRepository(db),
		entryRepo:                 repofac.NewEntryDatabaseRepository(db),
		entryPredictionRepo:       repofac.NewEntryPredictionDatabaseRepository(db),
		scoredEntryPredictionRepo: repofac.NewScoredEntryPredictionDatabaseRepository(db),
		tokenRepo:                 repofac.NewTokenDatabaseRepository(db),
	})

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	entryAgent := &domain.EntryAgent{EntryAgentInjector: httpAppContainer}
	seeds, err := domain.GenerateSeedEntries()
	if err != nil {
		log.Fatal(fmt.Errorf("cannot generate entries to seed: %w", err))
	}
	if err := entryAgent.SeedEntries(ctxWithTimeout, seeds); err != nil {
		log.Fatal(fmt.Errorf("cannot seed entries: %w", err))
	}

	domain.MustInflate()

	handlers.RegisterRoutes(httpAppContainer)

	// start cron
	scheduler.LoadCron(httpAppContainer).Start()

	// setup http server process
	httpServer := httpsrv.New(&http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServicePort),
		Handler: httpAppContainer.Router(),
	})

	// setup email queue runner
	emailQueueRunner := app.EmailQueueRunner{
		EmailQueueRunnerInjector: httpAppContainer,
	}

	// run service
	svc := &core.Service{
		Name: "prediction-league",
		Type: "service",
	}
	svc.MustRun(
		context.Background(),
		httpServer,
		emailQueueRunner,
	)
}

type dependencies struct {
	config                    domain.Config
	mysql                     coresql.Agent
	emailClient               domain.EmailClient
	emailQueue                chan domain.Email
	router                    *mux.Router
	templates                 *domain.Templates
	debugTimestamp            *time.Time
	standingsRepo             *repositories.StandingsDatabaseRepository
	entryRepo                 *repositories.EntryDatabaseRepository
	entryPredictionRepo       *repositories.EntryPredictionDatabaseRepository
	scoredEntryPredictionRepo *repositories.ScoredEntryPredictionDatabaseRepository
	tokenRepo                 *repositories.TokenDatabaseRepository
}

func (d dependencies) Config() domain.Config           { return d.config }
func (d dependencies) EmailClient() domain.EmailClient { return d.emailClient }
func (d dependencies) EmailQueue() chan domain.Email   { return d.emailQueue }
func (d dependencies) Router() *mux.Router             { return d.router }
func (d dependencies) Template() *domain.Templates     { return d.templates }
func (d dependencies) DebugTimestamp() *time.Time      { return d.debugTimestamp }
func (d dependencies) StandingsRepo() domain.StandingsRepository {
	return d.standingsRepo
}
func (d dependencies) EntryRepo() domain.EntryRepository { return d.entryRepo }
func (d dependencies) EntryPredictionRepo() domain.EntryPredictionRepository {
	return d.entryPredictionRepo
}
func (d dependencies) ScoredEntryPredictionRepo() domain.ScoredEntryPredictionRepository {
	return d.scoredEntryPredictionRepo
}
func (d dependencies) TokenRepo() domain.TokenRepository { return d.tokenRepo }

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
