package domain

import (
	"time"
)

// TimeFrame defines a timeframe represented by two timestamps
type TimeFrame struct {
	From  time.Time
	Until time.Time
}

// TODO - return error for failure context
// Valid determines whether the associated TimeFrame is chronologically sound
func (t TimeFrame) Valid() bool {
	switch {
	case t.From.Equal(time.Time{}) || t.Until.Equal(time.Time{}):
		// must have both From and Until
		return false
	case !t.From.Equal(time.Time{}) && !t.Until.Equal(time.Time{}):
		if !t.Until.After(t.From) {
			// Until must occur after From
			return false
		}
	}

	return true
}

// HasBegunBy determines whether the provided timestamp occurs before the start of the associated TimeFrame
func (t TimeFrame) HasBegunBy(ts time.Time) bool {
	return t.From.Before(ts) || t.From.Equal(ts)
}

// HasElapsedBy determines whether the provided timestamp occurs after the end of the associated TimeFrame
func (t TimeFrame) HasElapsedBy(ts time.Time) bool {
	return t.Until.Before(ts) || t.Until.Equal(ts)
}

// OverlapsWith determines whether the associated TimeFrame overlaps at any point with the provided TimeFrame
func (t TimeFrame) OverlapsWith(tf TimeFrame) bool {
	if tf.From.Equal(t.Until) || tf.Until.Equal(t.From) {
		// these timeframes are consecutive
		return false
	}

	tStartOverlaps := tf.HasBegunBy(t.From) && !tf.HasElapsedBy(t.From)
	tEndOverlaps := tf.HasBegunBy(t.Until) && !tf.HasElapsedBy(t.Until)

	tfStartOverlaps := t.HasBegunBy(tf.From) && !t.HasElapsedBy(tf.From)
	tfEndOverlaps := t.HasBegunBy(tf.Until) && !t.HasElapsedBy(tf.Until)

	return tStartOverlaps || tEndOverlaps || tfStartOverlaps || tfEndOverlaps
}

// BeginsWithin determines whether the associated TimeFrame begins within the provided TimeFrame
func (t TimeFrame) BeginsWithin(tf TimeFrame) bool {
	return !t.From.Before(tf.From) && t.HasBegunBy(tf.Until)
}

// EndsWithin determines whether the associated TimeFrame ends within the provided TimeFrame
func (t TimeFrame) EndsWithin(tf TimeFrame) bool {
	return !t.Until.Before(tf.From) && t.HasElapsedBy(tf.Until)
}

// SequencedTimeFrame represents a TimeFrame within the context of a wider sequence/schedule of TimeFrames
type SequencedTimeFrame struct {
	Count   int
	Total   int
	Current *TimeFrame
	Next    *TimeFrame
}
