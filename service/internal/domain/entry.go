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
	EntryStatusPending  = "pending"
	EntryStatusPaid     = "paid"
	EntryStatusComplete = "complete"
)

// Entry defines a user's entry into the prediction league
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

// EntryAgentInjector defines the dependencies required by our EntryAgent
type EntryAgentInjector interface {
	MySQL() coresql.Agent
}

// EntryAgent defines the behaviours for handling Entries
type EntryAgent struct{ EntryAgentInjector }

// CreateEntry handles the creation of a new Entry in the database
func (a EntryAgent) CreateEntry(ctx Context, e Entry, s *Season, realmPIN string) (Entry, error) {
	if s == nil {
		return Entry{}, InternalError{errors.New("invalid season")}
	}

	// check realm PIN is ok
	if err := validateRealmPIN(ctx, realmPIN); err != nil {
		return Entry{}, UnauthorizedError{errors.New("invalid PIN")}
	}

	// check season status is ok
	if s.GetStatus(time.Now()) != SeasonStatusAcceptingEntries {
		return Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	// generate a new entry ID
	uuid, err := uuid.NewV4()
	if err != nil {
		return Entry{}, InternalError{err}
	}

	// override these values
	e.ID = uuid
	e.SeasonID = s.ID
	e.Realm = ctx.GetRealm()
	e.Status = EntryStatusPending
	e.TeamIDSequence = []string{}

	// generate a unique lookup ref
	e.LookupRef, err = generateUniqueLookupRef(a.MySQL())
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, vPackageErrorToValidationError(err, e)
	}

	// check for existing nickname so that we can return a nice error message if it already exists
	existingNicknameEntries, err := dbSelectEntries(a.MySQL(), map[string]interface{}{
		"season_id":        e.SeasonID,
		"realm":            e.Realm,
		"entrant_nickname": e.EntrantNickname,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	// check for existing email so that we can return a nice error message if it already exists
	existingEmailEntries, err := dbSelectEntries(a.MySQL(), map[string]interface{}{
		"season_id":     e.SeasonID,
		"realm":         e.Realm,
		"entrant_email": e.EntrantEmail,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	if len(existingNicknameEntries) > 0 || len(existingEmailEntries) > 0 {
		// entry isn't unique!
		return Entry{}, ConflictError{errors.New("entry already exists")}
	}

	// write to database
	if err := dbInsertEntry(a.MySQL(), &e); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

// UpdateEntry handles the updating of an existing Entry in the database
func (a EntryAgent) UpdateEntry(ctx Context, e Entry) (Entry, error) {
	// ensure that Entry realm matches current realm
	if ctx.GetRealm() != e.Realm {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// sanitise entry
	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, vPackageErrorToValidationError(err, e)
	}

	// don't check if the email or nickname exists at this point, like we did for creating the Entry in the first place
	// in real terms, the api shouldn't allow these fields to be exposed after initial creation
	// there is a db constraint on these two fields anyway, so any values that have changed will be flagged when writing to the db

	// write to database
	if err := dbUpdateEntry(a.MySQL(), &e); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

// sanitiseEntry runs an Entry through the validation package and applies some tidy-up/formatting
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
