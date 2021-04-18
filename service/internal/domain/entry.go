package domain

import (
	"context"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"math/rand"
	repofac2 "prediction-league/service/internal/repositories/repofac"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// EntryStatusPending represents an Entry whose status is PENDING
	EntryStatusPending = "pending"
	// EntryStatusPaid represents an Entry whose status is PAID
	EntryStatusPaid = "paid"
	// EntryStatusReady represents an Entry whose status is READY
	EntryStatusReady = "ready"

	// EntryPaymentMethodPayPal represents an Entry that has been paid via PAYPAL
	EntryPaymentMethodPayPal = "paypal"
	// EntryPaymentMethodOther represents an Entry that has been paid by OTHER means
	EntryPaymentMethodOther = "other"

	// shortCodeLength represents the number of characters that a short code will contain
	shortCodeLength = 6
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

// EntryAgentInjector defines the dependencies required by our EntryAgent
type EntryAgentInjector interface {
	MySQL() coresql.Agent
}

// EntryAgent defines the behaviours for handling Entries
type EntryAgent struct{ EntryAgentInjector }

// CreateEntry handles the creation of a new Entry in the database
func (e EntryAgent) CreateEntry(ctx context.Context, entry Entry, s *Season) (Entry, error) {
	db := e.MySQL()

	if s == nil {
		return Entry{}, InternalError{errors.New("invalid season")}
	}

	// check realm PIN is ok
	ctxRealm := RealmFromContext(ctx)
	if !GuardFromContext(ctx).AttemptMatches(ctxRealm.PIN) {
		return Entry{}, UnauthorizedError{errors.New("invalid PIN")}
	}

	// check if season is currently accepting entries
	if !s.GetState(TimestampFromContext(ctx)).IsAcceptingEntries {
		return Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	// generate a new entry ID
	id, err := uuid.NewV4()
	if err != nil {
		return Entry{}, InternalError{err}
	}

	// override these values
	entry.ID = id
	entry.SeasonID = s.ID
	entry.RealmName = ctxRealm.Name
	entry.Status = EntryStatusPending
	entry.PaymentMethod = sqltypes.NullString{}
	entry.PaymentRef = sqltypes.NullString{}
	entry.EntryPredictions = []EntryPrediction{}
	entry.ApprovedAt = sqltypes.NullTime{}
	entry.CreatedAt = time.Time{}
	entry.UpdatedAt = sqltypes.NullTime{}

	// generate a unique lookup ref
	entry.ShortCode, err = GenerateUniqueShortCode(ctx, db)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&entry); err != nil {
		return Entry{}, err
	}

	entryRepo := repofac2.NewEntryDatabaseRepository(db)

	// check for existing nickname so that we can return a nice error message if it already exists
	existingNicknameEntries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id":        entry.SeasonID,
		"realm_name":       entry.RealmName,
		"entrant_nickname": entry.EntrantNickname,
	}, false)
	if err != nil {
		switch err.(type) {
		case MissingDBRecordError:
			// this is fine
			break
		default:
			return Entry{}, domainErrorFromRepositoryError(err)
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
		case MissingDBRecordError:
			// this is fine
			break
		default:
			return Entry{}, domainErrorFromRepositoryError(err)
		}
	}

	if len(existingNicknameEntries) > 0 || len(existingEmailEntries) > 0 {
		// entry isn't unique!
		return Entry{}, ConflictError{errors.New("entry already exists")}
	}

	// write entry to database
	if err := entryRepo.Insert(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// RetrieveEntryByID handles the retrieval of an existing Entry in the database by its ID
func (e EntryAgent) RetrieveEntryByID(ctx context.Context, id string) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(e.MySQL())

	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}
	entry := entries[0]

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry predictions
	entry.EntryPredictions, err = entryPredictionRepo.Select(ctx, map[string]interface{}{
		"entry_id": entry.ID,
	}, false)
	if err != nil {
		err = domainErrorFromRepositoryError(err)
		switch err.(type) {
		case NotFoundError:
			// all good
		default:
			return Entry{}, domainErrorFromRepositoryError(err)
		}
	}

	return entry, nil
}

