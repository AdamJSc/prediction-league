package models

import (
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"time"
)

const (
	EntryStatusPending = "pending"
	EntryStatusPaid    = "paid"
	EntryStatusReady   = "ready"

	EntryPaymentMethodPayPal = "paypal"
	EntryPaymentMethodOther  = "other"
)

// Entry defines a user's entry into the prediction league
type Entry struct {
	ID               uuid.UUID           `db:"id" v:"func:notEmpty"`
	ShortCode        string              `db:"short_code" v:"func:notEmpty"`
	SeasonID         string              `db:"season_id" v:"func:notEmpty"`
	RealmName        string              `db:"realm_name" v:"func:notEmpty"`
	EntrantName      string              `db:"entrant_name" v:"func:notEmpty"`
	EntrantNickname  string              `db:"entrant_nickname" v:"func:notEmpty"`
	EntrantEmail     string              `db:"entrant_email" v:"func:email"`
	Status           string              `db:"status" v:"func:isValidEntryStatus"`
	PaymentMethod    sqltypes.NullString `db:"payment_method" v:"func:isValidEntryPaymentMethod"`
	PaymentRef       sqltypes.NullString `db:"payment_ref"`
	EntryPredictions []EntryPrediction
	ApprovedAt       sqltypes.NullTime `db:"approved_at"`
	CreatedAt        time.Time         `db:"created_at"`
	UpdatedAt        sqltypes.NullTime `db:"updated_at"`
}

func (e *Entry) IsApproved() bool {
	return e.ApprovedAt.Valid
}

// EntryPrediction provides a data type for the prediction that is associated with an Entry
type EntryPrediction struct {
	ID        uuid.UUID         `db:"id"`
	EntryID   uuid.UUID         `db:"entry_id"`
	Rankings  RankingCollection `db:"rankings"`
	CreatedAt time.Time         `db:"created_at"`
}

// NewEntryPrediction returns a new EntryPrediction from the provided set of IDs
func NewEntryPrediction(ids []string) EntryPrediction {
	return EntryPrediction{
		Rankings: NewRankingCollectionFromIDs(ids),
	}
}

// ScoredEntryPrediction provides a data type for an EntryPrediction that has been scored against a Standings
type ScoredEntryPrediction struct {
	EntryPredictionID uuid.UUID          `db:"entry_prediction_id"`
	StandingsID       uuid.UUID          `db:"standings_id"`
	Rankings          []RankingWithScore `db:"rankings"`
	Score             int                `db:"score"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         sqltypes.NullTime  `db:"updated_at"`
}
