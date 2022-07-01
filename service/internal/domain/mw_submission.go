package domain

import (
	"fmt"
	"sort"
	"strings"
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

// newMatchWeekSubmissionFromEntryPredictionAndStandings converts legacy entities to newer domain entity
func newMatchWeekSubmissionFromEntryPredictionAndStandings(ep EntryPrediction, s Standings) *MatchWeekSubmission {
	return &MatchWeekSubmission{
		ID:              ep.ID,
		EntryID:         ep.EntryID,
		MatchWeekNumber: uint16(s.RoundNumber),
		TeamRankings:    newTeamRankingsFromRankingCollection(ep.Rankings),
		CreatedAt:       ep.CreatedAt,
		UpdatedAt:       nil,
	}
}

// newTeamRankingsFromRankingCollection converts legacy entity to newer domain entity
func newTeamRankingsFromRankingCollection(rc RankingCollection) []TeamRanking {
	rankings := make([]TeamRanking, 0)

	for _, r := range rc {
		rankings = append(rankings, TeamRanking{
			Position: uint16(r.Position),
			TeamID:   r.ID,
		})
	}

	return rankings
}

// newStandingsTeamRankingsFromRankingsWithMeta converts legacy entity to newer domain entity
func newStandingsTeamRankingsFromRankingsWithMeta(rwms []RankingWithMeta) []StandingsTeamRanking {
	rankings := make([]StandingsTeamRanking, 0)

	for _, rwm := range rwms {
		gamesPlayed := rwm.MetaData[MetaKeyPlayedGames]

		rankings = append(rankings, StandingsTeamRanking{
			TeamRanking: TeamRanking{
				Position: uint16(rwm.Position),
				TeamID:   rwm.ID,
			},
			GamesPlayed: uint16(gamesPlayed),
		})
	}

	return rankings
}

// newRankingsWithScoreFromResultTeamRankings converts newer domain entity to legacy entity to fulfil business logic proxy
func newRankingsWithScoreFromResultTeamRankings(resultRankings []ResultTeamRanking) []RankingWithScore {
	rankingsWithScore := make([]RankingWithScore, 0)

	for _, resultRank := range resultRankings {
		rankingsWithScore = append(rankingsWithScore, RankingWithScore{
			Ranking: Ranking{
				ID:       resultRank.TeamRanking.TeamID,
				Position: int(resultRank.TeamRanking.Position),
			},
			Score: int(resultRank.Hit),
		})
	}

	return rankingsWithScore
}

// duplicateStringsError defines an error that represents duplicate occurrences of a set of values
type duplicateStringsError struct {
	valueCountMap map[string]uint16
}

func (d duplicateStringsError) Error() string {
	if len(d.valueCountMap) == 0 {
		return ""
	}

	var msgs []string
	for value, count := range d.valueCountMap {
		msgs = append(msgs, fmt.Sprintf("'%s' (%d)", value, count))
	}

	sort.SliceStable(msgs, func(i, j int) bool {
		return msgs[i] < msgs[j]
	})

	return strings.Join(msgs, ", ")
}

func newDuplicateStringsError() duplicateStringsError {
	return duplicateStringsError{valueCountMap: make(map[string]uint16)}
}

func checkForDuplicateTeamRankings(input []TeamRanking) error {
	// map team id against number of occurrences
	idCountMap := make(map[string]uint16)
	for _, ranking := range input {
		if _, ok := idCountMap[ranking.TeamID]; !ok {
			idCountMap[ranking.TeamID] = 0
		}
		idCountMap[ranking.TeamID]++
	}

	// determine if any duplicate team ids, and return error if so
	dupeTeamIDsErr := newDuplicateStringsError()
	for id, count := range idCountMap {
		if count > 1 {
			dupeTeamIDsErr.valueCountMap[id] = count
		}
	}
	if len(dupeTeamIDsErr.valueCountMap) > 0 {
		return fmt.Errorf("duplicate team ids found: %w", dupeTeamIDsErr)
	}

	// TODO: feat - check for duplicate positions

	return nil
}
