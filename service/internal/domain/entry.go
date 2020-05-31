package domain

import (
	"context"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"regexp"
	"sort"
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
func (e EntryAgent) CreateEntry(ctx context.Context, entry models.Entry, s *models.Season) (models.Entry, error) {
	db := e.MySQL()

	if s == nil {
		return models.Entry{}, InternalError{errors.New("invalid season")}
	}

	// check realm PIN is ok
	ctxRealm := RealmFromContext(ctx)
	if !GuardFromContext(ctx).AttemptMatches(ctxRealm.PIN) {
		return models.Entry{}, UnauthorizedError{errors.New("invalid PIN")}
	}

	// check season status is ok
	if s.GetStatus(time.Now()) != models.SeasonStatusAcceptingEntries {
		return models.Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	// generate a new entry ID
	id, err := uuid.NewV4()
	if err != nil {
		return models.Entry{}, InternalError{err}
	}

	// override these values
	entry.ID = id
	entry.SeasonID = s.ID
	entry.RealmName = ctxRealm.Name
	entry.Status = models.EntryStatusPending
	entry.PaymentMethod = sqltypes.NullString{}
	entry.PaymentRef = sqltypes.NullString{}
	entry.EntrySelections = []models.EntrySelection{}
	entry.ApprovedAt = sqltypes.NullTime{}
	entry.CreatedAt = time.Time{}
	entry.UpdatedAt = sqltypes.NullTime{}

	// generate a unique lookup ref
	entry.ShortCode, err = GenerateUniqueShortCode(ctx, db)
	if err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&entry); err != nil {
		return models.Entry{}, err
	}

	entryRepo := repositories.NewEntryDatabaseRepository(db)

	// check for existing nickname so that we can return a nice error message if it already exists
	existingNicknameEntries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id":        entry.SeasonID,
		"realm_name":       entry.RealmName,
		"entrant_nickname": entry.EntrantNickname,
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
		"season_id":     entry.SeasonID,
		"realm_name":    entry.RealmName,
		"entrant_email": entry.EntrantEmail,
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

	// write entry to database
	if err := entryRepo.Insert(ctx, &entry); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	return entry, nil
}

// RetrieveEntryByID handles the retrieval of an existing Entry in the database by its ID
func (e EntryAgent) RetrieveEntryByID(ctx context.Context, id string) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(e.MySQL())
	entrySelectionRepo := repositories.NewEntrySelectionDatabaseRepository(e.MySQL())

	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}
	entry := entries[0]

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry selections
	entry.EntrySelections, err = entrySelectionRepo.Select(ctx, map[string]interface{}{
		"entry_id": entry.ID,
	}, false)
	if err != nil {
		err = domainErrorFromDBError(err)
		switch err.(type) {
		case NotFoundError:
			// all good
		default:
			return models.Entry{}, domainErrorFromDBError(err)
		}
	}

	return entry, nil
}

// RetrieveEntriesBySeasonID handles the retrieval of existing Entries in the database by their Season ID
func (e EntryAgent) RetrieveEntriesBySeasonID(ctx context.Context, seasonID string) ([]models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(e.MySQL())
	entrySelectionRepo := repositories.NewEntrySelectionDatabaseRepository(e.MySQL())

	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id": seasonID,
	}, false)
	if err != nil {
		return []models.Entry{}, domainErrorFromDBError(err)
	}

	for idx := range entries {
		entry := &entries[idx]

		// retrieve and inflate all entry selections
		entry.EntrySelections, err = entrySelectionRepo.Select(ctx, map[string]interface{}{
			"entry_id": entry.ID,
		}, false)
		if err != nil {
			err = domainErrorFromDBError(err)
			switch err.(type) {
			case NotFoundError:
				// all good
			default:
				return []models.Entry{}, domainErrorFromDBError(err)
			}
		}
	}

	return entries, nil
}

// UpdateEntry handles the updating of an existing Entry in the database
func (e EntryAgent) UpdateEntry(ctx context.Context, entry models.Entry) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(e.MySQL())

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := entryRepo.ExistsByID(ctx, entry.ID.String()); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&entry); err != nil {
		return models.Entry{}, err
	}

	// don't check if the email or nickname exists at this point, like we did for creating the Entry in the first place
	// in real terms, the api shouldn't allow these fields to be exposed after initial creation
	// there is a db constraint on these two fields anyway, so any values that have changed will be flagged when writing to the db

	// write to database
	if err := entryRepo.Update(ctx, &entry); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	return entry, nil
}

// AddEntrySelectionToEntry adds the provided EntrySelection to the provided Entry
func (e EntryAgent) AddEntrySelectionToEntry(ctx context.Context, entrySelection models.EntrySelection, entry models.Entry) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(e.MySQL())
	entrySelectionRepo := repositories.NewEntrySelectionDatabaseRepository(e.MySQL())

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := entryRepo.ExistsByID(ctx, entry.ID.String()); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	season, err := datastore.Seasons.GetByID(entry.SeasonID)
	if err != nil {
		return models.Entry{}, NotFoundError{err}
	}

	var invalidTeamIDs, missingTeamIDs, duplicateTeamIDs []string
	var teamIDCount = make(map[string]int)

	for _, teamID := range season.TeamIDs {
		teamIDCount[teamID] = 0
	}

	for _, selectionID := range entrySelection.Rankings.GetIDs() {
		if _, ok := teamIDCount[selectionID]; ok {
			teamIDCount[selectionID]++
			continue
		}
		invalidTeamIDs = append(invalidTeamIDs, selectionID)
	}

	for teamID, count := range teamIDCount {
		switch {
		case count == 0:
			missingTeamIDs = append(missingTeamIDs, teamID)
		case count > 1:
			duplicateTeamIDs = append(duplicateTeamIDs, teamID)
		}
	}

	var validationMsgs []string
	if len(invalidTeamIDs) > 0 {
		for _, teamID := range invalidTeamIDs {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Invalid Team ID: %s", teamID))
		}
	}
	if len(missingTeamIDs) > 0 {
		for _, teamID := range missingTeamIDs {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Missing Team ID: %s", teamID))
		}
	}
	if len(duplicateTeamIDs) > 0 {
		for _, teamID := range duplicateTeamIDs {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Duplicate Team ID: %s", teamID))
		}
	}

	if len(validationMsgs) > 0 {
		return models.Entry{}, ValidationError{Reasons: validationMsgs}
	}

	entrySelection.EntryID = entry.ID

	if err := entrySelectionRepo.Insert(ctx, &entrySelection); err != nil {
		return models.Entry{}, domainErrorFromDBError(err)
	}

	entry.EntrySelections = append(entry.EntrySelections, entrySelection)

	return entry, nil
}

