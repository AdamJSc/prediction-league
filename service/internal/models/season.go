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

// FootballDataOrgSeasonIdentifier defines a season identifier for use with the football-data.org API
type FootballDataOrgSeasonIdentifier struct {
	ClientResourceIdentifier
	SeasonID string
}

func (f FootballDataOrgSeasonIdentifier) Value() string {
	return f.SeasonID
}

// Season defines the structure of a Season against which Entries are played
type Season struct {
	ID          string
	ClientID    FootballDataOrgSeasonIdentifier
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
func (s SeasonCollection) GetByID(seasonID string) (Season, error) {
	for id, season := range s {
		if id == seasonID {
			return season, nil
		}
	}

	return Season{}, errors.New("not found")
}
