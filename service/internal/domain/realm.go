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
	Site     RealmSite     `yaml:"site"`
}

// GetFullHomeURL returns the home page path appended to the realm origin
func (r Realm) GetFullHomeURL() string {
	return r.Site.Origin + r.Site.Paths.Home
}

// GetFullLeaderboardURL returns the leaderboard page path appended to the realm origin
func (r Realm) GetFullLeaderboardURL() string {
	return r.Site.Origin + r.Site.Paths.Leaderboard
}

// GetFullMyTableURL returns the my table page path appended to the realm origin
func (r Realm) GetFullMyTableURL() string {
	return r.Site.Origin + r.Site.Paths.MyTable
}

// GetMagicLoginURL generates a login URL using the provided Token
func (r Realm) GetMagicLoginURL(t *Token) string {
	tID := ""
	if t != nil {
		tID = "/" + t.ID
	}
	return r.Site.Origin + r.Site.Paths.Login + tID
}

// RealmConfig represents the core configuration of a Realm
type RealmConfig struct {
	Name     string `yaml:"name"`      // realm id stored in database for entries
	GameName string `yaml:"game_name"` // name of the game, referenced in transactional emails and html titles
	PIN      string `yaml:"pin"`       // pin to enter the game
	SeasonID string `yaml:"season_id"` // id of season to associate with the realm
}

// RealmContact represents the contact details of a realm
type RealmContact struct {
	EmailDoNotReply string `yaml:"email_do_not_reply"` // admin/sender email for transactional emails
	EmailProper     string `yaml:"email_proper"`       // formatted admin/sender email
	EmailSanitised  string `yaml:"email_sanitised"`    // sanitised admin/sender email
	SignOffName     string `yaml:"sign_off_name"`      // name to sign-off emails with
	SenderDomain    string `yaml:"sender_domain"`      // domain to issue transactional emails from
	SenderName      string `yaml:"sender_name"`        // sender/from name for transactional emails
}

// RealmEntryFee represents the entry fee settings for a realm
type RealmEntryFee struct {
	Amount    float32  `yaml:"amount"`    // entry payment numerical amount
	Breakdown []string `yaml:"breakdown"` // breakdown of entry fee to display on entry page
	Label     string   `yaml:"label"`     // entry payment formatted/display amount
}

// RealmFAQ defines the structure of a frequently-asked question
type RealmFAQ struct {
	Question string        `yaml:"question"`
	Answer   template.HTML `yaml:"answer"`
}

// RealmSite defines data points that are rendered as markup
type RealmSite struct {
	AnalyticsCode          string        `yaml:"analytics_code"`             // google analytics code
	Description            string        `yaml:"description"`                // content of og:description tag
	HomePageBannerImageURL string        `yaml:"home_page_banner_image_url"` // url for home page banner image
	HomePageHeading        string        `yaml:"home_page_heading"`          // heading to render on home page
	HomePageTagline        template.HTML `yaml:"home_page_tagline"`          // tagline to render on home page
	ImageAlt               string        `yaml:"image_alt"`                  // textual description of the image used for og:image
	ImageURL               string        `yaml:"image_url"`                  // url for og:image tag
	MenuBarIconURL         string        `yaml:"menu_bar_icon_url"`          // url for image rendered as icon in menu bar
	MenuBarTitle           string        `yaml:"menu_bar_title"`             // title to render inside menu bar
	Origin                 string        `yaml:"origin"`                     // url for og:url tag
	TwitterUsername        string        `yaml:"twitter_username"`           // twitter handle associated with site
	Paths                  RealmSitePaths
}

// RealmSitePaths store the paths to each page
type RealmSitePaths struct {
	FAQ         string
	Home        string
	Join        string
	Leaderboard string
	Login       string
	MyTable     string
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

	// populate site paths
	realm.Site.Paths = NewRealmSitePaths()

	return realm, nil
}

func NewRealmSitePaths() RealmSitePaths {
	return RealmSitePaths{
		FAQ:         "/faq",
		Home:        "/",
		Join:        "/join",
		Leaderboard: "/leaderboard",
		Login:       "/login",
		MyTable:     "/prediction",
	}
}
