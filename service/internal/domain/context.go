package domain

import (
	"context"
	"time"
)

type ContextKey string

const (
	contextKeyTimestamp        ContextKey = "TIMESTAMP"
	contextKeyGuard            ContextKey = "GUARD"
	contextKeyRealm            ContextKey = "REALM"
	contextKeyBasicAuthSuccess ContextKey = "BASIC_AUTH_SUCCESS"
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

// AttemptMatches returns true if provided target matches
// the attempt value already on the guard, otherwise false
func (g *Guard) AttemptMatches(target string) bool {
	if g.Attempt == "" || target == "" {
		return false
	}
	return g.Attempt == target
}

// NewContext returns a new context with domain-specific default values
func NewContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	ctx = context.WithValue(ctx, contextKeyGuard, &Guard{})
	ctx = context.WithValue(ctx, contextKeyRealm, &Realm{})

	var basicAuth bool
	ctx = context.WithValue(ctx, contextKeyBasicAuthSuccess, &basicAuth)

	var time = time.Now()
	ctx = context.WithValue(ctx, contextKeyTimestamp, &time)

	return ctx, cancel
}

// TimestampFromContext retrieves a timestamp override if set on the provided context, otherwise returns current timestamp
func TimestampFromContext(ctx context.Context) time.Time {
	var ts = time.Now()

	val := ctx.Value(contextKeyTimestamp)

	switch val.(type) {
	case time.Time:
		return val.(time.Time)
	}

	return ts
}

// GuardFromContext retrieves the Guard from the provided context
func GuardFromContext(ctx context.Context) *Guard {
	var g = &Guard{}

	val := ctx.Value(contextKeyGuard)
	switch val.(type) {
	case *Guard:
		return val.(*Guard)
	default:
		ctx = context.WithValue(ctx, contextKeyGuard, g)
	}

	return g
}

// RealmFromContext retrieves the Realm from the provided context
func RealmFromContext(ctx context.Context) *Realm {
	var r = &Realm{}

	// check that we are dealing with a valid context
	if ctx == nil {
		ctx = context.WithValue(ctx, contextKeyRealm, r)
		return r
	}

	val := ctx.Value(contextKeyRealm)
	switch val.(type) {
	case *Realm:
		return val.(*Realm)
	default:
		ctx = context.WithValue(ctx, contextKeyRealm, r)
	}

	return r
}

// SetBasicAuthSuccessfulOnContext sets a successful Basic Auth attempt on the provided context
func SetBasicAuthSuccessfulOnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyBasicAuthSuccess, true)
}

// IsBasicAuthSuccessful determines whether a successful Basic Auth attempt exists on the provided context
func IsBasicAuthSuccessful(ctx context.Context) bool {
	val := ctx.Value(contextKeyBasicAuthSuccess)

	switch val.(type) {
	case bool:
		return val.(bool)
	}

	return false
}

// SetTimestampOnContext sets the provided timestamp on the provided context
func SetTimestampOnContext(ctx context.Context, ts time.Time) context.Context {
	return context.WithValue(ctx, contextKeyTimestamp, ts)
}
