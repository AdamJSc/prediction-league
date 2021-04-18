package domain

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

// Realm represents a realm in which the system has been configured to run
type Realm struct {
	Name         string
	Origin       string `yaml:"origin"`
	SenderDomain string `yaml:"sender_domain"`
	Contact      struct {
		Name            string `yaml:"name"`
		EmailProper     string `yaml:"email_proper"`
		EmailSanitised  string `yaml:"email_sanitised"`
		EmailDoNotReply string `yaml:"email_do_not_reply"`
	} `yaml:"contact"`
	PIN           string        `yaml:"pin"`
	SeasonID      string        `yaml:"season_id"`
	EntryFee      RealmEntryFee `yaml:"entry_fee"`
	AnalyticsCode string        `yaml:"analytics_code"`
}

// RealmEntryFee represents the entry fee settings for a realm
type RealmEntryFee struct {
	Amount    float32  `yaml:"amount"`
	Label     string   `yaml:"label"`
	Breakdown []string `yaml:"breakdown"`
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
