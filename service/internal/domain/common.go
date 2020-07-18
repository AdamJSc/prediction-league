package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/LUSHDigital/core-mage/env"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"prediction-league/service/internal/views"
	"strings"
	"time"
)

// Realm represents a realm in which the system has been configured to run
type Realm struct {
	Name     string
	PIN      string        `yaml:"pin"`
	SeasonID string        `yaml:"season_id"`
	EntryFee RealmEntryFee `yaml:"entry_fee"`
}

// RealmEntryFee represents the entry fee settings for a realm
type RealmEntryFee struct {
	Amount    float32  `yaml:"amount"`
	Label     string   `yaml:"label"`
	Breakdown []string `yaml:"breakdown"`
}

// formatRealmNameFromRaw converts a raw realm name (as the prefix to an env key) to a formatted realm name
func formatRealmNameFromRaw(rawRealmName string) string {
	formattedName := rawRealmName

	formattedName = strings.Trim(formattedName, " ")
	formattedName = strings.Replace(formattedName, "_", ".", -1)
	formattedName = strings.ToLower(formattedName)

	return formattedName
}

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
	Realms               map[string]Realm
}

// MustLoadConfigFromEnvPaths loads provided env paths and instantiates a new default config
func MustLoadConfigFromEnvPaths(paths ...string) Config {
	// attempt to load all provided env paths
	for _, path := range paths {
		env.Load(path)
	}

	// ensure that config parses correctly
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	config.Realms = mustParseRealmsFromPath(fmt.Sprintf("./data/realms.yml"))

	return config
}

// mustParseRealmsFromPath parses the realms from the contents of the YAML file at the provided path
func mustParseRealmsFromPath(yamlPath string) map[string]Realm {
	contents, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		log.Fatal(err)
	}

	var payload struct {
		Realms map[string]Realm `yaml:"realms"`
	}

	// parse file contents
	if err := yaml.Unmarshal(contents, &payload); err != nil {
		log.Fatal(err)
	}

	// populate names of realms with map key
	for key := range payload.Realms {
		r := payload.Realms[key]
		r.Name = key
		payload.Realms[key] = r
	}

	return payload.Realms
}

// generateRandomAlphaNumericString returns a randomised string of given length
func generateRandomAlphaNumericString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}

	return string(b)
}

var templateFunctions = template.FuncMap{
	"timestamp_as_unix": func(ts time.Time) int64 {
		var emptyTime time.Time

		if ts.Equal(emptyTime) {
			return 0
		}
		return ts.Unix()
	},
	"jsonify_strings": func(input []string) string {
		bytes, err := json.Marshal(input)
		if err != nil {
			return ""
		}

		return string(bytes)
	},
}

// ParseTemplates parses our HTML templates and returns them collectively for use
func ParseTemplates() *views.Templates {
	// prepare the templates
	tpl := template.New("prediction-league").Funcs(templateFunctions)

	walkPathAndParseTemplates(tpl, "./service/views/html")

	// return our wrapped template struct
	return &views.Templates{Template: tpl}
}

// walkPathAndParseTemplates recursively parses templates within a given top-level directory
func walkPathAndParseTemplates(tpl *template.Template, path string) {
	// walk through our views folder and parse each item to pack the assets
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// we already have an error from a recursive call, so just return with that
		if err != nil {
			return err
		}

		// skip directories, we're only interested in files
		if info.IsDir() {
			return nil
		}

		// open the current file
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		file := bytes.NewReader(contents)

		// copy file contents as a byte stream and then parse as a template
		var b bytes.Buffer
		if _, err = io.Copy(&b, file); err != nil {
			return err
		}
		tpl = template.Must(tpl.Parse(b.String()))

		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}
