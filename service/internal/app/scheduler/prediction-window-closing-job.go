package scheduler

import (
	"fmt"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"strings"
)

// predictionWindowClosingCronSpec determines the frequency by which the PredictionWindowClosingWorker will run
// (i.e. every day at 4:48pm)
const predictionWindowClosingCronSpec = "48 16 * * *"

// newPredictionWindowClosingJob returns a new job that issues emails to entrants
// when an active Prediction Window is due to close for the provided season
func newPredictionWindowClosingJob(s domain.Season, d app.DependencyInjector) (*job, error) {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-closing-%s", s.ID))

	ea, err := domain.NewEntryAgent(d.EntryRepo(), d.EntryPredictionRepo(), d.Seasons())
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate entry agent: %w", err)
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

	w, err := domain.NewPredictionWindowClosingWorker(s, d.Clock(), ea, ca)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate prediction window closing worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, w, d.Logger())

	return &job{
		spec: predictionWindowClosingCronSpec,
		task: task,
	}, nil
}
