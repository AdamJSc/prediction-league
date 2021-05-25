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
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cl := &domain.RealClock{}

	// setup logger
	l, err := logger.NewLogger(os.Stdout, cl)
	if err != nil {
		log.Fatalf("cannot instantiate new logger: %s", err.Error())
	}

	// setup env and config
	cfg := domain.MustLoadConfigFromEnvPaths(l, ".env", "infra/app.env")

	// setup db connection
	db, err := mysqldb.ConnectAndMigrate(cfg.MySQLURL, cfg.MigrationsURL, l)
	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			log.Println("database migration: no changes")
		default:
			log.Fatalf("failed to connect and migrate database: %s", err.Error())
		}
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("cannot close db connection: %s", err.Error())
		}
	}()

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
		log.Fatalf("cannot instantiate scored entry prediction repo: %s", err.Error())
	}
	sr, err := mysqldb.NewStandingsRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate standings repo: %s", err.Error())
	}
	tr, err := mysqldb.NewTokenRepo(db)
	if err != nil {
		log.Fatalf("cannot instantiate token repo: %s", err.Error())
	}
	sc, err := domain.GetSeasonCollection()
	if err != nil {
		log.Fatalf("cannot instantiate seasons collection: %s", err.Error())
	}
	tc := domain.GetTeamCollection()

	chEml := make(chan domain.Email)
	tpl := domain.MustParseTemplates("./service/views")

	// setup server
	httpAppContainer := app.NewHTTPAppContainer(dependencies{
		config:                    cfg,
		emailClient:               mailgun.NewClient(cfg.MailgunAPIKey),
		emailQueue:                chEml,
		router:                    mux.NewRouter(),
		templates:                 tpl,
		debugTimestamp:            parseTimeString(ts),
		standingsRepo:             sr,
		entryRepo:                 er,
		entryPredictionRepo:       epr,
		scoredEntryPredictionRepo: sepr,
		tokenRepo:                 tr,
		seasons:                   sc,
		teams:                     tc,
		clock:                     cl,
		logger:                    l,
	})

	ea, err := domain.NewEntryAgent(er, epr, sc)
	if err != nil {
		log.Fatalf("cannot instantiate entry agent: %s", err.Error())
	}
	sa, err := domain.NewStandingsAgent(sr)
	if err != nil {
		log.Fatalf("cannot instantiate standings agent: %s", err.Error())
	}
	sepa, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		log.Fatalf("cannot instantiate scored entry prediction agent: %s", err.Error())
	}
	ca, err := domain.NewCommunicationsAgent(cfg, er, epr, sr, chEml, tpl, sc, tc)
	if err != nil {
		log.Fatalf("cannot instantiate communications agent: %s", err.Error())
	}
	var fds domain.FootballDataSource
	if cfg.FootballDataAPIToken != "" {
		fds, err = footballdataorg.NewClient(cfg.FootballDataAPIToken, tc)
		if err != nil {
			log.Fatalf("cannot instantiate football data org source: %s", err.Error())
		}
	}
	rlms := cfg.Realms

	seeds, err := domain.GenerateSeedEntries()
	if err != nil {
		log.Fatalf("cannot generate entries to seed: %s", err.Error())
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ea.SeedEntries(ctxWithTimeout, seeds); err != nil {
		log.Fatalf("cannot seed entries: %s", err.Error())
	}

	app.RegisterRoutes(httpAppContainer)

	// start cron
	crFac, err := app.NewCronFactory(ea, sa, sepa, ca, sc, tc, rlms, cl, l, fds)
	if err != nil {
		log.Fatalf("cannot instantiate cron factory: %s", err.Error())
	}
	cr, err := crFac.Make()
	if err != nil {
		log.Fatalf("cannot make cron: %s", err.Error())
	}
	cr.Start()

	// setup http server process
	httpServer := app.NewServer(&http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServicePort),
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
	config                    *domain.Config
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
	seasons                   domain.SeasonCollection
	teams                     domain.TeamCollection
	clock                     domain.Clock
	logger                    domain.Logger
}

func (d dependencies) Config() *domain.Config          { return d.config }
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
func (d dependencies) Seasons() domain.SeasonCollection  { return d.seasons }
func (d dependencies) Teams() domain.TeamCollection      { return d.teams }
func (d dependencies) Clock() domain.Clock               { return d.clock }
func (d dependencies) Logger() domain.Logger             { return d.logger }

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
