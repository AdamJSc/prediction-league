package domain

import (
	"fmt"
	"github.com/LUSHDigital/core-sql/sqltypes"
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
