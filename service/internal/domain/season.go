package domain

import (
	"errors"
	"fmt"
	"time"
)

const (
	SeasonStatusForthcoming      = "forthcoming"
	SeasonStatusAcceptingEntries = "accepting_entries"
	SeasonStatusActive           = "active"
	SeasonStatusElapsed          = "elapsed"
)

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

// Seasons returns a pre-determined data structure of all Seasons that can be referenced within the system
func Seasons() SeasonCollection {
	return map[string]Season{
		"201920_1": {
			ID:          "201920_1",
			Name:        "Premier League 2019/20",
			EntriesFrom: time.Date(2019, 7, 1, 0, 0, 0, 0, UKLocation),
			StartDate:   time.Date(2019, 8, 9, 19, 0, 0, 0, UKLocation),
			EndDate:     time.Date(2020, 5, 17, 23, 59, 59, 0, UKLocation),
		},
	}
}

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

// ValidateSeason returns an error if validation rules are not satisfied for a provided Season
func ValidateSeason(s Season) error {
	if err := sanitiseSeason(&s); err != nil {
		return err
	}

	if !s.EntriesFrom.Before(s.StartDate) {
		return ValidationError{
			Reasons: []string{"EntriesFrom date cannot occur before Start date"},
		}
	}

	if !s.StartDate.Before(s.EndDate) {
		return ValidationError{
			Reasons: []string{"End date cannot occur before Start date"},
		}
	}

	return nil
}

// sanitiseSeason sanitises and validates a Season
func sanitiseSeason(s *Season) error {
	var validationMsgs []string

	// validate
	for k, v := range map[string]string{
		"ID":   s.ID,
		"Name": s.Name,
	} {
		if v == "" {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}

	emptyTime := time.Time{}.Format(time.RFC3339Nano)
	if s.StartDate.Format(time.RFC3339Nano) == emptyTime {
		validationMsgs = append(validationMsgs, "Start Date must not be empty")
	}
	if s.EndDate.Format(time.RFC3339Nano) == emptyTime {
		validationMsgs = append(validationMsgs, "End Date must not be empty")
	}
	if s.EntriesFrom.Format(time.RFC3339Nano) == emptyTime {
		validationMsgs = append(validationMsgs, "Entries From Date must not be empty")
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	return nil
}
