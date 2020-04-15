package domain

import (
	"context"
	"fmt"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/ladydascalie/v"
	"prediction-league/service/internal/app"
	"strconv"
	"time"
)

type Season struct {
	ID           string            `json:"id" db:"id"`
	YearRef      int               `json:"year_ref" db:"year_ref"`
	Variant      int               `json:"variant" db:"variant"`
	Name         string            `json:"name" db:"name" v:"func:notempty"`
	EntriesFrom  time.Time         `json:"entries_from" db:"entries_from" v:"func:notempty"`
	EntriesUntil sqltypes.NullTime `json:"entries_until" db:"entries_until"`
	StartDate    time.Time         `json:"start_date" db:"start_date" v:"func:notempty"`
	EndDate      time.Time         `json:"end_date" db:"end_date" v:"func:notempty"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    sqltypes.NullTime `json:"updated_at" db:"updated_at"`
}

func (s Season) GenerateID() string {
	return fmt.Sprintf("%d_%d", s.YearRef, s.Variant)
}

type SeasonAgentInjector interface {
	app.MySQLInjector
}

type SeasonAgent struct {
	SeasonAgentInjector
}

func (a SeasonAgent) InsertSeason(ctx context.Context, s *Season) error {
	if err := sanitiseSeason(s); err != nil {
		return err
	}

	if err := insertSeason(a.MySQL(), s); err != nil {
		return domainErrorFromDBError(err)
	}

	return nil
}

func sanitiseSeason(s *Season) error {
	if err := v.Struct(s); err != nil {
		return vPackageErrorToValidationError(err, *s)
	}

	startRef := s.StartDate.Format("2006")
	endRef := s.EndDate.Format("06")

	yearRef, err := strconv.Atoi(fmt.Sprintf("%s%s", startRef, endRef))
	if err != nil {
		return InternalError{err}
	}

	s.YearRef = yearRef

	if s.Variant == 0 {
		s.Variant = 1
	}

	s.ID = s.GenerateID()

	if s.EndDate.Before(s.StartDate) {
		return ValidationError{
			Reasons: []string{"End date cannot occur before start date"},
			Fields: []string{"start_date", "end_date"},
		}
	}

	if !s.EntriesUntil.Valid {
		s.EntriesUntil = sqltypes.ToNullTime(s.StartDate)
	}

	if s.EntriesUntil.Time.Before(s.EntriesFrom) {
		return ValidationError{
			Reasons: []string{"Entry period must end before it begins"},
			Fields: []string{"entries_until", "start_date"},
		}
	}

	if s.EntriesUntil.Time.After(s.StartDate) {
		return ValidationError{
			Reasons: []string{"Entry period must end before start date commences"},
			Fields: []string{"entries_until", "start_date"},
		}
	}

	return nil
}
