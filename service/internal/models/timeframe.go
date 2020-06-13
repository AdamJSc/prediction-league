package models

import "time"

var emptyTime time.Time

// TimeFrame defines a timeframe represented by two timestamps
type TimeFrame struct {
	From  time.Time
	Until time.Time
}

// Valid determines whether the associated TimeFrame is chronologically sound
func (t TimeFrame) Valid() bool {
	switch {
	case t.From.Equal(emptyTime) || t.Until.Equal(emptyTime):
		// must have both From and Until
		return false
	case !t.From.Equal(emptyTime) && !t.Until.Equal(emptyTime):
		if !t.Until.After(t.From) {
			// Until must occur after From
			return false
		}
	}

	return true
}

// HasBegunBy determines whether the provided timestamp occurs before the start of the associated TimeFrame
func (t TimeFrame) HasBegunBy(ts time.Time) bool {
	if t.From.Equal(emptyTime) || !ts.Before(t.From) {
		return true
	}

	return false
}

// HasElapsedBy determines whether the provided timestamp occurs after the end of the associated TimeFrame
func (t TimeFrame) HasElapsedBy(ts time.Time) bool {
	if t.Until.Equal(emptyTime) || !ts.After(t.Until) {
		return false
	}

	return true
}
