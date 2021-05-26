package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type DependencyInjector interface {
	ConfigInjector
	EmailClientInjector
	EmailQueueInjector
	RouterInjector
	TemplateInjector
	DebugTimestampInjector
	StandingsRepoInjector
	EntryRepoInjector
	EntryPredictionRepoInjector
	ScoredEntryPredictionRepoInjector
	TokenRepoInjector
	SeasonsInjector
	TeamsInjector
	RealmsInjector
	ClockInjector
	LoggerInjector
}

type ConfigInjector interface{ Config() *config }
type EmailClientInjector interface{ EmailClient() domain.EmailClient }
type EmailQueueInjector interface{ EmailQueue() chan domain.Email }
type RouterInjector interface{ Router() *mux.Router }
type TemplateInjector interface{ Template() *domain.Templates }
type DebugTimestampInjector interface{ DebugTimestamp() *time.Time }
type StandingsRepoInjector interface {
	StandingsRepo() domain.StandingsRepository
}
type EntryRepoInjector interface {
	EntryRepo() domain.EntryRepository
}
type EntryPredictionRepoInjector interface {
	EntryPredictionRepo() domain.EntryPredictionRepository
}
type ScoredEntryPredictionRepoInjector interface {
	ScoredEntryPredictionRepo() domain.ScoredEntryPredictionRepository
}
type TokenRepoInjector interface {
	TokenRepo() domain.TokenRepository
}
type SeasonsInjector interface {
	Seasons() domain.SeasonCollection
}
type TeamsInjector interface {
	Teams() domain.TeamCollection
}
type RealmsInjector interface {
	Realms() domain.RealmCollection
}
type ClockInjector interface {
	Clock() domain.Clock
}
type LoggerInjector interface {
	Logger() domain.Logger
}

const authCookieName = "PL_AUTH"

// closeBody closes the body of the provided request
func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}

// getRouteParam retrieves the provided named route parameter from the provided request object
func getRouteParam(r *http.Request, name string, value interface{}) error {
	val, ok := mux.Vars(r)[name]
	if !ok {
		return fmt.Errorf("invalid param: %s", name)
	}

	switch value.(type) {
	case *string:
		typed := value.(*string)
		*typed = val
	case *int:
		typed := value.(*int)
		int, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*typed = int
	}
	return nil
}

// contextFromRequest extracts data from a given request object and returns an inflated context
func contextFromRequest(r *http.Request, c *HTTPAppContainer) (context.Context, context.CancelFunc, error) {
	ctx, cancel := domain.NewContext()

	config := c.Config()
	realms := c.Realms()
	debugTs := c.DebugTimestamp()

	// realm name is host (strip port)
	realmName := stripPort(r.Host)

	// see if we can find this realm in our config
	realm, err := realms.GetByName(realmName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get realm with id '%s': %w", realmName, err)
	}

	ctxRealm := domain.RealmFromContext(ctx)
	*ctxRealm = realm

	// see if request contains any basic auth credentials
	var userPass []byte
	authHeader := r.Header.Get("Authorization")
	split := strings.Split(authHeader, "Basic ")
	if len(split) == 2 {
		userPass, _ = base64.StdEncoding.DecodeString(split[1])
		// if username and password supplied, check if it matches the expected from env/config
		if string(userPass) == config.AdminBasicAuth {
			ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		}
	}

	// if debug timestamp has been provided, add this to context
	var ts = time.Now()
	if debugTs != nil {
		ts = *debugTs
	}

	ctx = domain.SetTimestampOnContext(ctx, ts)

	return ctx, cancel, nil
}

// stripPort removes the port suffix from the provided host string
func stripPort(host string) string {
	return strings.Trim(strings.Split(host, ":")[0], " ")
}

// setAuthCookieValue sets an authorization cookie with the provided value
func setAuthCookieValue(value string, w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:    authCookieName,
		Value:   value,
		Domain:  stripPort(r.Host),
		Expires: time.Now().Add(domain.TokenValidityDuration[domain.TokenTypeAuth]),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
}

// getAuthCookieValue retrieves the current value of the authorization cookie
func getAuthCookieValue(r *http.Request) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == authCookieName {
			return cookie.Value
		}
	}

	return ""
}

// isLoggedIn determines whether the provided request represents a logged in user
func isLoggedIn(r *http.Request) bool {
	if cookieValue := getAuthCookieValue(r); cookieValue == "" {
		return false
	}
	return true
}

// newPage creates a new base page from the provided arguments
func newPage(r *http.Request, c *HTTPAppContainer, title, activePage, bannerTitle string, data interface{}) *view.Base {
	// ignore error because if the realm doesn't exist the other agent methods will prevent core functionality anyway
	// we only need the realm for populating a couple of trivial attributes which will be the least of our worries if any issues...
	ctx, cancel, _ := contextFromRequest(r, c)
	defer cancel()

	realm := domain.RealmFromContext(ctx)

	return &view.Base{
		Title:                 title,
		BannerTitle:           template.HTML(bannerTitle),
		ActivePage:            activePage,
		IsLoggedIn:            isLoggedIn(r),
		SupportEmailPlainText: realm.Contact.EmailSanitised,
		SupportEmailFormatted: realm.Contact.EmailProper,
		RealmName:             realm.Name,
		RunningVersion:        c.Config().RunningVersion,
		VersionTimestamp:      c.Config().VersionTimestamp,
		AnalyticsCode:         realm.AnalyticsCode,
		Data:                  data,
	}
}

// retrieveEntryByEmailOrNickname retrieves an entry from the provided input, whether this pertains to an email address or nickname
func retrieveEntryByEmailOrNickname(ctx context.Context, emailOrNickname, seasonID, realmName string, entryAgent *domain.EntryAgent) (*domain.Entry, error) {
	// see if we can retrieve by email
	entry, err := entryAgent.RetrieveEntryByEntrantEmail(ctx, emailOrNickname, seasonID, realmName)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// see if we can retrieve by nickname
			entry, err := entryAgent.RetrieveEntryByEntrantNickname(ctx, emailOrNickname, seasonID, realmName)
			if err != nil {
				return nil, err
			}

			return &entry, nil
		default:
			return nil, err
		}
	}

	return &entry, nil
}
