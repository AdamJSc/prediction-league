package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"time"
)

func main() {
	cl := getClock()
	if cl == nil {
		log.Fatal("clock is nil")
	}

	// TODO - logger: parse logger level from config
	l, err := logger.NewLogger(logger.LevelDebug, os.Stdout, cl)
	if err != nil {
		log.Fatalf("cannot instantiate logger: %s", err.Error())
	}

	if err := run(l, cl); err != nil {
		l.Errorf("run failed: %s", err.Error())
	}
}

func getClock() domain.Clock {
	ts := flag.String("ts", "", "override timestamp used by time-sensitive operations, in the format yyyymmddhhmmss")
	flag.Parse()

	if ts == nil {
		return &domain.RealClock{}
	}
	if t := parseTime("20060102150405", *ts); t != nil {
		return &domain.FrozenClock{Time: *t}
	}
	if t := parseTime("20060102", *ts); t != nil {
		return &domain.FrozenClock{Time: *t}
	}

	return &domain.RealClock{}
}

func parseTime(layout, value string) *time.Time {
	parsed, err := time.Parse(layout, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func run(l domain.Logger, cl domain.Clock) error {
	// setup env and config
	cfg, err := app.NewConfigFromEnvPaths(l, ".env", "infra/app.env")
	if err != nil {
		return fmt.Errorf("cannot parse config from end: %w", err)
	}

	// setup container
	cnt, cleanup, err := app.NewContainer(cfg, l, cl)
	if err != nil {
		return fmt.Errorf("cannot instantiate container: %w", err)
	}

	// defer cleanup
	defer func() {
		if err := cleanup(); err != nil {
			l.Errorf("cleanup failed: %s", err.Error())
		}
	}()

	// setup components
	cmpServer, err := app.NewHTTPServer(cnt)
	if err != nil {
		return fmt.Errorf("cannot instantiate service component: http server: %w", err)
	}
	cmpEmlQ, err := app.NewEmailQueueRunner(cnt)
	if err != nil {
		return fmt.Errorf("cannot instantiate email queue runner: %w", err)
	}
	cmpCron, err := app.NewCronHandler(cnt)
	if err != nil {
		return fmt.Errorf("cannot instantiate cron handler: %w", err)
	}
	cmpSeeder, err := app.NewSeeder(cnt)
	if err != nil {
		return fmt.Errorf("cannot instantiate seeder: %w", err)
	}

	// setup service
	svc, err := app.NewService("prediction-league", 5, l)
	if err != nil {
		return fmt.Errorf("cannot instantiate service: %w", err)
	}
	svc.MustRun(
		context.Background(),
		cmpServer,
		cmpEmlQ,
		cmpCron,
		cmpSeeder,
	)

	return nil
}
