package domain

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MatchWeekSubmission represents the league table submitted on behalf of the associated entry ID for the associated match week number
type MatchWeekSubmission struct {
	ID                      uuid.UUID     // unique id
	EntryID                 uuid.UUID     // associated entry id
	MatchWeekNumber         uint16        // match week number that submission applies to (should be unique per entry)
	TeamRankings            []TeamRanking // array of team ids with their respective positions
	LegacyEntryPredictionID uuid.UUID     // original entry prediction id (retained to accommodate eventual migration)
	CreatedAt               time.Time     // date that submission was created
	UpdatedAt               *time.Time    // date that submission was most recently updated, if applicable
}

// TeamRanking associates a team ID with their position
type TeamRanking struct {
	Position uint16 // team's position, as predicted within match week submission
	TeamID   string // team's id
}

// newMatchWeekSubmissionFromEntryPredictionAndStandings converts legacy entities to newer domain entity
func newMatchWeekSubmissionFromEntryPredictionAndStandings(ep EntryPrediction, s Standings) *MatchWeekSubmission {
	return &MatchWeekSubmission{
		// id omitted - this model is eventually upserted by legacy entry prediction id / match week number so id will be populated then
		EntryID:                 ep.EntryID,
		MatchWeekNumber:         uint16(s.RoundNumber),
		TeamRankings:            newTeamRankingsFromRankingCollection(ep.Rankings),
		LegacyEntryPredictionID: ep.ID,
		CreatedAt:               ep.CreatedAt,
		UpdatedAt:               nil,
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

// MatchWeekSubmissionRepository defines i/o operations on a MatchWeekSubmission
type MatchWeekSubmissionRepository interface {
	GetByEntryIDAndMatchWeekNumber(ctx context.Context, entryID uuid.UUID, mwNumber uint16) (*MatchWeekSubmission, error)
	Insert(ctx context.Context, submission *MatchWeekSubmission) error
	Update(ctx context.Context, submission *MatchWeekSubmission) error
}

// MatchWeekSubmissionAgent encapsulates business logic relating to the MatchWeekSubmission entity
type MatchWeekSubmissionAgent struct {
	repo MatchWeekSubmissionRepository
}

// UpsertByLegacy updates the provided MatchWeekSubmission if it exists, otherwise creates it as a new one
func (m *MatchWeekSubmissionAgent) UpsertByLegacy(ctx context.Context, submission *MatchWeekSubmission) error {
	entryID := submission.EntryID
	mwNumber := submission.MatchWeekNumber

	existing, err := m.repo.GetByEntryIDAndMatchWeekNumber(ctx, entryID, mwNumber)
	switch {
	case err == nil:
		// update existing submission
		submission.ID = existing.ID
		if err := m.repo.Update(ctx, submission); err != nil {
			return fmt.Errorf("cannot update submission: %w", err)
		}
		return nil
	case errors.As(err, &MissingDBRecordError{}):
		// create new submission
		if err := m.repo.Insert(ctx, submission); err != nil {
			return fmt.Errorf("cannot insert submission: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("cannot get submission by legacy id: %w", err)
	}
}

// NewMatchWeekSubmissionAgent creates a new agent instance from the provided repository
func NewMatchWeekSubmissionAgent(repo MatchWeekSubmissionRepository) (*MatchWeekSubmissionAgent, error) {
	if repo == nil {
		return nil, fmt.Errorf("match week submission repository: %w", ErrIsNil)
	}

	return &MatchWeekSubmissionAgent{repo: repo}, nil
}

// duplicatesError defines an error that represents duplicate occurrences of a set of values
type duplicatesError struct {
	valueCountMap map[string]uint16
}

func (d duplicatesError) Error() string {
	if len(d.valueCountMap) == 0 {
		return ""
	}

	var msgs []string
	for value, count := range d.valueCountMap {
		msgs = append(msgs, fmt.Sprintf("'%s' (x%d)", value, count))
	}

	sort.SliceStable(msgs, func(i, j int) bool {
		return msgs[i] < msgs[j]
	})

	return strings.Join(msgs, ", ")
}

func newDuplicatesError() duplicatesError {
	return duplicatesError{valueCountMap: make(map[string]uint16)}
}

func checkForDuplicateTeamRankings(input []TeamRanking) error {
	// map team id and position against number of occurrences
	idCounts := make(map[string]uint16)
	positionCounts := make(map[uint16]uint16)
	for _, ranking := range input {
		id := ranking.TeamID
		pos := ranking.Position

		if _, ok := idCounts[id]; !ok {
			idCounts[id] = 0
		}
		idCounts[id]++

		if _, ok := positionCounts[pos]; !ok {
			positionCounts[pos] = 0
		}
		positionCounts[pos]++
	}

	// determine if any duplicate team ids, and return error if so
	dupeTeamIDsErr := newDuplicatesError()
	for id, count := range idCounts {
		if count > 1 {
			dupeTeamIDsErr.valueCountMap[id] = count
		}
	}
	if len(dupeTeamIDsErr.valueCountMap) > 0 {
		return fmt.Errorf("duplicate team ids found: %w", dupeTeamIDsErr)
	}

	// determine if any duplicate positions, and return error if so
	dupePositionsErr := newDuplicatesError()
	for pos, count := range positionCounts {
		if count > 1 {
			dupePositionsErr.valueCountMap[strconv.Itoa(int(pos))] = count
		}
	}
	if len(dupePositionsErr.valueCountMap) > 0 {
		return fmt.Errorf("duplicate positions found: %w", dupePositionsErr)
	}

	return nil
}
