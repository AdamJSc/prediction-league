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

var runningVersion, versionTimestamp string

func main() {
	// base logger
	l, err := logger.NewLogger("DEBUG", os.Stdout, &domain.RealClock{})
	if err != nil {
		log.Fatalf("cannot instantiate logger: %s", err.Error())
	}

	// parse config
	config, err := app.NewConfigFromOptions(
		app.NewLoadEnvConfigOption(l, ".env", "infra/app.env"),
		app.NewRunningVersionConfigOption(runningVersion),
		app.NewVersionTimestampConfigOption(versionTimestamp))
	if err != nil {
		l.Errorf("cannot parse config from env: %s", err.Error())
		os.Exit(1)
	}

	// instantiate clock
	clock := getClock()
	if clock == nil {
		l.Error("clock is nil")
		os.Exit(1)
	}

	// re-configure logger
	l, err = logger.NewLogger(config.LogLevel, os.Stdout, clock)
	if err != nil {
		l.Errorf("cannot instantiate logger: %s", err.Error())
		os.Exit(1)
	}

	// build and run
	if err := run(config, l, clock); err != nil {
		l.Errorf("run failed: %s", err.Error())
		os.Exit(1)
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

	if t := parseTime("200601021504", *ts); t != nil {
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

func run(config *app.Config, l domain.Logger, clock domain.Clock) error {
	// setup container
	container, cleanup, err := app.NewContainer(config, l, clock)
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
	serverComponent, err := app.NewHTTPServer(container)
	if err != nil {
		return fmt.Errorf("cannot instantiate service component: http server: %w", err)
	}

	emailQueueComponent, err := app.NewEmailQueueRunner(container)
	if err != nil {
		return fmt.Errorf("cannot instantiate email queue runner: %w", err)
	}

	cronComponent, err := app.NewCronHandler(container)
	if err != nil {
		return fmt.Errorf("cannot instantiate cron handler: %w", err)
	}

	seederComponent, err := app.NewSeeder(container)
	if err != nil {
		return fmt.Errorf("cannot instantiate seeder: %w", err)
	}

	// setup service
	service, err := app.NewService("prediction-league", 5, l)
	if err != nil {
		return fmt.Errorf("cannot instantiate service: %w", err)
	}

	service.MustRun(
		context.Background(),
		serverComponent,
		emailQueueComponent,
		cronComponent,
		seederComponent,
	)

	return nil
}
