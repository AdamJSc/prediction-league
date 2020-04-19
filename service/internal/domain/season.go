package domain

import (
	"errors"
	"github.com/ladydascalie/v"
	"log"
	"time"
)

const (
	SeasonStatusForthcoming      = "forthcoming"
	SeasonStatusAcceptingEntries = "accepting_entries"
	SeasonStatusActive           = "active"
	SeasonStatusElapsed          = "elapsed"
)

type SeasonCollection map[string]Season

func (c SeasonCollection) GetByID(seasonID string) (Season, error) {
	for id, season := range c {
		if id == seasonID {
			return season, nil
		}
	}

	return Season{}, errors.New("not found")
}

func Seasons() SeasonCollection {
	ukLoc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatal(err)
	}

	return map[string]Season{
		"201920_1": {
			ID:          "201920_1",
			Name:        "Premier League 2019/20",
			EntriesFrom: time.Date(2019, 7, 1, 0, 0, 0, 0, ukLoc),
			StartDate:   time.Date(2019, 8, 9, 19, 0, 0, 0, ukLoc),
			EndDate:     time.Date(2020, 5, 17, 23, 59, 59, 0, ukLoc),
		},
	}
}

type Season struct {
	ID          string    `v:"func:notEmpty"`
	Name        string    `v:"func:notEmpty"`
	EntriesFrom time.Time `v:"func:notEmpty"`
	StartDate   time.Time `v:"func:notEmpty"`
	EndDate     time.Time `v:"func:notEmpty"`
}

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

func ValidateSeason(s Season) error {
	if err := v.Struct(s); err != nil {
		return vPackageErrorToValidationError(err, s)
	}

	if !s.EntriesFrom.Before(s.StartDate) {
		return ValidationError{
			Reasons: []string{"EntriesFrom date cannot occur before Start date"},
			Fields:  []string{"entries_from", "start_date"},
		}
	}

	if !s.StartDate.Before(s.EndDate) {
		return ValidationError{
			Reasons: []string{"End date cannot occur before Start date"},
			Fields:  []string{"start_date", "end_date"},
		}
	}

	return nil
}
