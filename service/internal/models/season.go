package models

import (
	"fmt"
	"time"
)

const (
	SeasonStatusForthcoming      = "forthcoming"
	SeasonStatusAcceptingEntries = "accepting_entries"
	SeasonStatusActive           = "active"
	SeasonStatusElapsed          = "elapsed"
)

// Season defines the structure of a Season against which Entries are played
type Season struct {
	ID              string
	ClientID        ResourceIdentifier
	Name            string
	Active          TimeFrame
	EntriesAccepted TimeFrame
	TeamIDs         []string
}

// GetStatus determines a Season's status based on a supplied timestamp/object
func (s Season) GetStatus(ts time.Time) string {
	switch {
	case ts.Before(s.EntriesAccepted.From):
		return SeasonStatusForthcoming
	case ts.Before(s.Active.From):
		return SeasonStatusAcceptingEntries
	case ts.Before(s.Active.Until):
		return SeasonStatusActive
	}
	return SeasonStatusElapsed
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
