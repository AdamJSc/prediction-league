package models

import "time"

// LeaderBoardRanking represents a single ranking on the leaderboard
type LeaderBoardRanking struct {
	RankingWithScore
	MinScore   int `json:"min_score"`
	TotalScore int `json:"total_score"`
}

// LeaderBoard represents the state of all cumulative entry scores for any given season and round number
type LeaderBoard struct {
	RoundNumber int                  `json:"round_number"`
	Rankings    []LeaderBoardRanking `json:"rankings"`
	LastUpdated *time.Time           `json:"last_updated"`
}
