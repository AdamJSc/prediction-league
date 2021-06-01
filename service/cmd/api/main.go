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
	// TODO - clock: replace with clock usage (implement domain.FrozenClock)
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
