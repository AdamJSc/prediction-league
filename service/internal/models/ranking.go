package models

import (
	"encoding/json"
	"fmt"
	"github.com/LUSHDigital/uuid"
	"strings"
)

// Ranking defines our base ranking structure
type Ranking struct {
	ID       string
	Position int
}

// RankingCollection defines our collection of Rankings
type RankingCollection []Ranking

func (r *RankingCollection) GetIDs() []string {
	var ids []string
	for _, ranking := range *r {
		ids = append(ids, ranking.ID)
	}

	return ids
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

	*r = NewRankingCollection(ids)

	return nil
}

// NewRankingCollection creates a new RankingCollection from a set of IDs
func NewRankingCollection(ids []string) RankingCollection {
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

// RankingWithMeta composes a Ranking with additional meta data
type RankingWithMeta struct {
	Ranking
	MetaData map[string]int
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
	Score int
}

// Standings provides a data type for league standings that have been retrieved from an external data source
type Standings struct {
	SeasonID    string
	RoundNumber int
	Rankings    []RankingWithMeta
}

// These methods make the Standings struct sortable
func (s Standings) Len() int      { return len(s.Rankings) }
func (s Standings) Swap(i, j int) { s.Rankings[i], s.Rankings[j] = s.Rankings[j], s.Rankings[i] }
func (s Standings) Less(i, j int) bool {
	return s.Rankings[i].Position < s.Rankings[j].Position
}

// ScoredEntrySelection provides a data type for an EntrySelection that has been
type ScoredEntrySelection struct {
	EntrySelectionID uuid.UUID
	RoundNumber      int
	Rankings         []RankingWithScore
}
