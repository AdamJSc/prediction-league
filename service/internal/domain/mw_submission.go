package domain

import (
	"time"

	"github.com/google/uuid"
)

// MatchWeekSubmission represents the league table submitted on behalf of the associated entry ID for the associated match week number
type MatchWeekSubmission struct {
	ID              uuid.UUID     // unique id
	EntryID         uuid.UUID     // associated entry id
	MatchWeekNumber uint16        // match week number that submission applies to (should be unique per entry)
	TeamRankings    []TeamRanking // array of team ids with their respective positions
	CreatedAt       time.Time     // date that submission was created
	UpdatedAt       *time.Time    // date that submission was most recently updated, if applicable
}

// TeamRanking associates a team ID with their position
type TeamRanking struct {
	Position uint16 // team's position, as predicted within match week submission
	TeamID   string // team's id
}
