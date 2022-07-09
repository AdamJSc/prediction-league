package app

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"
	"strings"

	"github.com/robfig/cron/v3"
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

// CronHandler encapsulates the logic required to generate our cron jobs
type CronHandler struct {
	entryAgent                 *domain.EntryAgent
	standingsAgent             *domain.StandingsAgent
	scoredEntryPredictionAgent *domain.ScoredEntryPredictionAgent
	commsAgent                 *domain.CommunicationsAgent
	mwSubmissionAgent          *domain.MatchWeekSubmissionAgent
	mwResultAgent              *domain.MatchWeekResultAgent
	seasonCollection           domain.SeasonCollection
	teamCollection             domain.TeamCollection
	realmCollection            domain.RealmCollection
	clock                      domain.Clock
	logger                     domain.Logger
	footballClient             domain.FootballDataSource
}

func (c *CronHandler) Run(_ context.Context) error {
	c.logger.Info("running cron handler...")
	cr, err := c.generateCron()
	if err != nil {
		return fmt.Errorf("cannot generate cron: %w", err)
	}
	cr.Start()
	return nil
}

func (c *CronHandler) Halt(_ context.Context) error {
	c.logger.Info("halting cron handler...")
	return nil
}

// generateCron generates a populated cron
func (c *CronHandler) generateCron() (*cron.Cron, error) {
	// get unique season IDs for all realms
	seasonIDs := make(map[string]struct{})
	for _, realm := range c.realmCollection {
		seasonIDs[realm.SeasonID] = struct{}{}
	}

	seasons := make([]domain.Season, 0)
	for id := range seasonIDs {
		s, err := c.seasonCollection.GetByID(id)
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
func (c *CronHandler) generateJobConfigs(seasons []domain.Season) ([]*jobConfig, error) {
	jobs := make([]*jobConfig, 0)

	for _, s := range seasons {
		j, err := c.newRetrieveLatestStandingsJob(s)
		if err != nil {
			return nil, fmt.Errorf("cannot generate new retrieve latest standings job: %w", err)
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

// newRetrieveLatestStandingsJob returns a new job that retrieves the latest standings, pertaining to the provided season
func (c *CronHandler) newRetrieveLatestStandingsJob(season domain.Season) (*jobConfig, error) {
	jobName := strings.ToLower(fmt.Sprintf("retrieve-latest-standings-%s", season.ID))

	params := domain.RetrieveLatestStandingsWorkerParams{
		Season:                     season,
		TeamCollection:             c.teamCollection,
		Clock:                      c.clock,
		Logger:                     c.logger,
		EntryAgent:                 c.entryAgent,
		StandingsAgent:             c.standingsAgent,
		ScoredEntryPredictionAgent: c.scoredEntryPredictionAgent,
		MatchWeekSubmissionAgent:   c.mwSubmissionAgent,
		MatchWeekResultAgent:       c.mwResultAgent,
		EmailIssuer:                c.commsAgent,
		FootballClient:             c.footballClient,
	}

	worker, err := domain.NewRetrieveLatestStandingsWorker(params)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate retrieve last standings worker: %w", err)
	}

	task, err := domain.HandleWorker(jobName, 5, worker, c.logger)

	return &jobConfig{
		spec: retrieveLatestStandingsCronSpec,
		task: task,
	}, nil
}

func NewCronHandler(c *container) (*CronHandler, error) {
	if c == nil {
		return nil, fmt.Errorf("container: %w", domain.ErrIsNil)
	}
	if c.entryAgent == nil {
		return nil, fmt.Errorf("entry agent: %w", domain.ErrIsNil)
	}
	if c.standingsAgent == nil {
		return nil, fmt.Errorf("standings agent: %w", domain.ErrIsNil)
	}
	if c.sepAgent == nil {
		return nil, fmt.Errorf("scored entry prediction agent: %w", domain.ErrIsNil)
	}
	if c.commsAgent == nil {
		return nil, fmt.Errorf("communications agent: %w", domain.ErrIsNil)
	}
	if c.mwSubmissionAgent == nil {
		return nil, fmt.Errorf("match week submission agent: %w", domain.ErrIsNil)
	}
	if c.mwResultAgent == nil {
		return nil, fmt.Errorf("match week result agent: %w", domain.ErrIsNil)
	}
	if c.seasons == nil {
		return nil, fmt.Errorf("season collection: %w", domain.ErrIsNil)
	}
	if c.teams == nil {
		return nil, fmt.Errorf("team collection: %w", domain.ErrIsNil)
	}
	if c.realms == nil {
		return nil, fmt.Errorf("realms: %w", domain.ErrIsNil)
	}
	if c.clock == nil {
		return nil, fmt.Errorf("clock: %w", domain.ErrIsNil)
	}
	if c.logger == nil {
		return nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}
	if c.ftblDataSrc == nil {
		return nil, fmt.Errorf("football data source: %w", domain.ErrIsNil)
	}

	return &CronHandler{
		entryAgent:                 c.entryAgent,
		standingsAgent:             c.standingsAgent,
		scoredEntryPredictionAgent: c.sepAgent,
		commsAgent:                 c.commsAgent,
		mwSubmissionAgent:          c.mwSubmissionAgent,
		mwResultAgent:              c.mwResultAgent,
		seasonCollection:           c.seasons,
		teamCollection:             c.teams,
		realmCollection:            c.realms,
		clock:                      c.clock,
		logger:                     c.logger,
		footballClient:             c.ftblDataSrc,
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
