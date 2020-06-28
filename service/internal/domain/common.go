package domain

import (
	"bytes"
	"github.com/LUSHDigital/core-mage/env"
	"github.com/kelseyhightower/envconfig"
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
	PIN      string
	SeasonID string
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
	FootballDataAPIToken string `envconfig:"FOOTBALLDATA_API_TOKEN" required:"true"`
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

	config.Realms = make(map[string]Realm)

	// next, let's populate all realms that are configured for the system
	for _, keyvalstring := range os.Environ() {
		if !strings.Contains(keyvalstring, "=") {
			// something's wrong, just skip to the next one...
			continue
		}

		split := strings.Split(keyvalstring, "=")
		key, val := split[0], split[1]

		// parse required values from key and val and determine if they are realm-related
		var rawRealmName, realmPIN, realmSeasonID string
		switch {
		case strings.HasSuffix(key, "_REALM_PIN"):
			// represents a realm PIN
			rawRealmName = strings.Split(key, "_REALM_PIN")[0]
			realmPIN = val
		case strings.HasSuffix(key, "_REALM_SEASON_ID"):
			// represents a realm season ID
			rawRealmName = strings.Split(key, "_REALM_SEASON_ID")[0]
			realmSeasonID = val
		}

		// did our key represent a realm-related item of data?
		if rawRealmName != "" {
			formattedRealmName := formatRealmNameFromRaw(rawRealmName)

			// retrieve realm in case we've already added some values previously
			realm := config.Realms[formattedRealmName]
			realm.Name = formattedRealmName

			switch {
			case realmPIN != "":
				// we now have a PIN for this realm
				realm.PIN = realmPIN
			case realmSeasonID != "":
				// we now have a season ID for this realm
				realm.SeasonID = realmSeasonID
			}

			// add the realm back to our config
			config.Realms[formattedRealmName] = realm
		}
	}

	return config
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
