package scheduler

import (
	"fmt"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"strings"
)

// predictionWindowOpenCronSpec determines the frequency by which the PredictionWindowOpenJob will run
// (i.e. every day at 12:34pm)
const predictionWindowOpenCronSpec = "34 12 * * *"

// newPredictionWindowOpenJob returns a new job that issues emails to entrants
// when a new Prediction Window has been opened for the provided season
func newPredictionWindowOpenJob(s domain.Season, d app.DependencyInjector) (*job, error) {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-open-%s", s.ID))

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

	w, err := domain.NewPredictionWindowOpenWorker(s, d.Clock(), ea, ca)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate prediction window open worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, w, d.Logger())

	return &job{
		spec: predictionWindowOpenCronSpec,
		task: task,
	}, nil
}
