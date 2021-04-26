package domain

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// LeaderBoard represents the state of all cumulative entry scores for any given season and round number
type LeaderBoard struct {
	RoundNumber int                  `json:"round_number"`
	Rankings    []LeaderBoardRanking `json:"rankings"`
	LastUpdated *time.Time           `json:"last_updated"`
}

// LeaderBoardRanking represents a single ranking on the leaderboard
type LeaderBoardRanking struct {
	RankingWithScore
	MinScore   int `json:"min_score"`
	TotalScore int `json:"total_score"`
}

// LeaderBoardAgent defines the behaviours for handling LeaderBoards
type LeaderBoardAgent struct {
	er   EntryRepository
	epr  EntryPredictionRepository
	sr   StandingsRepository
	sepr ScoredEntryPredictionRepository
}

// RetrieveLeaderBoardBySeasonAndRoundNumber handles the inflation of a LeaderBoard based on the provided season ID and round number
func (l *LeaderBoardAgent) RetrieveLeaderBoardBySeasonAndRoundNumber(ctx context.Context, seasonID string, roundNumber int) (*LeaderBoard, error) {
	// ensure that provided season exists
	if _, err := SeasonsDataStore.GetByID(seasonID); err != nil {
		return nil, NotFoundError{fmt.Errorf("season id %s: not found", seasonID)}
	}

	// retrieve the standings model that pertains to the provided ids
	retrievedStandings, err := l.sr.Select(ctx, map[string]interface{}{
		"season_id":    seasonID,
		"round_number": roundNumber,
	}, false)
	if err != nil {
		switch err.(type) {
		case MissingDBRecordError:
			// we don't have a standings record for the provided round number
			// if the current round number is 1, it means the game hasn't started yet
			// so let's return an empty leaderboard - otherwise we return a standard 404
			if roundNumber != 1 {
				return nil, domainErrorFromRepositoryError(err)
			}
			entries, err := l.er.SelectBySeasonIDAndApproved(ctx, seasonID, true)
			if err != nil {
				return nil, domainErrorFromRepositoryError(err)
			}
			lb, err := l.generateEmptyLeaderBoard(ctx, roundNumber, entries)
			if err != nil {
				return nil, fmt.Errorf("cannot generate empty leaderboard: %w", err)
			}
			return lb, nil
		}
		return nil, domainErrorFromRepositoryError(err)
	}

	if len(retrievedStandings) != 1 {
		// we should never have more than 1 standings record for any given season ID and round number combination
		return nil, fmt.Errorf("retrieved %d standings, should be 1", len(retrievedStandings))
	}

	standings := retrievedStandings[0]
	realmName := RealmFromContext(ctx).Name

	lbRankings, err := l.sepr.SelectEntryCumulativeScoresByRealm(ctx, realmName, seasonID, roundNumber)
	if err != nil {
		switch err.(type) {
		case MissingDBRecordError:
			// this should never happen, because we should only have a standings record (established above) if we also
			// have some affiliated scored entry predictions
			// however, as a safety net let's check again for the first round and return an empty leaderboard if we have one
			if roundNumber != 1 {
				return nil, domainErrorFromRepositoryError(err)
			}
			entries, err := l.er.SelectBySeasonIDAndApproved(ctx, seasonID, true)
			if err != nil {
				return nil, domainErrorFromRepositoryError(err)
			}
			lb, err := l.generateEmptyLeaderBoard(ctx, roundNumber, entries)
			if err != nil {
				return nil, fmt.Errorf("cannot generate empty leaderboard: %w", err)
			}
			return lb, nil
		}
		return nil, domainErrorFromRepositoryError(err)
	}

	lastUpdated := standings.CreatedAt
	if standings.UpdatedAt.Valid {
		lastUpdated = standings.UpdatedAt.Time
	}

	return &LeaderBoard{
		RoundNumber: roundNumber,
		Rankings:    lbRankings,
		LastUpdated: &lastUpdated,
	}, nil
}

// generateEmptyLeaderBoard returns a leaderboard that comprises all of the provided entries scored with a 0
func (l *LeaderBoardAgent) generateEmptyLeaderBoard(ctx context.Context, roundNumber int, entries []Entry) (*LeaderBoard, error) {
	lb := LeaderBoard{
		RoundNumber: roundNumber,
	}

	realmName := RealmFromContext(ctx).Name

	// sort entries by entrant nickname
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].EntrantNickname < entries[j].EntrantNickname
	})

	count := 0
	for _, entry := range entries {
		if entry.RealmName != realmName {
			continue
		}
		count++
		lb.Rankings = append(lb.Rankings, LeaderBoardRanking{
			RankingWithScore: RankingWithScore{
				Ranking: Ranking{
					ID:       entry.ID.String(),
					Position: count,
				},
			},
		})
	}

	return &lb, nil
}

// NewLeaderBoardAgent returns a new LeaderBoardAgent using the provided repositories
func NewLeaderBoardAgent(er EntryRepository, epr EntryPredictionRepository, sr StandingsRepository, sepr ScoredEntryPredictionRepository) (*LeaderBoardAgent, error) {
	switch {
	case er == nil:
		return nil, fmt.Errorf("entry repository: %w", ErrIsNil)
	case epr == nil:
		return nil, fmt.Errorf("entry prediction repository: %w", ErrIsNil)
	case sr == nil:
		return nil, fmt.Errorf("standings repository: %w", ErrIsNil)
	case sepr == nil:
		return nil, fmt.Errorf("scored entry prediction repository: %w", ErrIsNil)
	}

	return &LeaderBoardAgent{
		er:   er,
		epr:  epr,
		sr:   sr,
		sepr: sepr,
	}, nil
}
