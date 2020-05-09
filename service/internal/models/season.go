package models

import (
	"errors"
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
	ID          string
	Name        string
	EntriesFrom time.Time
	StartDate   time.Time
	EndDate     time.Time
}

// GetStatus determines a Season's status based on a supplied timestamp/object
func (s Season) GetStatus(ts time.Time) string {
	switch {
	case ts.Before(s.EntriesFrom):
		return SeasonStatusForthcoming
	case ts.Before(s.StartDate):
		return SeasonStatusAcceptingEntries
	case ts.Before(s.EndDate):
		return SeasonStatusActive
	}
	return SeasonStatusElapsed
}

// SeasonCollection is map of Seasons
type SeasonCollection map[string]Season

// GetByID retrieves a matching Season from the collection by its ID
func (c SeasonCollection) GetByID(seasonID string) (Season, error) {
	for id, season := range c {
		if id == seasonID {
			return season, nil
		}
	}

	return Season{}, errors.New("not found")
}
