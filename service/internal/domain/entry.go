package domain

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"github.com/ladydascalie/v"
	"strings"
	"time"
)

const (
	entryStatusPending = "pending"
	entryStatusPaid    = "paid"
	entryStatusReady   = "ready"
)

type Entry struct {
	ID              uuid.UUID           `json:"id" db:"id" v:"func:notEmpty"`
	SeasonID        string              `json:"season_id" db:"season_id" v:"func:notEmpty"`
	Realm           string              `json:"realm" db:"realm" v:"func:notEmpty"`
	EntrantName     string              `json:"entrant_name" db:"entrant_name" v:"func:notEmpty"`
	EntrantNickname string              `json:"entrant_nickname" db:"entrant_nickname" v:"func:notEmpty"`
	EntrantEmail    string              `json:"entrant_email" db:"entrant_email" v:"func:email"`
	Status          string              `json:"status" db:"status" v:"func:isEntryStatus"`
	PaymentRef      sqltypes.NullString `json:"-" db:"payment_ref"`
	TeamIDSequence  []string            `json:"team_id_sequence" db:"team_id_sequence"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdateAt        sqltypes.NullTime   `json:"updated_at" db:"updated_at"`
}

type EntryAgentInjector interface {
	MySQL() coresql.Agent
}

type EntryAgent struct{ EntryAgentInjector }

func (a EntryAgent) CreateEntryForSeason(ctx Context, e Entry, s Season, realmPIN int) (Entry, error) {
	if err := validateRealmPIN(ctx, realmPIN); err != nil {
		return Entry{}, err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return Entry{}, err
	}

	e.ID = uuid
	e.SeasonID = s.ID
	e.Realm = ctx.getRealm()
	e.Status = entryStatusPending

	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, err
	}

	if err := dbInsertEntry(a.MySQL(), &e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

func sanitiseEntry(e *Entry) error {
	if err := v.Struct(e); err != nil {
		return vPackageErrorToValidationError(err, *e)
	}

	e.EntrantName = strings.Trim(e.EntrantName, " ")
	e.EntrantNickname = strings.Trim(e.EntrantNickname, " ")
	e.EntrantEmail = strings.Trim(e.EntrantEmail, " ")

	return nil
}
