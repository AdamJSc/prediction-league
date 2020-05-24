package domain

import "prediction-league/service/internal/models"

// CalculateRankingsScores compares baseRC with comparisonRC to produce a new RankingWithScoreCollection
func CalculateRankingsScores(baseRC models.RankingCollection, comparisonRC models.RankingCollection) (models.RankingWithScoreCollection, error) {
	var collection models.RankingWithScoreCollection

	for _, r := range baseRC {
		var rws models.RankingWithScore
		rws.ID = r.ID
		rws.Position = r.Position

		var score int
		// TODO - calculate score

		rws.Score = score
		collection = append(collection, rws)
	}

	return collection, nil
}
