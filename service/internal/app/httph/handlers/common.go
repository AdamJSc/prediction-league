package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/domain"
	"strings"
	"time"
)

const authCookieName = "PL_AUTH"

func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}

func getRouteParam(r *http.Request, name string, value *string) error {
	val, ok := mux.Vars(r)[name]
	if !ok {
		return fmt.Errorf("invalid param: %s", name)
	}
	*value = val
	return nil
}

// contextFromRequest extracts data from a given request object and returns an inflated context
func contextFromRequest(r *http.Request, c *httph.HTTPAppContainer) (context.Context, context.CancelFunc, error) {
	ctx, cancel := domain.NewContext()

	config := c.Config()
	debugTs := c.DebugTimestamp()

	// realm name is host (strip port)
	realmName := stripPort(r.Host)

	// see if we can find this realm in our config
	realm, ok := config.Realms[realmName]
	if !ok {
		return nil, nil, errors.New("realm not configured")
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

func stripPort(host string) string {
	return strings.Trim(strings.Split(host, ":")[0], " ")
}

// setAuthCookieValue sets an authorization cookie with the provided value
func setAuthCookieValue(w http.ResponseWriter, value string, domain string) {
	cookie := &http.Cookie{
		Name:    authCookieName,
		Value:   value,
		Domain:  domain,
		Expires: time.Now().Add(60 * time.Minute),
	}
	http.SetCookie(w, cookie)
}

// getAuthCookieValue retrieves the current value of the authorization cookie
func getAuthCookieValue(r *http.Request) (string, error) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == authCookieName {
			return cookie.Value, nil
		}
	}

	return "", errors.New("auth cookie not set")
}

// isLoggedIn determines whether the provided request represents a logged in user
func isLoggedIn(r *http.Request) bool {
	if _, err := getAuthCookieValue(r); err != nil {
		return false
	}
	return true
}
