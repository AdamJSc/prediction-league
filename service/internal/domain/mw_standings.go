package domain

import (
	"time"

	"github.com/google/uuid"
)

// MatchWeekStandings represents the league table associated with the provided season ID and match week number
type MatchWeekStandings struct {
	ID              uuid.UUID     // unique id
	SeasonID        string        // associated season id
	MatchWeekNumber uint16        // match week number that standings applies to (should be unique per season)
	TeamRankings    []TeamRanking // array of team ids with their respective positions
	FinalisedAt     *time.Time    // date that standings were finalised, if applicable
	CreatedAt       time.Time     // date that standings were created
	UpdatedAt       *time.Time    // date that standings were most recently updated, if applicable
}
