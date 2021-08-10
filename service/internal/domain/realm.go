package domain

import (
	"fmt"
)

// Realm represents a realm in which the system has been configured to run
type Realm struct {
	Name          string
	Origin        string        `yaml:"origin"`
	SenderDomain  string        `yaml:"sender_domain"`
	Contact       RealmContact  `yaml:"contact"`
	PIN           string        `yaml:"pin"`
	SeasonID      string        `yaml:"season_id"`
	EntryFee      RealmEntryFee `yaml:"entry_fee"`
	AnalyticsCode string        `yaml:"analytics_code"`
	Image         string        `yaml:"image"`
}

// RealmContact represents the contact details of a realm
type RealmContact struct {
	Name            string `yaml:"name"`
	EmailProper     string `yaml:"email_proper"`
	EmailSanitised  string `yaml:"email_sanitised"`
	EmailDoNotReply string `yaml:"email_do_not_reply"`
}

// RealmEntryFee represents the entry fee settings for a realm
type RealmEntryFee struct {
	Amount    float32  `yaml:"amount"`
	Label     string   `yaml:"label"`
	Breakdown []string `yaml:"breakdown"`
}

// RealmCollection is map of Realm
type RealmCollection map[string]Realm

// GetByName retrieves a matching Realm from the collection by its name
func (rc RealmCollection) GetByName(rName string) (Realm, error) {
	for id, r := range rc {
		if id == rName {
			return r, nil
		}
	}

	return Realm{}, NotFoundError{fmt.Errorf("realm name %s: not found", rName)}
}
