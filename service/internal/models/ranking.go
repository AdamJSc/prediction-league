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

// GetIDs retrieves just the IDs of all Rankings in the RankingCollection
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

	*r = NewRankingCollectionFromIDs(ids)

	return nil
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

// RankingWithScoreCollection defines our collection of RankingWithScores
type RankingWithScoreCollection []RankingWithScore

// ScoredEntrySelection provides a data type for an EntrySelection that has been
type ScoredEntrySelection struct {
	EntrySelectionID uuid.UUID
	RoundNumber      int
	Rankings         []RankingWithScore
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