// UpdateEntryPaymentDetails provides a shortcut to updating the payment details for a provided entryID
func (e EntryAgent) UpdateEntryPaymentDetails(ctx context.Context, entryID, paymentMethod, paymentRef string) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(e.MySQL())

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
	if RealmFromContext(ctx).Name != entry.RealmName {
		return models.Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure that Guard value matches Entry ID
	if !GuardFromContext(ctx).AttemptMatches(entry.ID.String()) {
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
func (e EntryAgent) ApproveEntryByShortCode(ctx context.Context, shortCode string) (models.Entry, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(e.MySQL())

	// ensure basic auth has been provided and matches admin credentials
	if !IsBasicAuthSuccessful(ctx) {
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
	if RealmFromContext(ctx).Name != entry.RealmName {
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
func sanitiseEntry(entry *models.Entry) error {
	// only permit alphanumeric characters withing entrant nickname
	regexNickname, err := regexp.Compile("([A-Z]|[a-z]|[0-9])")
	if err != nil {
		return err
	}
	regexNicknameFindResult := strings.Join(regexNickname.FindAllString(entry.EntrantNickname, -1), "")

	// sanitise
	entry.ShortCode = strings.Trim(entry.ShortCode, " ")
	entry.SeasonID = strings.Trim(entry.SeasonID, " ")
	entry.RealmName = strings.Trim(entry.RealmName, " ")
	entry.EntrantName = strings.Trim(entry.EntrantName, " ")
	entry.EntrantNickname = strings.Trim(entry.EntrantNickname, " ")
	entry.EntrantEmail = strings.Trim(entry.EntrantEmail, " ")
	entry.Status = strings.Trim(entry.Status, " ")
	entry.PaymentMethod.String = strings.Trim(entry.PaymentMethod.String, " ")
	entry.PaymentRef.String = strings.Trim(entry.PaymentRef.String, " ")

	var validationMsgs []string

	// validate
	for k, v := range map[string]string{
		"ID":         entry.ID.String(),
		"Short Code": entry.ShortCode,
		"Season ID":  entry.SeasonID,
		"Realm Name": entry.RealmName,
		"Name":       entry.EntrantName,
		"Nickname":   entry.EntrantNickname,
	} {
		if v == "" {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}
	if len(regexNicknameFindResult) != len(entry.EntrantNickname) {
		// regex must have filtered out some invalid characters...
		validationMsgs = append(validationMsgs, "Nickname must only contain alphanumeric characters (A-Z, a-z, 0-9)")
	}
	if len(entry.EntrantNickname) > 12 {
		validationMsgs = append(validationMsgs, "Nickname must be 12 characters or fewer")
	}
	if !isValidEmail(entry.EntrantEmail) {
		validationMsgs = append(validationMsgs, "Email must be a valid email address")
	}
	if !isValidEntryStatus(entry.Status) {
		validationMsgs = append(validationMsgs, fmt.Sprintf("%s is not a valid status", entry.Status))
	}
	if entry.PaymentMethod.Valid && !isValidEntryPaymentMethod(entry.PaymentMethod.String) {
		validationMsgs = append(validationMsgs, fmt.Sprintf("%s is not a valid payment method", entry.PaymentMethod.String))
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	return nil
}

// GenerateUniqueShortCode generates a string that does not already exist as a Lookup Ref
func GenerateUniqueShortCode(ctx context.Context, db coresql.Agent) (string, error) {
	entryRepo := repositories.NewEntryDatabaseRepository(db)

	shortCode := generateRandomAlphaNumericString(4)

	_, err := entryRepo.Select(ctx, map[string]interface{}{
		"short_code": shortCode,
	}, false)
	switch err.(type) {
	case nil:
		// the lookup ref already exists, so we need to generate a new one
		return GenerateUniqueShortCode(ctx, db)
	case repositories.MissingDBRecordError:
		// the lookup ref we have generated is unique, we can return it
		return shortCode, nil
	}
	return "", err
}

// GetEntrySelectionValidAtTimestamp returns the latest EntrySelection that existed at the point of the provided timestamp
func GetEntrySelectionValidAtTimestamp(entrySelections []models.EntrySelection, ts time.Time) (models.EntrySelection, error) {
	desc := entrySelections
	sort.SliceStable(desc, func(i, j int) bool {
		// sort descending
		return desc[j].CreatedAt.Before(desc[i].CreatedAt)
	})

	for _, es := range desc {
		// let's iterate until we get to the first element
		// that was created prior to ts
		if es.CreatedAt.Before(ts) {
			return es, nil
		}
	}

	return models.EntrySelection{}, errors.New("not found")
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
