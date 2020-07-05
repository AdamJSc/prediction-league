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
	"strconv"
	"strings"
	"time"
)

const authCookieName = "PL_AUTH"

func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}

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
func setAuthCookieValue(value string, w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:    authCookieName,
		Value:   value,
		Domain:  stripPort(r.Host),
		Expires: time.Now().Add(domain.TokenDurationInMinutes * time.Minute),
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
