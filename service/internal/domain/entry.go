package domain

import (
	"context"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"regexp"
	"strings"
	"time"
)

// EntryAgentInjector defines the dependencies required by our EntryAgent
type EntryAgentInjector interface {
	MySQL() coresql.Agent
}

// EntryAgent defines the behaviours for handling Entries
type EntryAgent struct{ EntryAgentInjector }

// CreateEntry handles the creation of a new Entry in the database
func (a EntryAgent) CreateEntry(ctx Context, e models.Entry, s *Season) (models.Entry, error) {
	db := a.MySQL()

	if s == nil {
		return models.Entry{}, InternalError{errors.New("invalid season")}
	}

	// check realm PIN is ok
	if !ctx.Guard.AttemptMatchesTarget(ctx.Realm.PIN) {
		return models.Entry{}, UnauthorizedError{errors.New("invalid PIN")}
	}

	// check season status is ok
	if s.GetStatus(time.Now()) != SeasonStatusAcceptingEntries {
		return models.Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	// generate a new entry ID
	id, err := uuid.NewV4()
	if err != nil {
		return models.Entry{}, InternalError{err}
	}

	// override these values
	e.ID = id
	e.SeasonID = s.ID
	e.RealmName = ctx.Realm.Name
	e.Status = models.EntryStatusPending
	e.PaymentMethod = sqltypes.NullString{}
	e.PaymentRef = sqltypes.NullString{}
	e.TeamIDSequence = []string{}
	e.ApprovedAt = sqltypes.NullTime{}
	e.CreatedAt = time.Time{}
	e.UpdatedAt = sqltypes.NullTime{}

	// generate a unique lookup ref
	e.ShortCode, err = generateUniqueShortCode(ctx, db)
	if err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&e); err != nil {
		return models.Entry{}, err
	}

	entryRepo := repositories.NewEntryDatabaseRepository(db)

	// check for existing nickname so that we can return a nice error message if it already exists
	existingNicknameEntries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id":        e.SeasonID,
		"realm_name":       e.RealmName,
		"entrant_nickname": e.EntrantNickname,
	}, false)
	if err != nil {
		switch err.(type) {
		case repositories.MissingDBRecordError:
			// this is fine
			break
		default:
			return models.Entry{}, domainErrorFromDBError(err)
		}
	}

	// check for existing email so that we can return a nice error message if it already exists
	existingEmailEntries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id":     e.SeasonID,
		"realm_name":    e.RealmName,
		"entrant_email": e.EntrantEmail,
	}, false)
	if err != nil {
		switch err.(type) {
		case repositories.MissingDBRecordError:
			// this is fine
			break
		default:
			return models.Entry{}, domainErrorFromDBError(err)
		}
	}

	if len(existingNicknameEntries) > 0 || len(existingEmailEntries) > 0 {
		// entry isn't unique!
		return models.Entry{}, ConflictError{errors.New("entry already exists")}
	}

	// write to database
	if err := entryRepo.Insert(ctx, &e); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

// RetrieveEntryByID handles the retrieval of an existing Entry in the database by its ID
func (a EntryAgent) RetrieveEntryByID(ctx Context, id string) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(a.MySQL())

	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}
	e := entries[0]

	// ensure that Entry realm matches current realm
	if ctx.Realm.Name != e.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	return e, nil
}

// UpdateEntry handles the updating of an existing Entry in the database
func (a EntryAgent) UpdateEntry(ctx Context, e models.Entry) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(a.MySQL())

	// ensure that Entry realm matches current realm
	if ctx.Realm.Name != e.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := entryRepo.ExistsByID(ctx, e.ID.String()); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&e); err != nil {
		return models.Entry{}, err
	}

	// don't check if the email or nickname exists at this point, like we did for creating the Entry in the first place
	// in real terms, the api shouldn't allow these fields to be exposed after initial creation
	// there is a db constraint on these two fields anyway, so any values that have changed will be flagged when writing to the db

	// write to database
	if err := entryRepo.Update(ctx, &e); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	return e, nil
}

// UpdateEntryPaymentDetails provides a shortcut to updating the payment details for a provided entryID
func (a EntryAgent) UpdateEntryPaymentDetails(ctx Context, entryID, paymentMethod, paymentRef string) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(a.MySQL())

	// ensure that payment method is valid
	if !isValidEntryPaymentMethod(paymentMethod) {
		return models.Entry{}, ValidationError{
			Reasons: []string{"Invalid payment method"},
		}
	}

	// ensure that payment ref is not empty
	if paymentRef == "" {
		return models.Entry{}, ValidationError{
			Reasons: []string{"Invalid payment ref"},
		}
	}

	// retrieve entry
	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"id": entryID,
	}, false)
	if err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	if len(entries) != 1 {
		return models.Entry{}, InternalError{errors.New("entries count other than 1")}
	}

	entry := entries[0]

	// ensure that Entry realm matches current realm
	if ctx.Realm.Name != entry.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure that Guard value matches Entry ID
	if !ctx.Guard.AttemptMatchesTarget(entry.ID.String()) {
		return models.Entry{}, ValidationError{
			Reasons: []string{"Invalid Entry ID"},
		}
	}

	// check Entry status
	if entry.Status != models.EntryStatusPending {
		return models.Entry{}, ConflictError{errors.New("payment details can only be added if entry status is pending")}
	}

	entry.PaymentMethod = sqltypes.ToNullString(&paymentMethod)
	entry.PaymentRef = sqltypes.ToNullString(&paymentRef)
	entry.Status = models.EntryStatusPaid

	// write to database
	if err := entryRepo.Update(ctx, &entry); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	return entry, nil
}

