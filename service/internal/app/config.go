package app

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"prediction-league/service/internal/domain"
)

// Config encapsulate the required options
type Config struct {
	ServicePort          string `envconfig:"SERVICE_PORT" required:"true"`
	MySQLURL             string `envconfig:"MYSQL_URL" required:"true"`
	MigrationsURL        string `envconfig:"MIGRATIONS_URL" required:"true"`
	AdminBasicAuth       string `envconfig:"ADMIN_BASIC_AUTH" required:"true"`
	RunningVersion       string `envconfig:"RUNNING_VERSION" required:"true"`
	VersionTimestamp     string `envconfig:"VERSION_TIMESTAMP" required:"true"`
	LogLevel             string `envconfig:"LOG_LEVEL" required:"true"`
	FootballDataAPIToken string `envconfig:"FOOTBALLDATA_API_TOKEN" required:"true"`
	PayPalClientID       string `envconfig:"PAYPAL_CLIENT_ID" required:"true"`
	MailgunAPIKey        string `envconfig:"MAILGUN_API_KEY" required:"true"`
}

// NewConfigFromEnvPaths loads provided env paths and instantiates a new default config
func NewConfigFromEnvPaths(l domain.Logger, paths ...string) (*Config, error) {
	// attempt to load env vars from all provided paths
	for _, fpath := range paths {
		if err := godotenv.Load(fpath); err != nil {
			l.Infof("skipping env file '%s': not found", fpath)
		}
	}

	// parse config
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("cannot process env config: %w", err)
	}

	return cfg, nil
}
