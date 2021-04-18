package domain

import (
	"context"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	repofac2 "prediction-league/service/internal/repositories/repofac"
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

// LeaderBoardAgentInjector defines the dependencies required by our LeaderBoardAgent
type LeaderBoardAgentInjector interface {
	MySQL() coresql.Agent
}

// LeaderBoardAgent defines the behaviours for handling LeaderBoards
type LeaderBoardAgent struct {
	LeaderBoardAgentInjector
}

// RetrieveLeaderBoardBySeasonAndRoundNumber handles the inflation of a LeaderBoard based on the provided season ID and round number
func (l LeaderBoardAgent) RetrieveLeaderBoardBySeasonAndRoundNumber(ctx context.Context, seasonID string, roundNumber int) (*LeaderBoard, error) {
	// ensure that provided season exists
	if _, err := SeasonsDataStore.GetByID(seasonID); err != nil {
		return nil, NotFoundError{fmt.Errorf("season id %s: not found", seasonID)}
	}

	// retrieve the standings model that pertains to the provided ids
	standingsRepo := repofac2.NewStandingsDatabaseRepository(l.MySQL())
	retrievedStandings, err := standingsRepo.Select(ctx, map[string]interface{}{
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
			lb, err := generateEmptyLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, roundNumber, &EntryAgent{EntryAgentInjector: l})
			if err != nil {
				return nil, err
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

	scoredEntryPredictionRepo := repofac2.NewScoredEntryPredictionDatabaseRepository(l.MySQL())
	lbRankings, err := scoredEntryPredictionRepo.SelectEntryCumulativeScoresByRealm(ctx, realmName, seasonID, roundNumber)
	if err != nil {
		switch err.(type) {
		case MissingDBRecordError:
			// this should never happen, because we should only have a standings record (established above) if we also
			// have some affiliated scored entry predictions
			// however, as a safety net let's check again for the first round and return an empty leaderboard if we have one
			if roundNumber != 1 {
				return nil, domainErrorFromRepositoryError(err)
			}
			lb, err := generateEmptyLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, roundNumber, &EntryAgent{EntryAgentInjector: l})
			if err != nil {
				return nil, err
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

// generateEmptyLeaderBoardBySeasonAndRoundNumber returns a leaderboard that comprises all entries belonging to
// the provided season ID and realm name of the provided context, which are all scored with a 0
func generateEmptyLeaderBoardBySeasonAndRoundNumber(ctx context.Context, seasonID string, roundNumber int, entryAgent *EntryAgent) (*LeaderBoard, error) {
	entries, err := entryAgent.RetrieveEntriesBySeasonID(ctx, seasonID, true)
	if err != nil {
		return nil, err
	}

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
