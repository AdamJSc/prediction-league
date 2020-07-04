package models

// LeaderBoardRanking represents a single ranking on the leaderboard
type LeaderBoardRanking struct {
	RankingWithScore
	MinScore   int `json:"min_score"`
	TotalScore int `json:"total_score"`
}
