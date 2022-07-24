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
	MaxScore   int `json:"max_score"`
	TotalScore int `json:"total_score"`
	Movement   int `json:"movement"`
}

// LeaderBoardAgent defines the behaviours for handling LeaderBoards
type LeaderBoardAgent struct {
	er   EntryRepository
	epr  EntryPredictionRepository
	sr   StandingsRepository
	sepr ScoredEntryPredictionRepository
	sc   SeasonCollection
}

// RetrieveLeaderBoardBySeasonAndRoundNumber handles the inflation of a LeaderBoard based on the provided season ID and round number
func (l *LeaderBoardAgent) RetrieveLeaderBoardBySeasonAndRoundNumber(ctx context.Context, seasonID string, roundNumber int) (*LeaderBoard, error) {
	// ensure that provided season exists
	if _, err := l.sc.GetByID(seasonID); err != nil {
		return nil, NotFoundError{fmt.Errorf("season id %s: not found", seasonID)}
	}

	// retrieve the standings model that pertains to the provided ids
	retrievedStandings, err := l.sr.Select(ctx, map[string]interface{}{
		"season_id":    seasonID,
		"round_number": roundNumber,
	}, false)
	if err != nil {
		return l.emptyLeaderBoardOrError(ctx, err, seasonID, roundNumber)
	}

	if len(retrievedStandings) != 1 {
		// we should never have more than 1 standings record for any given season ID and round number combination
		return nil, fmt.Errorf("retrieved %d standings, should be 1", len(retrievedStandings))
	}

	standings := retrievedStandings[0]

	realm := RealmFromContext(ctx)
	realmName := realm.Config.Name

	rankingsThisRound, err := l.sepr.SelectEntryCumulativeScoresByRealm(ctx, realmName, seasonID, roundNumber)
	if err != nil {
		return l.emptyLeaderBoardOrError(ctx, err, seasonID, roundNumber)
	}

	if roundNumber > 1 {
		if rankingsPreviousRound, err := l.sepr.SelectEntryCumulativeScoresByRealm(ctx, realmName, seasonID, roundNumber-1); err == nil {
			rankingsThisRound = populateRankingsWithMovement(rankingsThisRound, rankingsPreviousRound)
		}
	}

	lastUpdated := standings.CreatedAt
	if standings.UpdatedAt != nil {
		lastUpdated = *standings.UpdatedAt
	}

	return &LeaderBoard{
		RoundNumber: roundNumber,
		Rankings:    rankingsThisRound,
		LastUpdated: &lastUpdated,
	}, nil
}

// emptyLeaderBoardOrError returns an empty leaderboard if the provided error represents a missing database entry
func (l *LeaderBoardAgent) emptyLeaderBoardOrError(ctx context.Context, err error, seasonID string, roundNumber int) (*LeaderBoard, error) {
	switch err.(type) {

	case MissingDBRecordError:
		// if the current round number is 1, it means the game hasn't started yet
		// so let's return an empty leaderboard - otherwise we return a standard 404
		if roundNumber != 1 {
			return nil, domainErrorFromRepositoryError(err)
		}

		entries, selectErr := l.er.SelectBySeasonIDAndApproved(ctx, seasonID, true)
		if selectErr != nil {
			return nil, domainErrorFromRepositoryError(selectErr)
		}

		lb, lbErr := l.generateEmptyLeaderBoard(ctx, roundNumber, entries)
		if lbErr != nil {
			return nil, fmt.Errorf("cannot generate empty leaderboard: %w", lbErr)
		}

		return lb, nil

	default:
		return nil, domainErrorFromRepositoryError(err)
	}
}

// populateRankingsWithMovement returns the current rankings enriched with movement tallies that are relative to the previous rankings
func populateRankingsWithMovement(currentRankings, previousRankings []LeaderBoardRanking) []LeaderBoardRanking {
	currentRankingsWithMovement := make([]LeaderBoardRanking, 0)

	for _, current := range currentRankings {
		for _, previous := range previousRankings {
			if previous.ID == current.ID {
				current.Movement = previous.Position - current.Position
			}
		}

		currentRankingsWithMovement = append(currentRankingsWithMovement, current)
	}

	return currentRankingsWithMovement
}

// generateEmptyLeaderBoard returns a leaderboard that comprises all the provided entries scored with a 0
func (l *LeaderBoardAgent) generateEmptyLeaderBoard(ctx context.Context, roundNumber int, entries []Entry) (*LeaderBoard, error) {
	lb := LeaderBoard{
		RoundNumber: roundNumber,
	}

	realm := RealmFromContext(ctx)
	realmName := realm.Config.Name

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
func NewLeaderBoardAgent(er EntryRepository, epr EntryPredictionRepository, sr StandingsRepository, sepr ScoredEntryPredictionRepository, sc SeasonCollection) (*LeaderBoardAgent, error) {
	switch {
	case er == nil:
		return nil, fmt.Errorf("entry repository: %w", ErrIsNil)
	case epr == nil:
		return nil, fmt.Errorf("entry prediction repository: %w", ErrIsNil)
	case sr == nil:
		return nil, fmt.Errorf("standings repository: %w", ErrIsNil)
	case sepr == nil:
		return nil, fmt.Errorf("scored entry prediction repository: %w", ErrIsNil)
	case sc == nil:
		return nil, fmt.Errorf("season collection: %w", ErrIsNil)
	}

	return &LeaderBoardAgent{
		er:   er,
		epr:  epr,
		sr:   sr,
		sepr: sepr,
		sc:   sc,
	}, nil
}
