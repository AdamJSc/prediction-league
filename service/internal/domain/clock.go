package domain

import "time"

// Clock defines an interface for providing a timestamp
type Clock interface {
	Now() time.Time
}

// RealClock provides the real-world time
type RealClock struct{}

// Now implements Clock
func (r *RealClock) Now() time.Time {
	return time.Now()
}

// FrozenClock provides a static/pre-determined time object
type FrozenClock struct{ time.Time }

// Now implements Clock
func (f *FrozenClock) Now() time.Time {
	return f.Time
}
