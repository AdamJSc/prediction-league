package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// MetaKeyPlayedGames defines the ranking with meta key that represents played games
	MetaKeyPlayedGames = "playedGames" // TODO: deprecate once migrated to MatchWeekStandings from Standings
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
