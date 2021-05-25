package scheduler

import (
	"fmt"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"strings"
)

// newRetrieveLatestStandingsJob returns a new job that retrieves the latest standings, pertaining to the provided season
func newRetrieveLatestStandingsJob(s domain.Season, fds domain.FootballDataSource, d app.DependencyInjector) (*job, error) {
	jobName := strings.ToLower(fmt.Sprintf("retrieve-latest-standings-%s", s.ID))

	ea, err := domain.NewEntryAgent(d.EntryRepo(), d.EntryPredictionRepo(), d.Seasons())
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate entry agent: %w", err)
	}

	sa, err := domain.NewStandingsAgent(d.StandingsRepo())
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate standings agent: %w", err)
	}

	sepa, err := domain.NewScoredEntryPredictionAgent(
		d.EntryRepo(),
		d.EntryPredictionRepo(),
		d.StandingsRepo(),
		d.ScoredEntryPredictionRepo(),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate scored entry prediction agent: %w", err)
	}

	ca, err := domain.NewCommunicationsAgent(
		d.Config(),
		d.EntryRepo(),
		d.EntryPredictionRepo(),
		d.StandingsRepo(),
		d.EmailQueue(),
		d.Template(),
		d.Seasons(),
		d.Teams(),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate communications agent: %w", err)
	}

	w, err := domain.NewRetrieveLatestStandingsWorker(s, d.Teams(), d.Clock(), ea, sa, sepa, ca, fds)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate retrieve last standings worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, w, d.Logger())

	return &job{
		spec: "@every 0h15m",
		task: task,
	}, nil
}
