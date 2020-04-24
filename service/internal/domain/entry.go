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
	EntryStatusPending = "pending"
	EntryStatusPaid    = "paid"
	EntryStatusReady   = "ready"

	EntryPaymentMethodPayPal = "paypal"
	EntryPaymentMethodOther  = "other"
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
	Status          string              `db:"status" v:"func:isValidEntryStatus"`
	PaymentMethod   sqltypes.NullString `db:"payment_method" v:"func:isValidEntryPaymentMethod"`
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
func (a EntryAgent) CreateEntry(ctx Context, e Entry, s *Season) (Entry, error) {
	db := a.MySQL()

	if s == nil {
		return Entry{}, InternalError{errors.New("invalid season")}
	}

	// check realm PIN is ok
	if err := validateRealmPIN(ctx, ctx.GetGuardValue()); err != nil {
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
	e.PaymentMethod = sqltypes.NullString{}
	e.PaymentRef = sqltypes.NullString{}
	e.TeamIDSequence = []string{}

	// generate a unique lookup ref
	e.LookupRef, err = generateUniqueLookupRef(db)
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, vPackageErrorToValidationError(err, e)
	}

	// check for existing nickname so that we can return a nice error message if it already exists
	existingNicknameEntries, err := dbSelectEntries(db, map[string]interface{}{
		"season_id":        e.SeasonID,
		"realm":            e.Realm,
		"entrant_nickname": e.EntrantNickname,
	}, false)
	if err != nil {
		switch err.(type) {
		case dbMissingRecordError:
			// this is fine
			break
		default:
			return Entry{}, domainErrorFromDBError(err)
		}
	}

	// check for existing email so that we can return a nice error message if it already exists
	existingEmailEntries, err := dbSelectEntries(db, map[string]interface{}{
		"season_id":     e.SeasonID,
		"realm":         e.Realm,
		"entrant_email": e.EntrantEmail,
	}, false)
	if err != nil {
		switch err.(type) {
		case dbMissingRecordError:
			// this is fine
			break
		default:
			return Entry{}, domainErrorFromDBError(err)
		}
	}

	if len(existingNicknameEntries) > 0 || len(existingEmailEntries) > 0 {
		// entry isn't unique!
		return Entry{}, ConflictError{errors.New("entry already exists")}
	}

	// write to database
	if err := DBInsertEntry(db, &e); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

// TODO - require entry to be approved (separate to status)

// UpdateEntry handles the updating of an existing Entry in the database
func (a EntryAgent) UpdateEntry(ctx Context, e Entry) (Entry, error) {
	db := a.MySQL()

	// ensure that Entry realm matches current realm
	if ctx.GetRealm() != e.Realm {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := dbEntryExists(db, e.ID.String()); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&e); err != nil {
		return Entry{}, vPackageErrorToValidationError(err, e)
	}

	// don't check if the email or nickname exists at this point, like we did for creating the Entry in the first place
	// in real terms, the api shouldn't allow these fields to be exposed after initial creation
	// there is a db constraint on these two fields anyway, so any values that have changed will be flagged when writing to the db

	// write to database
	if err := dbUpdateEntry(db, &e); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

// UpdateEntryPaymentDetails provides a shortcut to updating the payment details for a provided entryID
func (a EntryAgent) UpdateEntryPaymentDetails(ctx Context, entryID, paymentMethod, paymentRef string) (Entry, error) {
	db := a.MySQL()

	// ensure that payment method is valid
	if !isValidEntryPaymentMethod(paymentMethod) {
		return Entry{}, ValidationError{
			Reasons: []string{"Invalid payment method"},
			Fields:  []string{"payment_method"},
		}
	}

	// ensure that payment ref is not empty
	if paymentRef == "" {
		return Entry{}, ValidationError{
			Reasons: []string{"Invalid payment ref"},
			Fields:  []string{"payment_ref"},
		}
	}

	// retrieve entry
	entries, err := dbSelectEntries(db, map[string]interface{}{
		"id": entryID,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	if len(entries) != 1 {
		return Entry{}, InternalError{errors.New("entries count other than 1")}
	}

	entry := entries[0]

	// ensure that Entry realm matches current realm
	if ctx.GetRealm() != entry.Realm {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure that Guard value matches Entry LookupRef
	if ctx.GetGuardValue() != entry.LookupRef {
		return Entry{}, ValidationError{
			Reasons: []string{"Invalid Lookup Ref"},
			Fields:  []string{"lookup_ref"},
		}
	}

	// check Entry status
	if entry.Status != EntryStatusPending {
		return Entry{}, ConflictError{errors.New("payment details can only be added if entry status is pending")}
	}

	entry.PaymentMethod = sqltypes.ToNullString(&paymentMethod)
	entry.PaymentRef = sqltypes.ToNullString(&paymentRef)
	entry.Status = EntryStatusPaid

	// write to database
	if err := dbUpdateEntry(db, &entry); err != nil {
		return Entry{}, domainErrorFromDBError(err)
	}

	return entry, nil
}

// sanitiseEntry runs an Entry through the validation package and applies some tidy-up/formatting
func sanitiseEntry(e *Entry) error {
	if err := v.Struct(e); err != nil {
		return vPackageErrorToValidationError(err, *e)
	}

	e.EntrantName = strings.Trim(e.EntrantName, " ")
	e.EntrantNickname = strings.Trim(e.EntrantNickname, " ")
	e.EntrantEmail = strings.Trim(e.EntrantEmail, " ")

	return nil
}

// generateUniqueLookupRef generates a string that does not already exist as a Lookup Ref
func generateUniqueLookupRef(db coresql.Agent) (string, error) {
	lookupRef := generateRandomAlphaNumericString(4)

	_, err := dbSelectEntries(db, map[string]interface{}{
		"lookup_ref": lookupRef,
	}, false)
	switch err.(type) {
	case nil:
		// the lookup ref already exists, so we need to generate a new one
		return generateUniqueLookupRef(db)
	case dbMissingRecordError:
		// the lookup ref we have generated is unique, we can return it
		return lookupRef, nil
	}
	return "", wrapDBError(err)
}

func isValidEntryStatus(status string) bool {
	switch status {
	case EntryStatusPending, EntryStatusPaid, EntryStatusReady:
		return true
	}

	return false
}

func isValidEntryPaymentMethod(paymentMethod string) bool {
	switch paymentMethod {
	case EntryPaymentMethodPayPal, EntryPaymentMethodOther:
		return true
	}

	return false
}
