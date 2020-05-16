package models

import (
	"github.com/LUSHDigital/uuid"
)

// Ranking defines our base ranking structure
type Ranking struct {
	ID       string
	Position int
}

// RankingCollection defines our collection of Rankings
type RankingCollection []Ranking

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
