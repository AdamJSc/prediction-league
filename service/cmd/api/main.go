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
	cl := &domain.RealClock{}
	l, err := logger.NewLogger(os.Stdout, cl)
	if err != nil {
		log.Fatalf("cannot instantiate logger: %s", err.Error())
	}

	if err := run(l, cl); err != nil {
		l.Errorf("run failed: %s", err.Error())
	}
}

func run(l domain.Logger, cl domain.Clock) error {
	// setup env and config
	cfg := domain.MustLoadConfigFromEnvPaths(l, ".env", "infra/app.env")

	// setup db connection
	db, err := mysqldb.ConnectAndMigrate(cfg.MySQLURL, cfg.MigrationsURL, l)
	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			l.Info("database migration: no changes")
		default:
			return fmt.Errorf("failed to connect and migrate database: %w", err)
		}
	}
	defer func() {
		if err := db.Close(); err != nil {
			l.Errorf("cannot close db connection: %s", err.Error())
		}
	}()

	// permit flag that provides a debug mode by overriding timestamp for time-sensitive operations
	ts := flag.String("ts", "", "override timestamp used by time-sensitive operations, in the format yyyymmddhhmmss")
	flag.Parse()

	er, err := mysqldb.NewEntryRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate entry repo: %w", err)
	}
	epr, err := mysqldb.NewEntryPredictionRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate entry prediction repo: %w", err)
	}
	sepr, err := mysqldb.NewScoredEntryPredictionRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate scored entry prediction repo: %w", err)
	}
	sr, err := mysqldb.NewStandingsRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate standings repo: %w", err)
	}
	tr, err := mysqldb.NewTokenRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate token repo: %w", err)
	}
	sc, err := domain.GetSeasonCollection()
	if err != nil {
		return fmt.Errorf("cannot instantiate seasons collection: %w", err)
	}
	tc := domain.GetTeamCollection()

	chEml := make(chan domain.Email)
	tpl, err := domain.ParseTemplates("./service/views")
	if err != nil {
		return fmt.Errorf("cannot parse templates: %w", err)
	}

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
		return fmt.Errorf("cannot instantiate entry agent: %w", err)
	}
	sa, err := domain.NewStandingsAgent(sr)
	if err != nil {
		return fmt.Errorf("cannot instantiate standings agent: %w", err)
	}
	sepa, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		return fmt.Errorf("cannot instantiate scored entry prediction agent: %w", err)
	}
	ca, err := domain.NewCommunicationsAgent(cfg, er, epr, sr, chEml, tpl, sc, tc)
	if err != nil {
		return fmt.Errorf("cannot instantiate communications agent: %w", err)
	}
	var fds domain.FootballDataSource
	if cfg.FootballDataAPIToken != "" {
		fds, err = footballdataorg.NewClient(cfg.FootballDataAPIToken, tc)
		if err != nil {
			return fmt.Errorf("cannot instantiate football data org source: %w", err)
		}
	}
	rlms := cfg.Realms

	seeds, err := domain.GenerateSeedEntries()
	if err != nil {
		return fmt.Errorf("cannot generate entries to seed: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ea.SeedEntries(ctxWithTimeout, seeds); err != nil {
		return fmt.Errorf("cannot seed entries: %w", err)
	}

	app.RegisterRoutes(httpAppContainer)

	// start cron
	crFac, err := app.NewCronFactory(ea, sa, sepa, ca, sc, tc, rlms, cl, l, fds)
	if err != nil {
		return fmt.Errorf("cannot instantiate cron factory: %w", err)
	}
	cr, err := crFac.Make()
	if err != nil {
		return fmt.Errorf("cannot make cron: %w", err)
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

	return nil
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
