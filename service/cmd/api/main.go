package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"prediction-league/service/internal/adapters/logger"
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
	// permit flag that provides a debug mode by overriding timestamp for time-sensitive operations
	ts := flag.String("ts", "", "override timestamp used by time-sensitive operations, in the format yyyymmddhhmmss")
	flag.Parse()

	// setup env and config
	cfg, err := app.NewConfigFromEnvPaths(l, ".env", "infra/app.env")
	if err != nil {
		return fmt.Errorf("cannot parse config from end: %w", err)
	}

	// setup container
	cnt, cleanup, err := app.NewContainer(cfg, l, cl, ts)
	if err != nil {
		return fmt.Errorf("cannot instantiate container: %w", err)
	}

	// defer cleanup
	defer func() {
		if err := cleanup(); err != nil {
			l.Errorf("cleanup failed: %s", err.Error())
		}
	}()

	if err := app.Seed(cnt); err != nil {
		return fmt.Errorf("cannot run seeder: %w", err)
	}

	app.RegisterRoutes(httpAppContainer)

	// TODO - implement as Service with Worker interface (Run/Halt)
	// start cron
	crFac, err := app.NewCronFactory(cnt)
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
	config                    *app.Config
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
	realms                    domain.RealmCollection
	clock                     domain.Clock
	logger                    domain.Logger
}

func (d dependencies) Config() *app.Config             { return d.config }
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
func (d dependencies) Realms() domain.RealmCollection    { return d.realms }
func (d dependencies) Clock() domain.Clock               { return d.clock }
func (d dependencies) Logger() domain.Logger             { return d.logger }
