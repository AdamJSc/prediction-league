package domain

import (
	"fmt"
	"prediction-league/service/internal/models"
)

// CalculateRankingsScores compares baseRC with comparisonRC to produce a new RankingWithScoreCollection
func CalculateRankingsScores(baseRC, comparisonRC models.RankingCollection) (*models.RankingWithScoreCollection, error) {
	var collection models.RankingWithScoreCollection

	if err := rankingsIDsMatch(baseRC, comparisonRC); err != nil {
		return nil, err
	}

	for _, baseRanking := range baseRC {
		var rws models.RankingWithScore
		rws.ID = baseRanking.ID
		rws.Position = baseRanking.Position

		comparisonRanking, err := comparisonRC.GetByID(baseRanking.ID)
		if err != nil {
			return nil, NotFoundError{err}
		}

		// score should be the absolute value of the difference between our ranking positions
		diff := baseRanking.Position - comparisonRanking.Position
		switch {
		case diff < 0:
			rws.Score = -diff
		default:
			rws.Score = diff
		}

		collection = append(collection, rws)
	}

	return &collection, nil
}

// rankingsIDsMatch returns an error if the provided RankingCollections do not match their respective IDs in full
func rankingsIDsMatch(base, comparison models.RankingCollection) error {
	baseIDs := base.GetIDs()
	compIDs := comparison.GetIDs()

	if len(baseIDs) != len(compIDs) {
		return fmt.Errorf("mismatched baseIDs length: base %d, comparison %d", len(baseIDs), len(compIDs))
	}

	mapBaseIDs := make(map[string]int)
	mapCompIDs := make(map[string]int)

	for _, id := range baseIDs {
		count := mapBaseIDs[id]
		mapBaseIDs[id] = count + 1
	}

	for _, id := range compIDs {
		count := mapCompIDs[id]
		mapCompIDs[id] = count + 1
	}

	for id, compCount := range mapCompIDs {
		baseCount, ok := mapBaseIDs[id]
		if !ok {
			return fmt.Errorf("base collection does not have id: '%s'", id)
		}

		if baseCount != compCount {
			return fmt.Errorf("mismatched counts: id '%s' base collection count = %d, collection count = %d", id, baseCount, compCount)
		}
	}

	return nil
}
