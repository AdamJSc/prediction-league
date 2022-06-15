package app

import (
	"fmt"
	"prediction-league/service/internal/domain"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config encapsulate the required options
type Config struct {
	ServicePort          string `envconfig:"SERVICE_PORT" required:"true"`
	MySQLURL             string `envconfig:"MYSQL_URL" required:"true"`
	MigrationsURL        string `envconfig:"MIGRATIONS_URL" required:"true"`
	AdminBasicAuth       string `envconfig:"ADMIN_BASIC_AUTH" required:"true"`
	LogLevel             string `envconfig:"LOG_LEVEL" required:"true"`
	FootballDataAPIToken string `envconfig:"FOOTBALLDATA_API_TOKEN" required:"true"`
	PayPalClientID       string `envconfig:"PAYPAL_CLIENT_ID" required:"true"`
	MailgunAPIKey        string `envconfig:"MAILGUN_API_KEY" required:"true"`
	BuildVersion         string
	BuildTimestamp       string
}

// ConfigOption defines a type of function for modifying a Config object
type ConfigOption func(config *Config) error

// NewConfigFromOptions returns the Config object produce by the provided functional options
func NewConfigFromOptions(options ...ConfigOption) (*Config, error) {
	config := &Config{}

	for _, opt := range options {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// NewLoadEnvConfigOption enriches Config based on the provided env files
func NewLoadEnvConfigOption(logger domain.Logger, paths ...string) ConfigOption {
	return func(config *Config) error {
		// attempt to load env vars from all provided paths
		for _, fpath := range paths {
			if err := godotenv.Load(fpath); err != nil {
				logger.Infof("skipping env file '%s': not found", fpath)
			}
		}

		// parse config
		if err := envconfig.Process("", config); err != nil {
			return fmt.Errorf("cannot process env config: %w", err)
		}

		return nil
	}
}

// NewBuildVersionConfigOption sets the provided build version value on the Config object
func NewBuildVersionConfigOption(buildVersion string) ConfigOption {
	return func(config *Config) error {
		config.BuildVersion = buildVersion
		return nil
	}
}

// NewBuildTimestampConfigOption sets the provided build timestamp value on the Config object
func NewBuildTimestampConfigOption(buildTimestamp string) ConfigOption {
	return func(config *Config) error {
		config.BuildTimestamp = buildTimestamp
		return nil
	}
}
