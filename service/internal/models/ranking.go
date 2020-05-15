package models

import "github.com/LUSHDigital/uuid"

// Ranking define our base ranking structure
type Ranking struct {
	ID       ResourceIdentifier
	Position int
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

// EntrySelection provides a data type for the team selection that is associated with an Entry
type EntrySelection struct {
	ID        uuid.UUID
	EntryID   uuid.UUID
	Selection []Ranking
}

// ScoredEntrySelection provides a data type for an EntrySelection that has been
type ScoredEntrySelection struct {
	EntrySelectionID uuid.UUID
	RoundNumber      int
	Rankings         []RankingWithScore
}
