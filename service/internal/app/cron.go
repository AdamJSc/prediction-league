package app

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"prediction-league/service/internal/domain"
	"strings"
)

const (
	// predictionWindowOpenCronSpec determines the frequency by which the PredictionWindowOpenWorker will run
	// (i.e. every day at 12:34pm)
	predictionWindowOpenCronSpec = "34 12 * * *"

	// predictionWindowClosingCronSpec determines the frequency by which the PredictionWindowClosingWorker will run
	// (i.e. every day at 4:48pm)
	predictionWindowClosingCronSpec = "48 16 * * *"

	// retrieveLatestStandingsCronSpec determines the frequency by which the RetrieveLatestStandingsWorker will run
	retrieveLatestStandingsCronSpec = "@every 0h15m"
)

// CronFactory encapsulates the logic required to generate our cron jobs
type CronFactory struct {
	ea   *domain.EntryAgent
	sa   *domain.StandingsAgent
	sepa *domain.ScoredEntryPredictionAgent
	ca   *domain.CommunicationsAgent
	sc   domain.SeasonCollection
	tc   domain.TeamCollection
	rc   domain.RealmCollection
	cl   domain.Clock
	l    domain.Logger
	fds  domain.FootballDataSource
}

func NewCronFactory(cnt *container) (*CronFactory, error) {
	if cnt == nil {
		return nil, fmt.Errorf("container: %w", domain.ErrIsNil)
	}
	if cnt.entryAgent == nil {
		return nil, fmt.Errorf("entry agent: %w", domain.ErrIsNil)
	}
	if cnt.standingsAgent == nil {
		return nil, fmt.Errorf("standings agent: %w", domain.ErrIsNil)
	}
	if cnt.sepAgent == nil {
		return nil, fmt.Errorf("scored entry prediction agent: %w", domain.ErrIsNil)
	}
	if cnt.commsAgent == nil {
		return nil, fmt.Errorf("communications agent: %w", domain.ErrIsNil)
	}
	if cnt.seasons == nil {
		return nil, fmt.Errorf("season collection: %w", domain.ErrIsNil)
	}
	if cnt.teams == nil {
		return nil, fmt.Errorf("team collection: %w", domain.ErrIsNil)
	}
	if cnt.realms == nil {
		return nil, fmt.Errorf("realms: %w", domain.ErrIsNil)
	}
	if cnt.clock == nil {
		return nil, fmt.Errorf("clock: %w", domain.ErrIsNil)
	}
	if cnt.logger == nil {
		return nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}
	// do not check fds, allow nil
	return &CronFactory{
		cnt.entryAgent,
		cnt.standingsAgent,
		cnt.sepAgent,
		cnt.commsAgent,
		cnt.seasons,
		cnt.teams,
		cnt.realms,
		cnt.clock,
		cnt.logger,
		cnt.ftblDataSrc,
	}, nil
}

// Make generates our populated cron
func (c *CronFactory) Make() (*cron.Cron, error) {
	// get unique season IDs for all realms
	sIDs := make(map[string]struct{})
	for _, rlm := range c.rc {
		sIDs[rlm.SeasonID] = struct{}{}
	}

	seasons := make([]domain.Season, 0)
	for id := range sIDs {
		s, err := c.sc.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve season by id '%s': %w", id, err)
		}

		seasons = append(seasons, s)
	}

	if len(seasons) < 1 {
		return nil, errors.New("need at least one season for active realms")
	}

	j, err := c.generateJobConfigs(seasons)
	if err != nil {
		return nil, fmt.Errorf("cannot generate job configs: %w", err)
	}

	cr, err := newCronFromJobConfigs(j)
	if err != nil {
		return nil, fmt.Errorf("cannot initialise cron from job configs: %w", err)
	}

	return cr, nil
}

// generateJobConfigs returns a slice of job configs for all jobs required by each provided Season
func (c *CronFactory) generateJobConfigs(seasons []domain.Season) ([]*jobConfig, error) {
	jobs := make([]*jobConfig, 0)

	for _, s := range seasons {
		oj, err := c.newPredictionWindowOpenJob(s)
		if err != nil {
			return nil, fmt.Errorf("cannot generate new prediction window open job: %w", err)
		}

		cj, err := c.newPredictionWindowClosingJob(s)
		if err != nil {
			return nil, fmt.Errorf("cannot generate new prediction window closing job: %w", err)
		}

		jobs = append(jobs, oj, cj)

		// TODO - remove switch when replaced with alt domain.FootballDataSource
		switch {
		case c.fds == nil:
			c.l.Infof("skipping retrieve latest standings job for season %s: no football data source configured", s.ID)
		}

		if c.fds != nil {
			j, err := c.newRetrieveLatestStandingsJob(s)
			if err != nil {
				return nil, fmt.Errorf("cannot generate new retrieve latest standings job: %w", err)
			}

			jobs = append(jobs, j)
		}
	}

	return jobs, nil
}

// newPredictionWindowOpenJob returns a new job that issues emails to entrants
// when a new Prediction Window has been opened for the provided season
func (c *CronFactory) newPredictionWindowOpenJob(s domain.Season) (*jobConfig, error) {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-open-%s", s.ID))

	w, err := domain.NewPredictionWindowOpenWorker(s, c.cl, c.ea, c.ca)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate prediction window open worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, w, c.l)

	return &jobConfig{
		spec: predictionWindowOpenCronSpec,
		task: task,
	}, nil
}

// newPredictionWindowClosingJob returns a new job that issues emails to entrants
// when an active Prediction Window is due to close for the provided season
func (c *CronFactory) newPredictionWindowClosingJob(s domain.Season) (*jobConfig, error) {
	jobName := strings.ToLower(fmt.Sprintf("prediction-window-closing-%s", s.ID))

	w, err := domain.NewPredictionWindowClosingWorker(s, c.cl, c.ea, c.ca)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate prediction window closing worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, w, c.l)

	return &jobConfig{
		spec: predictionWindowClosingCronSpec,
		task: task,
	}, nil
}

// newRetrieveLatestStandingsJob returns a new job that retrieves the latest standings, pertaining to the provided season
func (c *CronFactory) newRetrieveLatestStandingsJob(s domain.Season) (*jobConfig, error) {
	jobName := strings.ToLower(fmt.Sprintf("retrieve-latest-standings-%s", s.ID))

	w, err := domain.NewRetrieveLatestStandingsWorker(s, c.tc, c.cl, c.ea, c.sa, c.sepa, c.ca, c.fds)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate retrieve last standings worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, w, c.l)

	return &jobConfig{
		spec: retrieveLatestStandingsCronSpec,
		task: task,
	}, nil
}

// newCronFromJobConfigs initialises a cron using the provided job configs
func newCronFromJobConfigs(jobs []*jobConfig) (*cron.Cron, error) {
	cr := cron.New()

	for _, j := range jobs {
		if _, err := cr.AddFunc(j.spec, j.task); err != nil {
			return nil, fmt.Errorf("cannot add function: %w", err)
		}
	}

	return cr, nil
}

// jobConfig encapsulates our cron jobConfig attributes
type jobConfig struct {
	spec string
	task func()
}
