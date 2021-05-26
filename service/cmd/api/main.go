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

	app.RegisterRoutes(cnt)

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
	httpServer := app.NewServer(cnt)

	// setup email queue runner
	emailQueueRunner, err := app.NewEmailQueueRunner(cnt)
	if err != nil {
		return fmt.Errorf("cannot instantiate email queue runner: %w", err)
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
