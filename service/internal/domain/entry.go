package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// EntryStatusPending represents an Entry whose status is PENDING
	EntryStatusPending = "pending"
	// EntryStatusPaid represents an Entry whose status is PAID
	EntryStatusPaid = "paid"

	// EntryPaymentMethodPayPal represents an Entry that has been paid via PAYPAL
	EntryPaymentMethodPayPal = "paypal"
	// EntryPaymentMethodOther represents an Entry that has been paid by OTHER means
	EntryPaymentMethodOther = "other"

	// RankingLimitNone defines the value used to represent no limit to ranking changes
	RankingLimitNone = -1
	// RankingLimitRegular defines the standard value for limiting ranking changes
	RankingLimitRegular = 2
)

// Entry defines a user's entry into the prediction league
type Entry struct {
	ID               uuid.UUID `db:"id"`
	SeasonID         string    `db:"season_id"`
	RealmName        string    `db:"realm_name"`
	EntrantName      string    `db:"entrant_name"`
	EntrantNickname  string    `db:"entrant_nickname"`
	EntrantEmail     string    `db:"entrant_email"`
	Status           string    `db:"status"`
	PaymentMethod    *string   `db:"payment_method"`
	PaymentRef       *string   `db:"payment_ref"`
	EntryPredictions []EntryPrediction
	ApprovedAt       *time.Time `db:"approved_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at"`
}

func (e *Entry) IsApproved() bool {
	return e.ApprovedAt != nil
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

// EntryRepository defines the interface for transacting with our Entry data source
type EntryRepository interface {
	Insert(ctx context.Context, entry *Entry) error
	Update(ctx context.Context, entry *Entry) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]Entry, error)
	SelectBySeasonIDAndApproved(ctx context.Context, seasonID string, approved bool) ([]Entry, error)
	ExistsByID(ctx context.Context, id string) error
}

// EntryPredictionRepository defines the interface for transacting with our EntryPredictions data source
type EntryPredictionRepository interface {
	Insert(ctx context.Context, entryPrediction *EntryPrediction) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]EntryPrediction, error)
	SelectByEntryIDAndTimestamp(ctx context.Context, entryID string, ts time.Time) (EntryPrediction, error)
	ExistsByID(ctx context.Context, id string) error
}

// EntryAgent defines the behaviours for handling Entries
type EntryAgent struct {
	er  EntryRepository
	epr EntryPredictionRepository
	sr  StandingsRepository
	sc  SeasonCollection
	cl  Clock
}

