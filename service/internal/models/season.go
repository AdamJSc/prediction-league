package models

import (
	"fmt"
	"time"
)

const (
	SeasonStatusPending = "pending"
	SeasonStatusActive  = "active"
	SeasonStatusElapsed = "elapsed"
)

// Season defines the structure of a Season against which Entries are played
type Season struct {
	ID                 string
	ClientID           ResourceIdentifier
	Name               string
	Active             TimeFrame
	EntriesAccepted    TimeFrame
	SelectionsAccepted []TimeFrame
	TeamIDs            []string
}

// GetState determines a Season's state based on a supplied timestamp
func (s Season) GetState(ts time.Time) SeasonState {
	var state SeasonState

	// determine season's current status
	state.Status = SeasonStatusPending
	switch {
	case s.Active.HasBegunBy(ts) && !s.Active.HasElapsedBy(ts):
		state.Status = SeasonStatusActive
	case s.Active.HasElapsedBy(ts):
		state.Status = SeasonStatusElapsed
	}

	// is season currently accepting entries?
	if s.EntriesAccepted.HasBegunBy(ts) && !s.EntriesAccepted.HasElapsedBy(ts) {
		state.IsAcceptingEntries = true
	}

	// is season currently accepting selections?
	for _, tf := range s.SelectionsAccepted {
		thisTf := tf

		if tf.HasBegunBy(ts) && !tf.HasElapsedBy(ts) {
			// next selections window should be the current timeframe if selections are currently being accepted
			state.IsAcceptingSelections = true
			state.NextSelectionsWindow = &thisTf
			break
		}

		// if we aren't currently accepting selections, does this tf represent the next time that we are?
		if !state.IsAcceptingSelections && !tf.HasBegunBy(ts) {
			nextTf := tf
			state.NextSelectionsWindow = &nextTf
			break
		}
	}

	return state
}

// SeasonState defines the state of a Season
type SeasonState struct {
	Status                string
	IsAcceptingEntries    bool
	IsAcceptingSelections bool
	NextSelectionsWindow  *TimeFrame
}

// SeasonCollection is map of Seasons
type SeasonCollection map[string]Season

// GetByID retrieves a matching Season from the collection by its ID
func (s SeasonCollection) GetByID(seasonID string) (Season, error) {
	for id, season := range s {
		if id == seasonID {
			return season, nil
		}
	}

	return Season{}, fmt.Errorf("season id %s: not found", seasonID)
}
