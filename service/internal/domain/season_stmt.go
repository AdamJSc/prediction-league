package domain

import (
	coresql "github.com/LUSHDigital/core-sql"
	"time"
)

func insertSeason(db coresql.Agent, s *Season) error {
	stmt := `INSERT INTO season (id, year_ref, variant, name, entries_from, entries_until, start_date, end_date, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	if _, err := db.Query(
		stmt,
		s.ID,
		s.YearRef,
		s.Variant,
		s.Name,
		s.EntriesFrom,
		s.EntriesUntil,
		s.StartDate,
		s.EndDate,
		now,
	); err != nil {
		return wrapDBError(err)
	}

	s.CreatedAt = now

	return nil
}
