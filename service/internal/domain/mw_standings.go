package domain

import (
	"time"

	"github.com/google/uuid"
)

// MatchWeekStandings represents the league table associated with the provided season ID and match week number
type MatchWeekStandings struct {
	ID              uuid.UUID              // unique id
	SeasonID        string                 // associated season id
	MatchWeekNumber uint16                 // match week number that standings applies to (should be unique per season)
	TeamRankings    []StandingsTeamRanking // array of team ids with their respective positions and number of games played
	FinalisedAt     *time.Time             // date that standings were finalised, if applicable
	CreatedAt       time.Time              // date that standings were created
	UpdatedAt       *time.Time             // date that standings were most recently updated, if applicable
}

// StandingsTeamRanking associates a team ranking with number of games played
type StandingsTeamRanking struct {
	TeamRanking        // team id + position
	GamesPlayed uint16 // number of games the associated team has played
}

// getTeamRankingsfromStandingsTeamRankings returns only the embedded team rankings from the provided slice of standings team rankings
func getTeamRankingsfromStandingsTeamRankings(standingsRankings []StandingsTeamRanking) []TeamRanking {
	teamRankings := make([]TeamRanking, 0)

	for _, standRank := range standingsRankings {
		teamRankings = append(teamRankings, standRank.TeamRanking)
	}

	return teamRankings
}

// newMatchWeekStandingsFromStandings converts legacy entity to newer domain entity
func newMatchWeekStandingsFromStandings(s Standings) *MatchWeekStandings {
	var finalisedAt *time.Time
	if s.Finalised {
		finalisedAt = s.UpdatedAt
	}

	return &MatchWeekStandings{
		ID:              s.ID,
		SeasonID:        s.SeasonID,
		MatchWeekNumber: uint16(s.RoundNumber),
		TeamRankings:    newStandingsTeamRankingsFromRankingsWithMeta(s.Rankings),
		FinalisedAt:     finalisedAt,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
