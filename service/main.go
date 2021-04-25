package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/app/scheduler"
	"prediction-league/service/internal/domain"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// setup logger
	l, err := logger.NewLogger(os.Stdout, &logger.RealClock{})
	if err != nil {
		log.Fatalf("cannot instantiate new logger: %s", err.Error())
	}

	// setup env and config
	config := domain.MustLoadConfigFromEnvPaths(l, ".env", "infra/app.env")

	// setup db connection
	db, err := mysqldb.ConnectAndMigrate(config.MySQLURL, config.MigrationsURL, l)
	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			log.Println("database migration: no changes")
		default:
			log.Fatalf("failed to connect and migrate database: %s", err.Error())
		}
	}
	defer db.Close()

	// permit flag that provides a debug mode by overriding timestamp for time-sensitive operations
	ts := flag.String("ts", "", "override timestamp used by time-sensitive operations, in the format yyyymmddhhmmss")
	flag.Parse()

	er, err := mysqldb.NewEntryRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate entry repo: %s", err.Error())
	}
	epr, err := mysqldb.NewEntryPredictionRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate entry prediction repo: %s", err.Error())
	}
	sepr, err := mysqldb.NewScoredEntryPredictionRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate score entry prediction repo: %s", err.Error())
	}
	sr, err := mysqldb.NewStandingsRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate standings repo: %s", err.Error())
	}
	tr, err := mysqldb.NewTokenRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate token repo: %s", err.Error())
	}

	// setup server
	httpAppContainer := app.NewHTTPAppContainer(dependencies{
		config:                    config,
		emailClient:               mailgun.NewClient(config.MailgunAPIKey),
		emailQueue:                make(chan domain.Email),
		router:                    mux.NewRouter(),
		templates:                 domain.MustParseTemplates("./service/views"),
		debugTimestamp:            parseTimeString(ts),
		standingsRepo:             sr,
		entryRepo:                 er,
		entryPredictionRepo:       epr,
		scoredEntryPredictionRepo: sepr,
		tokenRepo:                 tr,
	})

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	entryAgent := &domain.EntryAgent{EntryAgentInjector: httpAppContainer}
	seeds, err := domain.GenerateSeedEntries()
	if err != nil {
		log.Fatalf("cannot generate entries to seed: %s", err.Error())
	}
	if err := entryAgent.SeedEntries(ctxWithTimeout, seeds); err != nil {
		log.Fatalf("cannot seed entries: %s", err.Error())
	}

	domain.MustInflate()

	app.RegisterRoutes(httpAppContainer)

	// start cron
	scheduler.LoadCron(httpAppContainer).Start()

	// setup http server process
	httpServer := app.NewServer(&http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServicePort),
		Handler: httpAppContainer.Router(),
	})

	// setup email queue runner
	emailQueueRunner := app.EmailQueueRunner{
		EmailQueueRunnerInjector: httpAppContainer,
	}

	// run service
	svc := &app.Service{
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
	emailClient               domain.EmailClient
	emailQueue                chan domain.Email
	router                    *mux.Router
	templates                 *domain.Templates
	debugTimestamp            *time.Time
	standingsRepo             *mysqldb.StandingsRepo
	entryRepo                 *mysqldb.EntryRepo
	entryPredictionRepo       *mysqldb.EntryPredictionRepo
	scoredEntryPredictionRepo *mysqldb.ScoredEntryPredictionRepo
	tokenRepo                 *mysqldb.TokenRepo
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
