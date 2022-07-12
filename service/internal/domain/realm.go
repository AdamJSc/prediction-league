package domain

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/gomarkdown/markdown"
	"gopkg.in/yaml.v2"
)

// Realm represents an instance of the game, often pertaining to the domain on which the server is accessible
type Realm struct {
	Config   RealmConfig   `yaml:"config"`
	Contact  RealmContact  `yaml:"contact"`
	EntryFee RealmEntryFee `yaml:"entry_fee"`
	FAQs     []RealmFAQ    `yaml:"faqs"`
}

// RealmConfig represents the core configuration of a Realm
type RealmConfig struct {
	Name            string `yaml:"name"`
	AnalyticsCode   string `yaml:"analytics_code"`
	HomePageHeading string `yaml:"home_page_heading"`
	HomePageTagline string `yaml:"home_page_tagline"`
	Image           string `yaml:"image"`
	MenuTitle       string `yaml:"menu_title"`
	Origin          string `yaml:"origin"`
	PIN             string `yaml:"pin"`
	SeasonID        string `yaml:"season_id"`
}

// RealmContact represents the contact details of a realm
type RealmContact struct {
	EmailDoNotReply string `yaml:"email_do_not_reply"`
	EmailProper     string `yaml:"email_proper"`
	EmailSanitised  string `yaml:"email_sanitised"`
	Name            string `yaml:"name"`
	SenderDomain    string `yaml:"sender_domain"`
}

// RealmEntryFee represents the entry fee settings for a realm
type RealmEntryFee struct {
	Amount    float32  `yaml:"amount"`
	Breakdown []string `yaml:"breakdown"`
	Label     string   `yaml:"label"`
}

// RealmFAQ defines the structure of a frequently-asked question
type RealmFAQ struct {
	Question string        `yaml:"question"`
	Answer   template.HTML `yaml:"answer"`
}

// RealmCollection is slice of Realm
type RealmCollection []Realm

// GetByName returns the Realm that matches the provided name
func (rc RealmCollection) GetByName(name string) (Realm, error) {
	for _, r := range rc {
		if name == r.Config.Name {
			return r, nil
		}
	}

	return Realm{}, NotFoundError{fmt.Errorf("realm name '%s': not found", name)}
}

// GetRealmCollection returns the required RealmCollection
func GetRealmCollection() (RealmCollection, error) {
	dirPath := filepath.Join("data", "realms")

	infos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read dir '%s': %w", dirPath, err)
	}

	realmCollection := make(RealmCollection, 0)

	for _, info := range infos {
		if info.IsDir() {
			// parse realm from directory
			realmDirPath := filepath.Join(dirPath, info.Name())

			realm, err := parseRealmFromDir(realmDirPath)
			if err != nil {
				return nil, fmt.Errorf("cannot parse realm from dir '%s': %w", realmDirPath, err)
			}

			realmCollection = append(realmCollection, realm)
		}
	}

	return realmCollection, nil
}

// parseRealmFromDir parses the files in the provided directory and returns a RealmCollection
func parseRealmFromDir(dirPath string) (Realm, error) {
	var realm Realm

	filePaths := []string{
		"main.yml",
		"faqs.yml",
	}

	for _, filePath := range filePaths {
		fullPath := filepath.Join(dirPath, filePath)

		b, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return Realm{}, fmt.Errorf("cannot read file '%s': %w", filePath, err)
		}

		if err := yaml.Unmarshal(b, &realm); err != nil {
			return Realm{}, fmt.Errorf("cannot unmarshal realm from file '%s': %w", filePath, err)
		}
	}

	// convert markdown faq answers to html
	for idx, faq := range realm.FAQs {
		asHTML := markdown.ToHTML([]byte(faq.Answer), nil, nil)
		realm.FAQs[idx].Answer = template.HTML(asHTML)
	}

	return realm, nil
}
