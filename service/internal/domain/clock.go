package domain

import "time"

// Clock defines an interface for providing a timestamp
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock with the
type RealClock struct{}

// Now implements Clock
func (r *RealClock) Now() time.Time {
	return time.Now()
}
