package models

import "time"

// TimeFrame defines a timeframe represented by two timestamps
type TimeFrame struct {
	From  *time.Time
	Until *time.Time
}

// Valid determines whether the associated TimeFrame is chronologically sound
func (t TimeFrame) Valid() bool {
	switch {
	case t.From == nil && t.Until == nil:
		// must have at least a From or an Until
		return false
	case t.From != nil && t.Until != nil:
		if !t.Until.After(*t.From) {
			// Until must occur after From
			return false
		}
	}

	return true
}

// HasBegunBy determines whether the provided timestamp occurs before the start of the associated TimeFrame
func (t TimeFrame) HasBegunBy(ts time.Time) bool {
	if t.From == nil || !ts.Before(*t.From) {
		return true
	}

	return false
}

// HasElapsedBy determines whether the provided timestamp occurs after the end of the associated TimeFrame
func (t TimeFrame) HasElapsedBy(ts time.Time) bool {
	if t.Until == nil || !ts.After(*t.Until) {
		return false
	}

	return true
}
