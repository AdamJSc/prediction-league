package domain

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Guard represents an arbitrary guard that can be used by agent methods
// to determine whether or not an operation should continue
type Guard struct {
	Attempt string
}

// SetAttempt sets the value that attempts to match the target
// that will eventually be assessed by an agent method
func (g *Guard) SetAttempt(attempt string) {
	g.Attempt = attempt
}

// AttemptMatchesTarget returns true if provided target matches
// the attempt value already on the guard, otherwise false
func (g *Guard) AttemptMatchesTarget(target string) bool {
	if g.Attempt == "" || target == "" {
		return false
	}
	return g.Attempt == target
}

// Context wraps a standard context for the purpose of additional helper methods
type Context struct {
	context.Context
	Guard               Guard
	Realm               Realm
	BasicAuthSuccessful bool
}

// NewContext returns a new Context
func NewContext() Context {
	return Context{Context: context.Background()}
}

// ContextFromRequest extracts data from a given request object and returns a domain object Context
func ContextFromRequest(r *http.Request, config Config) (Context, context.CancelFunc, error) {
	domainCtx := NewContext()

	// set timeout limit
	ctxWithTimeout, cancel := context.WithTimeout(domainCtx.Context, 10*time.Second)
	domainCtx.Context = ctxWithTimeout

	// realm name is host (strip port)
	realmName := strings.Trim(strings.Split(r.Host, ":")[0], " ")

	// see if we can find this realm in our config
	realm, ok := config.Realms[realmName]
	if !ok {
		return Context{}, nil, errors.New("realm not configured")
	}

	domainCtx.Realm = realm

	// see if request contains any basic auth credentials
	var userPass []byte
	authHeader := r.Header.Get("Authorization")
	split := strings.Split(authHeader, "Basic ")
	if len(split) == 2 {
		userPass, _ = base64.StdEncoding.DecodeString(split[1])
		// if username and password supplied, check if it matches the expected from env/config
		if string(userPass) == config.AdminBasicAuth {
			domainCtx.BasicAuthSuccessful = true
		}
	}

	return domainCtx, cancel, nil
}