// RetrieveEntryByEntrantEmail handles the retrieval of an existing Entry in the database by its email
func (e EntryAgent) RetrieveEntryByEntrantEmail(ctx context.Context, email, seasonID, realmName string) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(e.MySQL())

	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id":     seasonID,
		"realm_name":    realmName,
		"entrant_email": email,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}
	entry := entries[0]

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry predictions
	entry.EntryPredictions, err = entryPredictionRepo.Select(ctx, map[string]interface{}{
		"entry_id": entry.ID,
	}, false)
	if err != nil {
		err = domainErrorFromRepositoryError(err)
		switch err.(type) {
		case NotFoundError:
			// all good
		default:
			return Entry{}, domainErrorFromRepositoryError(err)
		}
	}

	return entry, nil
}

// RetrieveEntryByEntrantNickname handles the retrieval of an existing Entry in the database by its nickname
func (e EntryAgent) RetrieveEntryByEntrantNickname(ctx context.Context, nickname, seasonID, realmName string) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(e.MySQL())

	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"season_id":        seasonID,
		"realm_name":       realmName,
		"entrant_nickname": nickname,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}
	entry := entries[0]

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry predictions
	entry.EntryPredictions, err = entryPredictionRepo.Select(ctx, map[string]interface{}{
		"entry_id": entry.ID,
	}, false)
	if err != nil {
		err = domainErrorFromRepositoryError(err)
		switch err.(type) {
		case NotFoundError:
			// all good
		default:
			return Entry{}, domainErrorFromRepositoryError(err)
		}
	}

	return entry, nil
}

