package domain

import (
	"errors"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/ladydascalie/v"
	"log"
	"time"
)

const (
	seasonStatusForthcoming      = "forthcoming"
	seasonStatusAcceptingEntries = "accepting_entries"
	seasonStatusActive           = "active"
	seasonStatusElapsed          = "elapsed"
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

func (c SeasonCollection) GetByTimestamp(ts time.Time) (Season, error) {
	for _, season := range c {
		if season.EntriesFrom.Before(ts) && season.EndDate.After(ts) {
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
			StartDate:   time.Date(2020, 8, 9, 19, 0, 0, 0, ukLoc),
			EndDate:     time.Date(2020, 5, 17, 23, 59, 59, 0, ukLoc),
		},
	}
}

type Season struct {
	// TODO - remove unneeded fields
	ID           string            `json:"id" db:"id" v:"func:notEmpty"`
	Name         string            `json:"name" db:"name" v:"func:notEmpty"`
	EntriesFrom  time.Time         `json:"entries_from" db:"entries_from" v:"func:notEmpty"`
	EntriesUntil sqltypes.NullTime `json:"entries_until" db:"entries_until"`
	StartDate    time.Time         `json:"start_date" db:"start_date" v:"func:notEmpty"`
	EndDate      time.Time         `json:"end_date" db:"end_date" v:"func:notEmpty"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    sqltypes.NullTime `json:"updated_at" db:"updated_at"`
}

func (s Season) GetStatus(ts time.Time) string {
	switch {
	case ts.Before(s.EntriesFrom):
		return seasonStatusForthcoming
	case ts.Before(s.StartDate):
		return seasonStatusAcceptingEntries
	case ts.Before(s.EndDate):
		return seasonStatusActive
	}
	return seasonStatusElapsed
}

func SanitiseSeason(s *Season) error {
	if err := v.Struct(s); err != nil {
		return vPackageErrorToValidationError(err, *s)
	}

	if s.EndDate.Before(s.StartDate) {
		return ValidationError{
			Reasons: []string{"End date cannot occur before start date"},
			Fields:  []string{"start_date", "end_date"},
		}
	}

	if !s.EntriesUntil.Valid {
		s.EntriesUntil = sqltypes.ToNullTime(s.StartDate)
	}

	if s.EntriesUntil.Time.Before(s.EntriesFrom) {
		return ValidationError{
			Reasons: []string{"Entry period must end before it begins"},
			Fields:  []string{"entries_until", "start_date"},
		}
	}

	if s.EntriesUntil.Time.After(s.StartDate) {
		return ValidationError{
			Reasons: []string{"Entry period must end before start date commences"},
			Fields:  []string{"entries_until", "start_date"},
		}
	}

	return nil
}
