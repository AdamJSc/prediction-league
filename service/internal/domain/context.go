package domain

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

const (
	ctxKeyRealm         = "REALM"
	ctxKeyRealmPIN      = "REALM_PIN"
	ctxKeyRealmSeasonID = "REALM_SEASON_ID"
	ctxKeyGuardValue    = "GUARD_VALUE"

	envKeyAdminBasicAuth = "ADMIN_BASIC_AUTH"
)

// Guard represents an arbitrary guard that can be used by agent methods
// to determine whether or not an operation should continue
type Guard struct {
	Attempt string
}

// SetAttempt sets the value that attempts to match the target
// that will eventually be assessed by an agent method
func (g Guard) SetAttempt(attempt string) {
	g.Attempt = attempt
}

// AttemptMatchesTarget returns true if provided target matches
// the attempt value already on the guard, otherwise false
func (g Guard) AttemptMatchesTarget(target string) bool {
	if g.Attempt == "" || target == "" {
		return false
	}
	return g.Attempt == target
}

// Context wraps a standard context for the purpose of additional helper methods
type Context struct {
	context.Context
	Guard Guard
	Realm Realm
}

// setString sets a context value whose type is a string
func (c *Context) setString(key string, value string) {
	c.Context = context.WithValue(c.Context, key, value)
}

// getString retrieves a context value whose type is a string
func (c *Context) getString(key string) string {
	var value string
	ctxValue := c.Context.Value(key)

	if ctxValue != nil {
		value = ctxValue.(string)
	}

	return value
}

// SetRealm sets the context's Realm value
func (c *Context) SetRealm(realm string) {
	c.setString(ctxKeyRealm, realm)
}

// GetRealm retrieves the context's Realm value
func (c *Context) GetRealm() string {
	return c.getString(ctxKeyRealm)
}

// SetRealmPIN sets the context's Realm PIN value
func (c *Context) SetRealmPIN(pin string) {
	c.setString(ctxKeyRealmPIN, pin)
}

// GetRealmPIN retrieves the context's Realm PIN value
func (c *Context) GetRealmPIN() string {
	return c.getString(ctxKeyRealmPIN)
}

// SetRealmSeasonID sets the context's Realm Season ID value
func (c *Context) SetRealmSeasonID(seasonID string) {
	c.setString(ctxKeyRealmSeasonID, seasonID)
}

// GetRealmSeasonID retrieves the context's Realm Season ID value
func (c *Context) GetRealmSeasonID() string {
	return c.getString(ctxKeyRealmSeasonID)
}

// SetGuardValue sets an arbitrary guard value on the context that
// can be used by an agent method to determine access to the request
func (c *Context) SetGuardValue(guardValue string) {
	c.setString(ctxKeyGuardValue, guardValue)
}

// GetGuardValue retrieves an arbitrary guard value on the context that
// can be used by an agent method to determine access to the request
func (c *Context) GetGuardValue() string {
	return c.getString(ctxKeyGuardValue)
}

// NewContext returns a new Context
func NewContext() Context {
	return Context{Context: context.Background()}
}

// ContextFromRequest extracts data from a given request object and returns a domain object Context
func ContextFromRequest(r *http.Request, config Config) Context {
	ctx := NewContext()

	// realm name is host (strip port)
	realmName := strings.Trim(strings.Split(r.Host, ":")[0], " ")

	// see if we can find this realm in our config
	if realm, ok := config.Realms[realmName]; ok {
		ctx.Realm = realm
	}

	// set basic auth username/password requirements
	var userPass []byte
	authHeader := r.Header.Get("Authorization")
	headerParts := strings.Split(authHeader, "Basic ")
	if len(headerParts) == 2 {
		userPass, _ = base64.StdEncoding.DecodeString(headerParts[1])
	}
	ctx.Context = context.WithValue(ctx.Context, envKeyAdminBasicAuth, string(userPass))

	return ctx
}

// validateRealmPIN checks that the supplied PIN matches the Realm PIN added to the Context
func validateRealmPIN(ctx Context, pin string) error {
	realmPIN := ctx.GetRealmPIN()
	if realmPIN == "" || realmPIN != pin {
		return UnauthorizedError{errors.New("unauthorized")}
	}

	return nil
}
