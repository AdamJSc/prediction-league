package models

import (
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"time"
)

// Standings provides a data type for league standings that have been retrieved from an external data source
type Standings struct {
	ID          uuid.UUID         `db:"id"`
	SeasonID    string            `db:"season_id"`
	RoundNumber int               `db:"round_number"`
	Rankings    []RankingWithMeta `db:"rankings"`
	Finalised   bool              `db:"finalised"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   sqltypes.NullTime `db:"updated_at"`
}

// These methods make the Standings struct sortable
func (s Standings) Len() int      { return len(s.Rankings) }
func (s Standings) Swap(i, j int) { s.Rankings[i], s.Rankings[j] = s.Rankings[j], s.Rankings[i] }
func (s Standings) Less(i, j int) bool {
	return s.Rankings[i].Position < s.Rankings[j].Position
}
