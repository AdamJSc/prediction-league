package domain

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
)

// Config represents a struct of required config options
type Config struct {
	ServicePort          string `envconfig:"SERVICE_PORT" required:"true"`
	MySQLURL             string `envconfig:"MYSQL_URL" required:"true"`
	MigrationsURL        string `envconfig:"MIGRATIONS_URL" required:"true"`
	AdminBasicAuth       string `envconfig:"ADMIN_BASIC_AUTH" required:"true"`
	RunningVersion       string `envconfig:"RUNNING_VERSION" required:"true"`
	VersionTimestamp     string `envconfig:"VERSION_TIMESTAMP" required:"true"`
	FootballDataAPIToken string `envconfig:"FOOTBALLDATA_API_TOKEN" required:"true"`
	PayPalClientID       string `envconfig:"PAYPAL_CLIENT_ID" required:"true"`
	MailgunAPIKey        string `envconfig:"MAILGUN_API_KEY" required:"true"`
	Realms               map[string]Realm
}

// MustLoadConfigFromEnvPaths loads provided env paths and instantiates a new default config
func MustLoadConfigFromEnvPaths(l Logger, paths ...string) Config {
	// attempt to load all provided env paths
	loadEnvFromPaths(l, paths...)

	// ensure that config parses correctly
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	config.Realms = mustParseRealmsFromPath(fmt.Sprintf("./data/realms.yml"))

	if config.PayPalClientID == "" {
		log.Println("missing config: paypal... entry signup payment step will be skipped...")
	}

	return config
}

func loadEnvFromPaths(l Logger, paths ...string) {
	for _, fpath := range paths {
		if err := godotenv.Load(fpath); err != nil {
			l.Infof("could not load environment file: %s: skipping...", fpath)
		}
	}
}