// ApproveEntryByShortCode provides a shortcut to approving an entry by its short code
func (a EntryAgent) ApproveEntryByShortCode(ctx Context, shortCode string) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(a.MySQL())

	// ensure basic auth has been provided and matches admin credentials
	if !ctx.BasicAuthSuccessful {
		return models.Entry{}, UnauthorizedError{}
	}

	// retrieve entry
	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"short_code": shortCode,
	}, false)
	if err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	if len(entries) != 1 {
		return models.Entry{}, InternalError{errors.New("entries count other than 1")}
	}

	entry := entries[0]

	// ensure that Entry realm matches current realm
	if ctx.Realm.Name != entry.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// check Entry status
	switch entry.Status {
	case models.EntryStatusPaid, models.EntryStatusReady:
		// all good
	default:
		return models.Entry{}, ConflictError{errors.New("entry can only be approved if status is pending or ready")}
	}

	// check if Entry has already been approved
	if entry.ApprovedAt.Valid {
		return models.Entry{}, ConflictError{errors.New("entry has already been approved")}
	}

	entry.ApprovedAt = sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	// write to database
	if err := entryRepo.Update(ctx, &entry); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	return entry, nil
}

// sanitiseEntry sanitises and validates an Entry
func sanitiseEntry(e *models.Entry) error {
	// only permit alphanumeric characters withing entrant nickname
	regexNickname, err := regexp.Compile("([A-Z]|[a-z]|[0-9])")
	if err != nil {
		return err
	}
	regexNicknameFindResult := strings.Join(regexNickname.FindAllString(e.EntrantNickname, -1), "")

	// sanitise
	e.ShortCode = strings.Trim(e.ShortCode, " ")
	e.SeasonID = strings.Trim(e.SeasonID, " ")
	e.RealmName = strings.Trim(e.RealmName, " ")
	e.EntrantName = strings.Trim(e.EntrantName, " ")
	e.EntrantNickname = strings.Trim(e.EntrantNickname, " ")
	e.EntrantEmail = strings.Trim(e.EntrantEmail, " ")
	e.Status = strings.Trim(e.Status, " ")
	e.PaymentMethod.String = strings.Trim(e.PaymentMethod.String, " ")
	e.PaymentRef.String = strings.Trim(e.PaymentRef.String, " ")

	var validationMsgs []string

	// validate
	for k, v := range map[string]string{
		"ID":         e.ID.String(),
		"Short Code": e.ShortCode,
		"Season ID":  e.SeasonID,
		"Realm Name": e.RealmName,
		"Name":       e.EntrantName,
		"Nickname":   e.EntrantNickname,
	} {
		if v == "" {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}
	if len(regexNicknameFindResult) != len(e.EntrantNickname) {
		// regex must have filtered out some invalid characters...
		validationMsgs = append(validationMsgs, "Nickname must only contain alphanumeric characters (A-Z, a-z, 0-9)")
	}
	if len(e.EntrantNickname) > 12 {
		validationMsgs = append(validationMsgs, "Nickname must be 12 characters or fewer")
	}
	if !isValidEmail(e.EntrantEmail) {
		validationMsgs = append(validationMsgs, "Email must be a valid email address")
	}
	if !isValidEntryStatus(e.Status) {
		validationMsgs = append(validationMsgs, fmt.Sprintf("%s is not a valid status", e.Status))
	}
	if e.PaymentMethod.Valid && !isValidEntryPaymentMethod(e.PaymentMethod.String) {
		validationMsgs = append(validationMsgs, fmt.Sprintf("%s is not a valid payment method", e.PaymentMethod.String))
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	return nil
}

// generateUniqueShortCode generates a string that does not already exist as a Lookup Ref
func generateUniqueShortCode(ctx context.Context, db coresql.Agent) (string, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(db)

	shortCode := generateRandomAlphaNumericString(4)

	_, err := entryRepo.Select(ctx, map[string]interface{}{
		"short_code": shortCode,
	}, false)
	switch err.(type) {
	case nil:
		// the lookup ref already exists, so we need to generate a new one
		return generateUniqueShortCode(ctx, db)
	case repositories.MissingDBRecordError:
		// the lookup ref we have generated is unique, we can return it
		return shortCode, nil
	}
	return "", err
}

func isValidEmail(email string) bool {
	var pattern, err = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if err != nil {
		return false
	}

	if !pattern.MatchString(email) {
		return false
	}

	return true
}

func isValidEntryStatus(status string) bool {
	switch status {
	case models.EntryStatusPending, models.EntryStatusPaid, models.EntryStatusReady:
		return true
	}

	return false
}

func isValidEntryPaymentMethod(paymentMethod string) bool {
	switch paymentMethod {
	case models.EntryPaymentMethodPayPal, models.EntryPaymentMethodOther:
		return true
	}

	return false
}