// CreateEntry handles the creation of a new Entry in the database
func (e *EntryAgent) CreateEntry(ctx context.Context, entry Entry, s *Season) (Entry, error) {
	if s == nil {
		return Entry{}, InternalError{errors.New("invalid season")}
	}

	// check realm PIN is ok
	realm := RealmFromContext(ctx)
	if !GuardFromContext(ctx).AttemptMatches(realm.Config.PIN) {
		return Entry{}, UnauthorizedError{errors.New("invalid PIN")}
	}

	// check if season is currently accepting entries
	if s.GetState(e.cl.Now()).EntriesStatus != SeasonStateActive {
		return Entry{}, ConflictError{errors.New("season is not currently accepting entries")}
	}

	// generate a new entry ID
	id, err := uuid.NewRandom()
	if err != nil {
		return Entry{}, InternalError{err}
	}

	// override these values
	entry.ID = id
	entry.SeasonID = s.ID
	entry.RealmName = realm.Config.Name
	entry.Status = EntryStatusPending
	entry.PaymentMethod = nil
	entry.PaymentRef = nil
	entry.EntryPredictions = []EntryPrediction{}
	entry.ApprovedAt = nil
	entry.CreatedAt = time.Time{}
	entry.UpdatedAt = nil

	// sanitise entry
	if err := sanitiseEntry(&entry); err != nil {
		return Entry{}, err
	}

	// check for existing nickname so that we can return a nice error message if it already exists
	existingNicknameEntries, err := e.er.Select(ctx, map[string]interface{}{
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
	existingEmailEntries, err := e.er.Select(ctx, map[string]interface{}{
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
	if err := e.er.Insert(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// RetrieveEntryByID handles the retrieval of an existing Entry in the database by its ID
func (e *EntryAgent) RetrieveEntryByID(ctx context.Context, id string) (Entry, error) {
	entries, err := e.er.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}
	entry := entries[0]

	realm := RealmFromContext(ctx)

	// ensure that Entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry predictions
	entry.EntryPredictions, err = e.epr.Select(ctx, map[string]interface{}{
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
func (e *EntryAgent) RetrieveEntryByEntrantEmail(ctx context.Context, email, seasonID, realmName string) (Entry, error) {
	entries, err := e.er.Select(ctx, map[string]interface{}{
		"season_id":     seasonID,
		"realm_name":    realmName,
		"entrant_email": email,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}
	entry := entries[0]

	realm := RealmFromContext(ctx)

	// ensure that Entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry predictions
	entry.EntryPredictions, err = e.epr.Select(ctx, map[string]interface{}{
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
func (e *EntryAgent) RetrieveEntryByEntrantNickname(ctx context.Context, nickname, seasonID, realmName string) (Entry, error) {
	entries, err := e.er.Select(ctx, map[string]interface{}{
		"season_id":        seasonID,
		"realm_name":       realmName,
		"entrant_nickname": nickname,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}
	entry := entries[0]

	realm := RealmFromContext(ctx)

	// ensure that Entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// retrieve and inflate all entry predictions
	entry.EntryPredictions, err = e.epr.Select(ctx, map[string]interface{}{
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
func (e *EntryAgent) RetrieveEntriesBySeasonID(ctx context.Context, seasonID string, onlyApproved bool) ([]Entry, error) {
	criteria := map[string]interface{}{
		"season_id": seasonID,
	}

	if onlyApproved {
		criteria["approved_at"] = DBQueryCondition{
			Operator: "IS NOT NULL",
		}
	}

	entries, err := e.er.Select(ctx, criteria, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	for idx := range entries {
		entry := &entries[idx]

		// retrieve and inflate all entry predictions
		entry.EntryPredictions, err = e.epr.Select(ctx, map[string]interface{}{
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
func (e *EntryAgent) UpdateEntry(ctx context.Context, entry Entry) (Entry, error) {
	realm := RealmFromContext(ctx)

	// ensure that Entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := e.er.ExistsByID(ctx, entry.ID.String()); err != nil {
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
	if err := e.er.Update(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// AddEntryPredictionToEntry adds the provided EntryPrediction to the provided Entry
func (e *EntryAgent) AddEntryPredictionToEntry(ctx context.Context, entryPrediction EntryPrediction, entry Entry) (Entry, error) {
	realm := RealmFromContext(ctx)

	// ensure that entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// ensure the entry exists
	if err := e.er.ExistsByID(ctx, entry.ID.String()); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	// retrieve the entry's Season
	season, err := e.sc.GetByID(entry.SeasonID)
	if err != nil {
		return Entry{}, NotFoundError{err}
	}

	// check if season is currently accepting predictions
	state := season.GetState(e.cl.Now())
	if state.PredictionsStatus != SeasonStateActive {
		return Entry{}, ConflictError{errors.New("season is not currently accepting predictions")}
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
	id, err := uuid.NewRandom()
	if err != nil {
		return Entry{}, InternalError{err}
	}

	entryPrediction.ID = id
	entryPrediction.EntryID = entry.ID

	if err := e.epr.Insert(ctx, &entryPrediction); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	entry.EntryPredictions = append(entry.EntryPredictions, entryPrediction)

	return entry, nil
}

// UpdateEntryPaymentDetails provides a shortcut to updating the payment details for a provided entryID
func (e *EntryAgent) UpdateEntryPaymentDetails(ctx context.Context, entryID, paymentMethod, paymentRef string, acceptsOther bool) (Entry, error) {
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
	entries, err := e.er.Select(ctx, map[string]interface{}{
		"id": entryID,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	if len(entries) != 1 {
		return Entry{}, InternalError{errors.New("entries count other than 1")}
	}

	entry := entries[0]

	realm := RealmFromContext(ctx)

	// ensure that Entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// check Entry status
	if entry.Status != EntryStatusPending {
		return Entry{}, ConflictError{errors.New("payment details can only be added if entry status is pending")}
	}

	entry.PaymentMethod = &paymentMethod
	entry.PaymentRef = &paymentRef
	entry.Status = EntryStatusPaid

	// write to database
	if err := e.er.Update(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// ApproveEntryByID provides a shortcut to approving an Entry by its ID
func (e *EntryAgent) ApproveEntryByID(ctx context.Context, id string) (Entry, error) {
	// ensure basic auth has been provided and matches admin credentials
	if !IsBasicAuthSuccessful(ctx) {
		return Entry{}, UnauthorizedError{}
	}

	// retrieve entry
	entries, err := e.er.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	if len(entries) != 1 {
		return Entry{}, InternalError{fmt.Errorf("entries count other than 1: %d", len(entries))}
	}

	entry := entries[0]

	realm := RealmFromContext(ctx)

	// ensure that Entry realm matches current realm
	if realm.Config.Name != entry.RealmName {
		return Entry{}, ConflictError{errors.New("invalid realm")}
	}

	// check Entry status
	switch entry.Status {
	case EntryStatusPaid:
		// all good
	default:
		return Entry{}, ConflictError{fmt.Errorf(
			"entry can only be approved if status is pending or ready: status is %s",
			entry.Status,
		)}
	}

	// check if Entry has already been approved
	if entry.ApprovedAt != nil {
		return Entry{}, ConflictError{errors.New("entry has already been approved")}
	}

	ts := e.cl.Now()
	entry.ApprovedAt = &ts

	// write to database
	if err := e.er.Update(ctx, &entry); err != nil {
		return Entry{}, domainErrorFromRepositoryError(err)
	}

	return entry, nil
}

// RetrieveEntryPredictionByTimestamp returns the entry prediction affiliated with the provided entry id that is valid at the point the provided timestamp occurs
func (e *EntryAgent) RetrieveEntryPredictionByTimestamp(ctx context.Context, entry Entry, ts time.Time) (EntryPrediction, error) {
	// retrieve entry prediction
	entryPrediction, err := e.epr.SelectByEntryIDAndTimestamp(ctx, entry.ID.String(), ts)
	if err != nil {
		return EntryPrediction{}, domainErrorFromRepositoryError(err)
	}

	return entryPrediction, nil
}

// RetrieveEntryPredictionsForActiveSeasonByTimestamp retrieves all entry predictions active at the provided timestamp
// for the provided active season
func (e *EntryAgent) RetrieveEntryPredictionsForActiveSeasonByTimestamp(
	ctx context.Context,
	season Season,
	ts time.Time,
) ([]EntryPrediction, error) {
	// ensure that season is active based on provided timestamp
	if !season.Live.HasBegunBy(ts) || season.Live.HasElapsedBy(ts) {
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

// GetPredictionRankingLimit returns any limits that apply to the provided entry ID based on the current clock timestamp.
//
// The Prediction Ranking Limit determines how many teams can change positions in a new Entry Prediction
// that is made by the provided entry ID.
// A return value of -1 indicates no limit, otherwise the return value represents the number of teams allowed to change (including 0).
func (e *EntryAgent) GetPredictionRankingLimit(ctx context.Context, entry Entry) (int, error) {
	st, err := e.sr.SelectLatestBySeasonIDAndTimestamp(ctx, entry.SeasonID, e.cl.Now())
	if err != nil {
		if errors.As(err, &MissingDBRecordError{}) {
			// no standings exist for season - i.e. round 1 has not yet begun
			// so return no ranking limit
			return RankingLimitNone, nil
		}

		return 0, fmt.Errorf("cannot retrieve standings by season and timestamp: %w", err)
	}

	ep, err := e.epr.SelectByEntryIDAndTimestamp(ctx, entry.ID.String(), e.cl.Now())
	if err != nil {
		if errors.As(err, &MissingDBRecordError{}) {
			// entry has not yet made a first entry prediction
			// so return no ranking limit
			return RankingLimitNone, nil
		}

		return 0, fmt.Errorf("cannot retrieve entry prediction by entry and timestamp: %w", err)
	}

	if !ep.CreatedAt.Before(st.CreatedAt) {
		// a new entry prediction has already been created since the most recent standings were created (round started)
		// so set a limit of 0
		return 0, nil
	}

	// latest entry prediction was created prior to the current standings (round)
	// so allow a regular limit
	return RankingLimitRegular, nil
}

// CheckRankingLimit returns an error if the cahgnes entailed by provided EntryPrediction exceeds the provided limit
func (e *EntryAgent) CheckRankingLimit(ctx context.Context, limit int, newEP EntryPrediction, entry Entry) error {
	if limit == RankingLimitNone {
		// no limit
		return nil
	}

	// retrieve most recent existing entry prediction
	latestEP, err := e.RetrieveEntryPredictionByTimestamp(ctx, entry, e.cl.Now())
	if err != nil {
		if errors.As(err, &NotFoundError{}) {
			// no existing latest entry prediction, so nothing to check limit against
			return nil
		}
		return fmt.Errorf("cannot retrieve entry prediction by timestamp: %w", err)
	}

	ids := GetChangedRankingIDs(newEP.Rankings, latestEP.Rankings)
	if len(ids) == 0 {
		return ConflictError{errors.New("no changes to rankings")}
	}
	if len(ids) > limit {
		return ConflictError{fmt.Errorf("cannot change %d rankings (%d permitted)", len(ids), limit)}
	}

	return nil
}

func (e *EntryAgent) SeedEntries(ctx context.Context, entries []Entry) error {
	for _, entry := range entries {
		if err := e.er.Insert(ctx, &entry); err != nil {
			switch err.(type) {
			case DuplicateDBRecordError:
				// already seeded, so we can fail silently
				continue
			}
			return fmt.Errorf("cannot seed entry: %w", err)
		}

		for _, ep := range entry.EntryPredictions {
			if err := e.epr.Insert(context.Background(), &ep); err != nil {
				switch err.(type) {
				case DuplicateDBRecordError:
					// already seeded, so we can fail silently
					continue
				}
				return fmt.Errorf("cannot seed entry prediction: %w", err)
			}
		}
	}

	return nil
}

// NewEntryAgent returns a new EntryAgent using the provided repositories
func NewEntryAgent(er EntryRepository, epr EntryPredictionRepository, sr StandingsRepository, sc SeasonCollection, cl Clock) (*EntryAgent, error) {
	switch {
	case er == nil:
		return nil, fmt.Errorf("entry repository: %w", ErrIsNil)
	case epr == nil:
		return nil, fmt.Errorf("entry prediction repository: %w", ErrIsNil)
	case sr == nil:
		return nil, fmt.Errorf("standings repository: %w", ErrIsNil)
	case sc == nil:
		return nil, fmt.Errorf("season collection: %w", ErrIsNil)
	case cl == nil:
		return nil, fmt.Errorf("clock: %w", ErrIsNil)
	}

	return &EntryAgent{er, epr, sr, sc, cl}, nil
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
	entry.SeasonID = strings.Trim(entry.SeasonID, " ")
	entry.RealmName = strings.Trim(entry.RealmName, " ")
	entry.EntrantName = strings.Trim(entry.EntrantName, " ")
	entry.EntrantNickname = strings.Trim(entry.EntrantNickname, " ")
	entry.EntrantEmail = strings.Trim(entry.EntrantEmail, " ")
	entry.Status = strings.Trim(entry.Status, " ")
	trimStringPtr(entry.PaymentMethod)
	trimStringPtr(entry.PaymentRef)

	var validationMsgs []string

	// validate
	for k, v := range map[string]string{
		"ID":         entry.ID.String(),
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
	if entry.PaymentMethod != nil && !isValidEntryPaymentMethod(*entry.PaymentMethod) {
		validationMsgs = append(validationMsgs, fmt.Sprintf("%s is not a valid payment method", *entry.PaymentMethod))
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	return nil
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
	case EntryStatusPending, EntryStatusPaid:
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

func trimStringPtr(str *string) {
	if str == nil {
		return
	}
	*str = strings.Trim(*str, " ")
}