// RetrieveEntriesBySeasonID handles the retrieval of existing Entries in the database by their Season ID
func (e EntryAgent) RetrieveEntriesBySeasonID(ctx context.Context, seasonID string, onlyApproved bool) ([]Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(e.MySQL())

	criteria := map[string]interface{}{
		"season_id": seasonID,
	}

	if onlyApproved {
		criteria["approved_at"] = DBQueryCondition{
			Operator: "IS NOT NULL",
		}
	}

	entries, err := entryRepo.Select(ctx, criteria, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	for idx := range entries {
		entry := &entries[idx]

		// retrieve and inflate all entry predictions
		entry.EntryPredictions, err = entryPredictionRepo.Select(ctx, map[string]interface{}{
			"entry_id": entry.ID,
		}, false)
		if err != nil {
			err = domainErrorFromRepositoryError(err)
			switch err.(type) {
			case NotFoundError:
				// all good
			default:
				return nil, domainErrorFromRepositoryError(err)
			}
		}
	}

	return entries, nil
}

// UpdateEntry handles the updating of an existing Entry in the database
func (e EntryAgent) UpdateEntry(ctx context.Context, entry Entry) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := entryRepo.ExistsByID(ctx, entry.ID.String()); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	// sanitise entry
	if err := sanitiseEntry(&entry); err != nil {
		return Entry{}, err
	}

	// don't check if the email or nickname exists at this point, like we did for creating the Entry in the first place
	// in real terms, the api shouldn't allow these fields to be exposed after initial creation
	// there is a db constraint on these two fields anyway, so any values that have changed will be flagged when writing to the db

	// write to database
	if err := entryRepo.Update(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// AddEntryPredictionToEntry adds the provided EntryPrediction to the provided Entry
func (e EntryAgent) AddEntryPredictionToEntry(ctx context.Context, entryPrediction EntryPrediction, entry Entry) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(e.MySQL())

	// check short code is ok
	if !GuardFromContext(ctx).AttemptMatches(entry.ShortCode) {
		return Entry{}, UnauthorizedError{errors.New("invalid short code")}
	}

	// ensure that entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := entryRepo.ExistsByID(ctx, entry.ID.String()); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	// retrieve the entry's Season
	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return Entry{}, NotFoundError{err}
	}

	// check if season is currently accepting entries
	if !season.GetState(TimestampFromContext(ctx)).IsAcceptingPredictions {
		return Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	var invalidRankingIDs, missingRankingIDs, duplicateRankingIDs []string
	var teamIDCount = make(map[string]int)

	for _, teamID := range season.TeamIDs {
		teamIDCount[teamID] = 0
	}

	// make sure we have no invalid IDs
	for _, predictionID := range entryPrediction.Rankings.GetIDs() {
		if _, ok := teamIDCount[predictionID]; ok {
			teamIDCount[predictionID]++
			continue
		}
		invalidRankingIDs = append(invalidRankingIDs, predictionID)
	}

	// make sure we have no missing or duplicate IDs
	for teamID, count := range teamIDCount {
		switch {
		case count == 0:
			missingRankingIDs = append(missingRankingIDs, teamID)
		case count > 1:
			duplicateRankingIDs = append(duplicateRankingIDs, teamID)
		}
	}

	var validationMsgs []string
	if len(invalidRankingIDs) > 0 {
		for _, teamID := range invalidRankingIDs {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Invalid Team ID: %s", teamID))
		}
	}
	if len(missingRankingIDs) > 0 {
		for _, teamID := range missingRankingIDs {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Missing Team ID: %s", teamID))
		}
	}
	if len(duplicateRankingIDs) > 0 {
		for _, teamID := range duplicateRankingIDs {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Duplicate Team ID: %s", teamID))
		}
	}

	if len(validationMsgs) > 0 {
		return Entry{}, ValidationError{Reasons: validationMsgs}
	}

	// generate a new entry ID
	id, err := uuid.NewV4()
	if err != nil {
		return Entry{}, InternalError{err}
	}

	entryPrediction.ID = id
	entryPrediction.EntryID = entry.ID

	if err := entryPredictionRepo.Insert(ctx, &entryPrediction); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	entry.EntryPredictions = append(entry.EntryPredictions, entryPrediction)

	return entry, nil
}

// UpdateEntryPaymentDetails provides a shortcut to updating the payment details for a provided entryID
func (e EntryAgent) UpdateEntryPaymentDetails(ctx context.Context, entryID, paymentMethod, paymentRef string, acceptsOther bool) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())

	// ensure that payment method is valid
	if !isValidEntryPaymentMethod(paymentMethod) {
		return Entry{}, ValidationError{
			Reasons: []string{"invalid payment method"},
		}
	}

	// ensure that we are only able to explicitly accept payment method "other"
	if paymentMethod == EntryPaymentMethodOther && !acceptsOther {
		return Entry{}, ConflictError{fmt.Errorf("cannot accept payment method: %s", EntryPaymentMethodOther)}
	}

	// ensure that payment ref is not empty
	if paymentRef == "" {
		return Entry{}, ValidationError{
			Reasons: []string{"invalid payment ref"},
		}
	}

	// retrieve entry
	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"id": entryID,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	if len(entries) != 1 {
		return Entry{}, InternalError{errors.New("entries count other than 1")}
	}

	entry := entries[0]

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure that Guard value matches Entry ID
	if !GuardFromContext(ctx).AttemptMatches(entry.ShortCode) {
		return Entry{}, ValidationError{
			Reasons: []string{"Invalid Entry ID"},
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
	if err := entryRepo.Update(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// ApproveEntryByShortCode provides a shortcut to approving an entry by its short code
func (e EntryAgent) ApproveEntryByShortCode(ctx context.Context, shortCode string) (Entry, error) {
	entryRepo := repofac2.NewEntryDatabaseRepository(e.MySQL())

	// ensure basic auth has been provided and matches admin credentials
	if !IsBasicAuthSuccessful(ctx) {
		return Entry{}, UnauthorizedError{}
	}

	// retrieve entry
	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"short_code": shortCode,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	if len(entries) != 1 {
		return Entry{}, InternalError{fmt.Errorf("entries count other than 1: %d", len(entries))}
	}

	entry := entries[0]

	// ensure that Entry realm matches current realm
	if RealmFromContext(ctx).Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// check Entry status
	switch entry.Status {
	case EntryStatusPaid, EntryStatusReady:
		// all good
	default:
		return Entry{}, ConflictError{fmt.Errorf(
			"entry can only be approved if status is pending or ready: status is %s",
			entry.Status,
		)}
	}

	// check if Entry has already been approved
	if entry.ApprovedAt.Valid {
		return Entry{}, ConflictError{errors.New("entry has already been approved")}
	}

	entry.ApprovedAt = sqltypes.ToNullTime(TimestampFromContext(ctx).Truncate(time.Second))

	// write to database
	if err := entryRepo.Update(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// RetrieveEntryPredictionByTimestamp returns the entry prediction affiliated with the provided entry id that is valid at the point the provided timestamp occurs
func (e EntryAgent) RetrieveEntryPredictionByTimestamp(ctx context.Context, entry Entry, ts time.Time) (EntryPrediction, error) {
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(e.MySQL())

	// retrieve entry prediction
	entryPrediction, err := entryPredictionRepo.SelectByEntryIDAndTimestamp(ctx, entry.ID.String(), ts)
	if err != nil {
		return EntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return entryPrediction, nil
}

// RetrieveEntryPredictionsForActiveSeasonByTimestamp retrieves all entry predictions active at the provided timestamp
// for the provided active season
func (e EntryAgent) RetrieveEntryPredictionsForActiveSeasonByTimestamp(
	ctx context.Context,
	season Season,
	timestamp *time.Time,
) ([]EntryPrediction, error) {
	ts := time.Now()
	if timestamp != nil {
		ts = *timestamp
	}

	// ensure that season is active based on provided timestamp
	if !season.Active.HasBegunBy(ts) || season.Active.HasElapsedBy(ts) {
		return nil, ConflictError{fmt.Errorf("season not active: id %s", season.ID)}
	}

	// retrieve all entries for provided season
	seasonEntries, err := e.RetrieveEntriesBySeasonID(ctx, season.ID, false)
	if err != nil {
		return nil, err
	}

	// get the entry prediction valid at the provided timestamp for each of the entries we've just retrieved
	var currentEntryPredictions []EntryPrediction
	for _, entry := range seasonEntries {
		es, err := getEntryPredictionValidAtTimestamp(entry.EntryPredictions, ts)
		if err != nil {
			// error indicates that no prediction has been found, so just ignore this entry and continue to the next
			continue
		}

		currentEntryPredictions = append(currentEntryPredictions, es)
	}

	return currentEntryPredictions, nil
}

// sanitiseEntry sanitises and validates an Entry
func sanitiseEntry(entry *Entry) error {
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
	entryRepo := repofac2.NewEntryDatabaseRepository(db)

	shortCode := generateRandomAlphaNumericString(shortCodeLength)

	_, err := entryRepo.Select(ctx, map[string]interface{}{
		"short_code": shortCode,
	}, false)
	switch err.(type) {
	case nil:
		// the short code already exists, so we need to generate a new one
		return GenerateUniqueShortCode(ctx, db)
	case MissingDBRecordError:
		// the lookup ref we have generated is unique, we can return it
		return shortCode, nil
	}
	return "", err
}

// getEntryPredictionValidAtTimestamp returns the latest EntryPrediction that existed at the point of the provided timestamp
func getEntryPredictionValidAtTimestamp(entryPredictions []EntryPrediction, ts time.Time) (EntryPrediction, error) {
	desc := entryPredictions
	sort.SliceStable(desc, func(i, j int) bool {
		// sort descending
		return desc[j].CreatedAt.Before(desc[i].CreatedAt)
	})

	for _, ep := range desc {
		// let's iterate until we get to the first element
		// that was created prior to ts
		if ep.CreatedAt.Before(ts) {
			return ep, nil
		}
	}

	return EntryPrediction{}, fmt.Errorf("entry prediction by timestamp %+v: not found", ts)
}

// generateRandomAlphaNumericString returns a randomised string of given length
func generateRandomAlphaNumericString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}

	return string(b)
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
