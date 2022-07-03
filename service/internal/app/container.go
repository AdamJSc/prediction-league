package app

import (
	"database/sql"
	"errors"
	"fmt"
	"prediction-league/service/internal/adapters"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"

	"github.com/golang-migrate/migrate/v4"
)

// container encapsulates the app dependencies
type container struct {
	config         *Config
	realms         domain.RealmCollection
	seasons        domain.SeasonCollection
	teams          domain.TeamCollection
	templates      *domain.Templates
	commsAgent     *domain.CommunicationsAgent
	entryAgent     *domain.EntryAgent
	standingsAgent *domain.StandingsAgent
	sepAgent       *domain.ScoredEntryPredictionAgent
	tokenAgent     *domain.TokenAgent
	lbAgent        *domain.LeaderBoardAgent
	emailClient    domain.EmailClient
	emailQueue     domain.EmailQueue
	ftblDataSrc    domain.FootballDataSource
	logger         domain.Logger
	clock          domain.Clock
}

// NewContainer instantiates a new container from the provided config object
func NewContainer(cfg *Config, l domain.Logger, cl domain.Clock) (*container, func() error, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config: %w", domain.ErrIsNil)
	}
	if l == nil {
		return nil, nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}

	// setup db connection
	projectRootDir := "../.."
	migrationsURL := fmt.Sprintf("file://%s/%s", projectRootDir, cfg.MigrationsPath)
	db, err := sqlConnectAndMigrate(cfg.MySQLURL, migrationsURL, l)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect and migrate database: %w", err)
	}

	// parse templates
	tpl, err := domain.ParseTemplates("./service/views")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse templates: %w", err)
	}

	// instantiate email queue
	emlQ := domain.NewInMemEmailQueue()

	// instantiate email client
	var emlCl domain.EmailClient
	switch {
	case cfg.MailgunAPIKey != "":
		l.Info("mailgun client credentials found...")
		emlCl, err = mailgun.NewClient(cfg.MailgunAPIKey)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot instantiate mailgun email client: %w", err)
		}
	default:
		l.Info("missing mailgun client credentials: transactional email content will be printed to log...")
		emlCl, err = domain.NewNoopEmailClient(l)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot instantiate logger email client: %w", err)
		}
	}

	// instantiate collections
	rc, err := domain.GetRealmCollection()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate realm collection: %w", err)
	}
	sc, err := domain.GetSeasonCollection()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate seasons collection: %w", err)
	}
	tc := domain.GetTeamCollection()

	// instantiate football-data.org client
	var fds domain.FootballDataSource
	switch {
	case cfg.FootballDataAPIToken != "":
		hc := adapters.NewRealHTTPClient(10)
		fds, err = footballdataorg.NewClient(cfg.FootballDataAPIToken, tc, hc)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot instantiate football-data.org client: %w", err)
		}
	default:
		l.Info("missing football data api token: retrieving latest standings will not occur in upstream...")
		fds, err = domain.NewNoopFootballDataSource(l)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot instantiate noop football data source: %w", err)
		}
	}

	// instantiate repos
	er, err := mysqldb.NewEntryRepo(db)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate entry repo: %w", err)
	}
	epr, err := mysqldb.NewEntryPredictionRepo(db)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate entry prediction repo: %w", err)
	}
	sepr, err := mysqldb.NewScoredEntryPredictionRepo(db)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate scored entry prediction repo: %w", err)
	}
	sr, err := mysqldb.NewStandingsRepo(db)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate standings repo: %w", err)
	}
	tr, err := mysqldb.NewTokenRepo(db)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate token repo: %w", err)
	}

	// instantiate agents
	ca, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate communications agent: %w", err)
	}
	ea, err := domain.NewEntryAgent(er, epr, sr, sc, cl)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate entry agent: %w", err)
	}
	sa, err := domain.NewStandingsAgent(sr)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate standings agent: %w", err)
	}
	sepa, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate scored entry prediction agent: %w", err)
	}
	ta, err := domain.NewTokenAgent(tr, cl, l)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate token agent: %w", err)
	}
	lba, err := domain.NewLeaderBoardAgent(er, epr, sr, sepr, sc)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate leaderboard agent: %w", err)
	}

	cnt := &container{
		cfg,
		rc,
		sc,
		tc,
		tpl,
		ca,
		ea,
		sa,
		sepa,
		ta,
		lba,
		emlCl,
		emlQ,
		fds,
		l,
		cl,
	}

	// define cleanup function
	cleanup := func() error {
		if err := db.Close(); err != nil {
			return fmt.Errorf("cannot close db connection: %w", err)
		}
		return nil
	}

	return cnt, cleanup, nil
}

func sqlConnectAndMigrate(dbURL, migURL string, l domain.Logger) (*sql.DB, error) {
	db, err := mysqldb.ConnectAndMigrate(dbURL, migURL, l)
	if err != nil {
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			l.Info("database migration: no changes")
		default:
			return nil, err
		}
	}

	return db, nil
}
