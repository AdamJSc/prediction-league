package domain

import (
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"github.com/ladydascalie/v"
	"strings"
	"time"
)

const (
	entryStatusPending  = "pending"
	entryStatusPaid     = "paid"
	entryStatusComplete = "complete"
)

type Entry struct {
	ID              uuid.UUID           `db:"id" v:"func:notEmpty"`
	LookupRef       string              `db:"lookup_ref" v:"func:notEmpty"`
	SeasonID        string              `db:"season_id" v:"func:notEmpty"`
	Realm           string              `db:"realm" v:"func:notEmpty"`
	EntrantName     string              `db:"entrant_name" v:"func:notEmpty"`
	EntrantNickname string              `db:"entrant_nickname" v:"func:notEmpty"`
	EntrantEmail    string              `db:"entrant_email" v:"func:email"`
	Status          string              `db:"status" v:"func:isEntryStatus"`
	PaymentRef      sqltypes.NullString `db:"payment_ref"`
	TeamIDSequence  []string            `db:"team_id_sequence"`
	CreatedAt       time.Time           `db:"created_at"`
	UpdatedAt       sqltypes.NullTime   `db:"updated_at"`
}

type EntryAgentInjector interface {
	MySQL() coresql.Agent
}

type EntryAgent struct{ EntryAgentInjector }

func (a EntryAgent) CreateEntry(ctx Context, e Entry, s *Season, realmPIN string) (Entry, error) {
	if s == nil {
		return Entry{}, InternalError{errors.New("invalid season")}
	}

	if err := validateRealmPIN(ctx, realmPIN); err != nil {
		return Entry{}, UnauthorizedError{errors.New("invalid PIN")}
	}

	if s.GetStatus(time.Now()) != seasonStatusAcceptingEntries {
		return Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return Entry{}, InternalError{err}
	}

	e.ID = uuid
	e.SeasonID = s.ID
	e.Realm = ctx.GetRealm()
	e.Status = entryStatusPending

	e.LookupRef, err = generateUniqueLookupRef(a.MySQL())
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, vPackageErrorToValidationError(err, e)
	}

	// check for existing nickname
	existingNicknameEntries, err := dbSelectEntries(a.MySQL(), map[string]interface{}{
		"season_id":        e.SeasonID,
		"realm":            e.Realm,
		"entrant_nickname": e.EntrantNickname,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	// check for existing email
	existingEmailEntries, err := dbSelectEntries(a.MySQL(), map[string]interface{}{
		"season_id":     e.SeasonID,
		"realm":         e.Realm,
		"entrant_email": e.EntrantEmail,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	if len(existingNicknameEntries) > 0 || len(existingEmailEntries) > 0 {
		return Entry{}, ConflictError{errors.New("entry already exists")}
	}

	if err := dbInsertEntry(a.MySQL(), &e); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

func sanitiseEntry(e *Entry) error {
	if err := v.Struct(e); err != nil {
		return vPackageErrorToValidationError(err, *e)
	}

	e.EntrantName = strings.Trim(e.EntrantName, " ")
	e.EntrantNickname = strings.Replace(e.EntrantNickname, " ", "", -1)
	e.EntrantEmail = strings.Trim(e.EntrantEmail, " ")

	return nil
}

func generateUniqueLookupRef(db coresql.Agent) (string, error) {
	lookupRef := generateRandomAlphaNumericString(4)

	existingLookupRefEntries, err := dbSelectEntries(db, map[string]interface{}{
		"lookup_ref": lookupRef,
	}, false)
	if err != nil {
		return "", wrapDBError(err)
	}

	if len(existingLookupRefEntries) > 0 {
		return generateUniqueLookupRef(db)
	}

	return lookupRef, nil
}
