package models

// LeaderboardRanking represents a single ranking on the leaderboard
type LeaderboardRanking struct {
	RankingWithScore
	MinScore   int `json:"min_score"`
	TotalScore int `json:"total_score"`
}
