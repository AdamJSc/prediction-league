package scheduler

import (
	"context"
	"fmt"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"strings"
	"time"
)

// newPredictionWindowBeginsJob returns a new job that issues emails to entrants
// when a new Prediction Window has been opened for the provided season
func newPredictionWindowBeginsJob(season models.Season, injector app.DependencyInjector) *job {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-begins-%s", season.ID))

	entryAgent := domain.EntryAgent{
		EntryAgentInjector: injector,
	}

	commsAgent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	var task = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// TODO - implementation
	}

	return &job{
		spec: "* * * * *", // TODO - determine interval
		task: task,
	}
}
