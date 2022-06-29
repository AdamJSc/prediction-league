package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// MetaKeyPlayedGames defines the ranking with meta key that represents played games
	MetaKeyPlayedGames = "playedGames" // TODO: deprecate once migrated to MatchWeekStandings over Standings
	// MetaKeyPoints defines the ranking with meta key that represents points
	MetaKeyPoints = "points" // TODO: feat - deprecate
	// MetaKeyGoalsFor defines the ranking with meta key that represents goals for
	MetaKeyGoalsFor = "goalsFor" // TODO: feat - deprecate
	// MetaKeyGoalsAgainst defines the ranking with meta key that represents goals against
	MetaKeyGoalsAgainst = "goalsAgainst" // TODO: feat - deprecate
	// MetaKeyGoalDifference defines the ranking with meta key that represents goal difference
	MetaKeyGoalDifference = "goalDifference" // TODO: feat - deprecate
)

// Ranking defines our base ranking structure
type Ranking struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

// RankingCollection defines our collection of Rankings
type RankingCollection []Ranking

// GetIDs retrieves just the IDs of all Rankings in the RankingCollection
func (r *RankingCollection) GetIDs() []string {
	var ids []string
	for _, ranking := range *r {
		ids = append(ids, ranking.ID)
	}

	return ids
}

// GetByID retrieves the Ranking from RankingCollection whose ID matches the provided ID
func (r *RankingCollection) GetByID(id string) (*Ranking, error) {
	for _, ranking := range *r {
		if ranking.ID == id {
			return &ranking, nil
		}
	}

	return nil, NotFoundError{fmt.Errorf("ranking id %s: not found", id)}
}

// MarshalJSON using custom structure
func (r *RankingCollection) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`["%s"]`, strings.Join(r.GetIDs(), `","`))), nil
}

// UnmarshalJSON to accommodate custom marshaling
func (r *RankingCollection) UnmarshalJSON(d []byte) error {
	var ids []string
	if err := json.Unmarshal(d, &ids); err != nil {
		return err
	}

	*r = NewRankingCollectionFromIDs(ids)

	return nil
}

// RankingWithMeta composes a Ranking with additional meta data
type RankingWithMeta struct {
	Ranking
	MetaData map[string]int `json:"meta_data"`
}

// NewRankingWithMeta provides an empty RankingWithMeta object with an instantiated map
func NewRankingWithMeta() RankingWithMeta {
	var r RankingWithMeta
	r.MetaData = make(map[string]int)
	return r
}

// RankingWithScore composes a Ranking with a corresponding Score value
type RankingWithScore struct {
	Ranking
	Score int `json:"score"`
}

// RankingWithScoreCollection defines our collection of RankingWithScores
type RankingWithScoreCollection []RankingWithScore

// GetTotal returns the total score of all RankingWithScore elements
func (r RankingWithScoreCollection) GetTotal() int {
	var total int

	for _, rws := range r {
		total = total + rws.Score
	}

	return total
}

// NewRankingCollectionFromIDs creates a new RankingCollection from a set of IDs
func NewRankingCollectionFromIDs(ids []string) RankingCollection {
	var (
		collection RankingCollection
		count      int
	)
	for _, id := range ids {
		count++
		collection = append(collection, Ranking{
			ID:       id,
			Position: count,
		})
	}
	return collection
}

// NewRankingCollectionFromRankingWithMetas creates a new RankingCollection from the provide RankingWithMetas
func NewRankingCollectionFromRankingWithMetas(rwms []RankingWithMeta) RankingCollection {
	var collection RankingCollection

	for _, rwm := range rwms {
		collection = append(collection, Ranking{
			ID:       rwm.ID,
			Position: rwm.Position,
		})
	}
	return collection
}

// NewRankingWithScoreCollectionFromIDs creates a new RankingWithScoreCollection from a set of IDs
func NewRankingWithScoreCollectionFromIDs(ids []string) RankingWithScoreCollection {
	var (
		collection RankingWithScoreCollection
		count      int
	)
	for _, id := range ids {
		count++

		var rws RankingWithScore
		rws.ID = id
		rws.Position = count

		collection = append(collection, rws)
	}
	return collection
}

// CalculateRankingsScores compares baseRC with comparisonRC to produce a new RankingWithScoreCollection
func CalculateRankingsScores(baseRC, comparisonRC RankingCollection) (*RankingWithScoreCollection, error) {
	var collection RankingWithScoreCollection

	if err := rankingsIDsMatch(baseRC, comparisonRC); err != nil {
		return nil, err
	}

	for _, baseRanking := range baseRC {
		var rws RankingWithScore
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

// GetChangedRankingIDs returns the Ranking IDs that differ between the two provided RankingCollection objects
func GetChangedRankingIDs(x RankingCollection, y RankingCollection) []string {
	diffMap := make(map[string]struct{}, 0)

	for _, xRnk := range x {
		yRnk, err := y.GetByID(xRnk.ID)
		if err != nil || yRnk.Position != xRnk.Position {
			diffMap[xRnk.ID] = struct{}{}
		}
	}

	for _, yRnk := range y {
		xRnk, err := x.GetByID(yRnk.ID)
		if err != nil || xRnk.Position != yRnk.Position {
			diffMap[yRnk.ID] = struct{}{}
		}
	}

	diff := make([]string, 0)
	for id := range diffMap {
		diff = append(diff, id)
	}

	return diff
}

// rankingsIDsMatch returns an error if the provided RankingCollections do not match their respective IDs in full
func rankingsIDsMatch(base, comparison RankingCollection) error {
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
