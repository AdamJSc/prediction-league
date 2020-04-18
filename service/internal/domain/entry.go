package domain

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"time"
)

type Entry struct {
	ID              string              `json:"id" db:"id"`
	SeasonID        string              `json:"season_id" db:"season_id"`
	Realm           string              `json:"realm" db:"realm"`
	EntrantName     string              `json:"entrant_name" db:"entrant_name"`
	EntrantNickname string              `json:"entrant_nickname" db:"entrant_nickname"`
	EntrantEmail    string              `json:"entrant_email" db:"entrant_email"`
	Status          string              `json:"status" db:"status"`
	PaymentRef      sqltypes.NullString `json:"-" db:"payment_ref"`
	TeamIDSequence  []string            `json:"team_id_sequence" db:"team_id_sequence"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdateAt        sqltypes.NullTime   `json:"updated_at" db:"updated_at"`
}

type EntryAgentInjector interface {
	MySQL() coresql.Agent
}

type EntryAgent struct{ EntryAgentInjector }

func (a EntryAgent) CreateEntry(ctx Context, e Entry, realmPIN int) (Entry, error) {
	if err := validateRealmPIN(ctx, realmPIN); err != nil {
		return Entry{}, err
	}

	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, err
	}

	if err := dbInsertEntry(a.MySQL(), &e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

func sanitiseEntry(e *Entry) error {
	// TODO - VALIDATION
	return nil
}
