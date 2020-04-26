package domain

import (
	"fmt"
	"github.com/LUSHDigital/core-mage/env"
	"github.com/kelseyhightower/envconfig"
	"log"
	"math/rand"
	"os"
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
	ServicePort    string `envconfig:"SERVICE_PORT" required:"true"`
	MySQLURL       string `envconfig:"MYSQL_URL" required:"true"`
	MigrationsURL  string `envconfig:"MIGRATIONS_URL" required:"true"`
	AdminBasicAuth string `envconfig:"ADMIN_BASIC_AUTH" required:"true"`
	Realms         map[string]Realm
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

// dbWhereStmt returns the WHERE clause portion of an SQL statement as a string, plus the parameters to
// pass to the operation, from a given map of criteria to query on
func dbWhereStmt(criteria map[string]interface{}, matchAny bool) (stmt string, params []interface{}) {
	var conditions []string

	for field, value := range criteria {
		conditions = append(conditions, fmt.Sprintf("(%s = ?)", field))
		params = append(params, value)
	}

	comparison := " AND "
	if matchAny {
		comparison = " OR "
	}

	if len(conditions) > 0 {
		stmt = `WHERE ` + strings.Join(conditions, comparison)
	}

	return stmt, params
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

// getDBFieldsStringFromFields returns a statement-ready string of fields names
func getDBFieldsStringFromFields(fields []string) string {
	return strings.Join(fields, ", ")
}

// getDBFieldsWithEqualsPlaceholdersStringFromFields returns a statement-ready string of fields names with "equals value" placeholders
func getDBFieldsWithEqualsPlaceholdersStringFromFields(fields []string) string {
	var fieldsWithEqualsPlaceholders []string

	for _, field := range fields {
		fieldsWithEqualsPlaceholders = append(fieldsWithEqualsPlaceholders, fmt.Sprintf("%s = ?", field))
	}

	return strings.Join(fieldsWithEqualsPlaceholders, ", ")
}
