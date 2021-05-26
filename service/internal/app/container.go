package app

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/gorilla/mux"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"time"
)

// container encapsulates the app dependencies
type container struct {
	*config
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
	emailQueue     chan domain.Email
	ftblDataSrc    domain.FootballDataSource
	router         *mux.Router
	debugTs        *time.Time
	logger         domain.Logger
	clock          domain.Clock
}

// NewContainer instantiates a new container from the provided config object
func NewContainer(cfg *config, l domain.Logger, cl domain.Clock, rawTs *string) (*container, func() error, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config: %w", domain.ErrIsNil)
	}
	if l == nil {
		return nil, nil, fmt.Errorf("logger: %w", domain.ErrIsNil)
	}

	// setup db connection
	db, err := sqlConnectAndMigrate(cfg.MySQLURL, cfg.MigrationsURL, l)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect and migrate database: %w", err)
	}

	// parse templates
	tpl, err := domain.ParseTemplates("./service/views")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse templates: %w", err)
	}

	// instantiate email queue
	chEml := make(chan domain.Email)

	// TODO - replace with alt domain.EmailClient if api key is missing
	// instantiate email client
	var emlCl domain.EmailClient
	if cfg.MailgunAPIKey != "" {
		emlCl, err = mailgun.NewClient(cfg.MailgunAPIKey)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot instantiate mailgun client: %w", err)
		}
	}

	// new router
	rtr := mux.NewRouter()

	// TODO - replace with clock usage
	debugTs, err := parseTimeString(rawTs)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse debug time string: %w", err)
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

	// TODO - replace with alt domain.FootballDataSource if api token is missing
	// instantiate football-data.org client
	var fds domain.FootballDataSource
	if cfg.FootballDataAPIToken != "" {
		fds, err = footballdataorg.NewClient(cfg.FootballDataAPIToken, tc)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot instantiate football-data.org client: %w", err)
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
	ca, err := domain.NewCommunicationsAgent(er, epr, sr, chEml, tpl, sc, tc, rc)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot instantiate communications agent: %w", err)
	}
	ea, err := domain.NewEntryAgent(er, epr, sc)
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
	ta, err := domain.NewTokenAgent(tr)
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
		chEml,
		fds,
		rtr,
		debugTs,
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

func parseTimeString(str *string) (*time.Time, error) {
	if str == nil || *str == "" {
		// nothing to parse
		return nil, nil
	}

	strVal := *str

	var (
		parsed time.Time
		err    error
	)

	parsed, err = time.Parse("20060102150405", strVal)
	if err != nil {
		parsed, err = time.Parse("20060102", strVal)
		if err != nil {
			return nil, fmt.Errorf("cannot parse time string '%s': %w", strVal, err)
		}
	}

	return &parsed, nil
}
