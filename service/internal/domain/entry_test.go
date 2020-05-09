package domain_test

import (
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"reflect"
	"testing"
	"time"
)

func TestEntryAgent_CreateEntry(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	ctx := domain.NewContext()
	ctx.Realm.Name = "TEST_REALM"
	ctx.Realm.PIN = "5678"
	ctx.Guard.SetAttempt("5678")

	season := models.Season{
		ID:          "199293_1",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	}

	paymentMethod := "entry_payment_method"
	paymentRef := "entry_payment_ref"

	entry := models.Entry{
		// these values should be populated
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",

		// these values should be overridden
		ID:             uuid.Must(uuid.NewV4()),
		ShortCode:      "entry_short_code",
		SeasonID:       "entry_season_id",
		RealmName:      "entry_realm_name",
		Status:         "entry_status",
		PaymentMethod:  sqltypes.ToNullString(&paymentMethod),
		PaymentRef:     sqltypes.ToNullString(&paymentRef),
		TeamIDSequence: []string{"entry_team_id_1", "entry_team_id_2"},
		ApprovedAt:     sqltypes.ToNullTime(time.Now()),
		CreatedAt:      time.Time{},
		UpdatedAt:      sqltypes.ToNullTime(time.Now()),
	}

	t.Run("create a valid entry with a valid guard value must succeed", func(t *testing.T) {
		// should succeed
		createdEntry, err := agent.CreateEntry(ctx, entry, &season)
		if err != nil {
			t.Fatal(err)
		}

		// check raw values that shouldn't have changed
		if entry.EntrantName != createdEntry.EntrantName {
			expectedGot(t, entry.EntrantName, createdEntry.EntrantName)
		}
		if entry.EntrantNickname != createdEntry.EntrantNickname {
			expectedGot(t, entry.EntrantName, createdEntry.EntrantName)
		}
		if entry.EntrantEmail != createdEntry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, createdEntry.EntrantEmail)
		}

		// check sanitised values
		expectedSeasonID := season.ID
		expectedRealm := ctx.Realm.Name
		expectedStatus := models.EntryStatusPending

		if createdEntry.ID.String() == "" {
			expectedNonEmpty(t, "Entry.ID")
		}
		if createdEntry.ShortCode == "" {
			expectedNonEmpty(t, "Entry.ShortCode")
		}
		if expectedSeasonID != createdEntry.SeasonID {
			expectedGot(t, expectedSeasonID, createdEntry.SeasonID)
		}
		if expectedRealm != createdEntry.RealmName {
			expectedGot(t, expectedRealm, createdEntry.RealmName)
		}
		if expectedStatus != createdEntry.Status {
			expectedGot(t, expectedStatus, createdEntry.Status)
		}
		if createdEntry.PaymentMethod.Valid {
			expectedEmpty(t, "Entry.PaymentMethod", createdEntry.PaymentMethod)
		}
		if createdEntry.PaymentRef.Valid {
			expectedEmpty(t, "Entry.PaymentRef", createdEntry.PaymentRef)
		}
		if !reflect.DeepEqual([]string{}, createdEntry.TeamIDSequence) {
			expectedEmpty(t, "Entry.TeamIDSequence", createdEntry.TeamIDSequence)
		}
		if createdEntry.ApprovedAt.Valid {
			expectedEmpty(t, "Entry.ApprovedAt", createdEntry.ApprovedAt)
		}
		if createdEntry.CreatedAt.Equal(time.Time{}) {
			expectedNonEmpty(t, "Entry.CreatedAt")
		}
		if createdEntry.UpdatedAt.Valid {
			expectedEmpty(t, "Entry.UpdatedAt", createdEntry.UpdatedAt)
		}

		// inserting same entry a second time should fail
		_, err = agent.CreateEntry(ctx, entry, &season)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create an entry with a nil season pointer must fail", func(t *testing.T) {
		_, err := agent.CreateEntry(ctx, entry, nil)
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("create an entry with an invalid guard value must fail", func(t *testing.T) {
		ctxWithInvalidGuardValue := ctx
		ctxWithInvalidGuardValue.Guard.SetAttempt("not_the_correct_realm_pin")
		_, err := agent.CreateEntry(ctxWithInvalidGuardValue, entry, &season)
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("create an entry for a season that isn't accepting entries must fail", func(t *testing.T) {
		seasonNotAcceptingEntries := season

		// entry window doesn't begin until tomorrow
		seasonNotAcceptingEntries.EntriesFrom = time.Now().Add(24 * time.Hour)
		seasonNotAcceptingEntries.StartDate = time.Now().Add(48 * time.Hour)

		_, err := agent.CreateEntry(ctx, entry, &seasonNotAcceptingEntries)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}

		// entry window has already elapsed
		seasonNotAcceptingEntries.EntriesFrom = time.Now().Add(-48 * time.Hour)
		seasonNotAcceptingEntries.StartDate = time.Now().Add(-24 * time.Hour)

		_, err = agent.CreateEntry(ctx, entry, &seasonNotAcceptingEntries)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create an entry with missing required fields must fail", func(t *testing.T) {
		var invalidEntry models.Entry
		var err error

		// missing entrant name
		invalidEntry = entry
		invalidEntry.EntrantName = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant nickname
		invalidEntry = entry
		invalidEntry.EntrantNickname = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant email
		invalidEntry = entry
		invalidEntry.EntrantEmail = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid entrant email
		invalidEntry = entry
		invalidEntry.EntrantEmail = "not_a_valid_email@"
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("create an entry with invalid nickname must fail", func(t *testing.T) {
		invalidEntry := entry
		invalidEntry.EntrantEmail = "harry.redknapp.alternative.email@football.net"

		invalidEntry.EntrantNickname = "!@£$" // contains non-alphanumeric characters
		_, err := agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		invalidEntry.EntrantNickname = "1234567890123456789" // longer than 12 characters
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("create an entry with existing entrant data must fail", func(t *testing.T) {
		invalidEntry := entry
		invalidEntry.EntrantName = "Not Harry Redknapp"
		// nickname doesn't change so it already exists
		invalidEntry.EntrantEmail = "not.harry.redknapp@football.net"
		_, err := agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}

		invalidEntry = entry
		invalidEntry.EntrantName = "Not Harry Redknapp"
		invalidEntry.EntrantNickname = "NotHarryR"
		// email doesn't change so it already exists
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_RetrieveEntry(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	ctx := domain.NewContext()
	ctx.Realm.Name = "TEST_REALM"
	ctx.Realm.PIN = "5678"
	ctx.Guard.SetAttempt("5678")

	// seed initial entry
	entry, err := agent.CreateEntry(ctx, models.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",
	}, &models.Season{
		ID:          "12345",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve an existent entry with valid credentials must succeed", func(t *testing.T) {
		// should succeed
		retrievedEntry, err := agent.RetrieveEntryByID(ctx, entry.ID.String())
		if err != nil {
			t.Fatal(err)
		}

		// check values
		if entry.ID != retrievedEntry.ID {
			expectedGot(t, entry.ID, retrievedEntry.ID)
		}
		if entry.ShortCode != retrievedEntry.ShortCode {
			expectedGot(t, entry.ShortCode, retrievedEntry.ShortCode)
		}
		if entry.SeasonID != retrievedEntry.SeasonID {
			expectedGot(t, entry.SeasonID, retrievedEntry.SeasonID)
		}
		if entry.RealmName != retrievedEntry.RealmName {
			expectedGot(t, entry.RealmName, retrievedEntry.RealmName)
		}
		if entry.EntrantName != retrievedEntry.EntrantName {
			expectedGot(t, entry.EntrantName, retrievedEntry.EntrantName)
		}
		if entry.EntrantNickname != retrievedEntry.EntrantNickname {
			expectedGot(t, entry.EntrantNickname, retrievedEntry.EntrantNickname)
		}
		if entry.EntrantEmail != retrievedEntry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, retrievedEntry.EntrantEmail)
		}
		if entry.Status != retrievedEntry.Status {
			expectedGot(t, entry.Status, retrievedEntry.Status)
		}
		if entry.PaymentMethod != retrievedEntry.PaymentMethod {
			expectedGot(t, entry.PaymentMethod, retrievedEntry.PaymentMethod)
		}
		if entry.PaymentRef != retrievedEntry.PaymentRef {
			expectedGot(t, entry.PaymentRef, retrievedEntry.PaymentRef)
		}
		if !reflect.DeepEqual(entry.TeamIDSequence, retrievedEntry.TeamIDSequence) {
			expectedGot(t, entry.PaymentRef, retrievedEntry.PaymentRef)
		}
		if entry.ApprovedAt.Time.In(domain.Locations["UTC"]) != retrievedEntry.ApprovedAt.Time.In(domain.Locations["UTC"]) {
			expectedGot(t, entry.ApprovedAt, retrievedEntry.ApprovedAt)
		}
		if entry.CreatedAt.In(domain.Locations["UTC"]) != retrievedEntry.CreatedAt.In(domain.Locations["UTC"]) {
			expectedGot(t, entry.CreatedAt, retrievedEntry.CreatedAt)
		}
		if entry.UpdatedAt.Time.In(domain.Locations["UTC"]) != retrievedEntry.UpdatedAt.Time.In(domain.Locations["UTC"]) {
			expectedGot(t, entry.UpdatedAt, retrievedEntry.UpdatedAt)
		}
	})

	t.Run("retrieve a non-existent entry must fail", func(t *testing.T) {
		_, err = agent.RetrieveEntryByID(ctx, "not_a_valid_id")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve an entry with a mismatched realm must fail", func(t *testing.T) {
		ctxWithInvalidRealm := ctx
		ctxWithInvalidRealm.Realm.Name = "DIFFERENT_REALM"

		_, err = agent.RetrieveEntryByID(ctxWithInvalidRealm, entry.ID.String())
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_UpdateEntry(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	ctx := domain.NewContext()
	ctx.Realm.Name = "TEST_REALM"
	ctx.Realm.PIN = "5678"
	ctx.Guard.SetAttempt("5678")

	// seed initial entry
	entry, err := agent.CreateEntry(ctx, models.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",
	}, &models.Season{
		ID:          "12345",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("update an existent entry with a valid alternative entry must succeed", func(t *testing.T) {
		// define changed entry values
		changedEntryPaymentRef := "changed_entry_payment_ref"

		changedEntry := models.Entry{
			ID:              entry.ID,
			ShortCode:       "changed_entry_short_code",
			SeasonID:        "67890",
			RealmName:       entry.RealmName,
			EntrantName:     "Jamie Redknapp",
			EntrantNickname: "MrJamieR",
			EntrantEmail:    "jamie.redknapp@football.net",
			Status:          models.EntryStatusReady,
			PaymentRef:      sqltypes.ToNullString(&changedEntryPaymentRef),
			TeamIDSequence:  []string{"changed_team_id_1", "changed_team_id_2", "changed_team_id_3"},
			CreatedAt:       entry.CreatedAt,
		}

		// should succeed
		updatedEntry, err := agent.UpdateEntry(ctx, changedEntry)
		if err != nil {
			t.Fatal(err)
		}

		// check values that shouldn't have changed
		if entry.ID != updatedEntry.ID {
			expectedGot(t, entry.ID, updatedEntry.ID)
		}
		if entry.RealmName != updatedEntry.RealmName {
			expectedGot(t, entry.RealmName, updatedEntry.RealmName)
		}
		if entry.CreatedAt.In(domain.Locations["UTC"]) != updatedEntry.CreatedAt.In(domain.Locations["UTC"]) {
			expectedGot(t, entry.CreatedAt, updatedEntry.CreatedAt)
		}

		// check values that should have changed
		if changedEntry.ShortCode != updatedEntry.ShortCode {
			expectedGot(t, changedEntry.ShortCode, updatedEntry.ShortCode)
		}
		if changedEntry.SeasonID != updatedEntry.SeasonID {
			expectedGot(t, changedEntry.SeasonID, updatedEntry.SeasonID)
		}
		if changedEntry.EntrantName != updatedEntry.EntrantName {
			expectedGot(t, changedEntry.EntrantName, updatedEntry.EntrantName)
		}
		if changedEntry.EntrantNickname != updatedEntry.EntrantNickname {
			expectedGot(t, changedEntry.EntrantNickname, updatedEntry.EntrantNickname)
		}
		if changedEntry.EntrantEmail != updatedEntry.EntrantEmail {
			expectedGot(t, changedEntry.EntrantEmail, updatedEntry.EntrantEmail)
		}
		if changedEntry.Status != updatedEntry.Status {
			expectedGot(t, changedEntry.Status, updatedEntry.Status)
		}
		if changedEntry.PaymentRef != updatedEntry.PaymentRef {
			expectedGot(t, changedEntry.PaymentRef, updatedEntry.PaymentRef)
		}
		if !reflect.DeepEqual(changedEntry.TeamIDSequence, updatedEntry.TeamIDSequence) {
			expectedGot(t, changedEntry.PaymentRef, updatedEntry.PaymentRef)
		}
		if !updatedEntry.UpdatedAt.Valid {
			expectedNonEmpty(t, "Entry.UpdatedAt")
		}
	})

	t.Run("update an existent entry with a changed realm must fail", func(t *testing.T) {
		_, err = agent.UpdateEntry(ctx, models.Entry{ID: entry.ID, RealmName: "NOT_THE_ORIGINAL_REALM"})
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update a non-existent entry must fail", func(t *testing.T) {
		_, err = agent.UpdateEntry(ctx, models.Entry{ID: uuid.Must(uuid.NewV4()), RealmName: entry.RealmName})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("update an existing entry with missing required fields must fail", func(t *testing.T) {
		var invalidEntry models.Entry
		var err error

		// missing entrant name
		invalidEntry = entry
		invalidEntry.EntrantName = ""
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant nickname
		invalidEntry = entry
		invalidEntry.EntrantNickname = ""
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant email
		invalidEntry = entry
		invalidEntry.EntrantEmail = ""
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid entrant email
		invalidEntry = entry
		invalidEntry.EntrantEmail = "not_a_valid_email@"
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("update an existing entry with invalid fields must fail", func(t *testing.T) {
		var invalidEntry models.Entry
		var err error

		// invalid nickname (non-alphanumeric characters)
		invalidEntry = entry
		invalidEntry.EntrantNickname = "!@£$%"
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid nickname (more than 12 characters)
		invalidEntry = entry
		invalidEntry.EntrantNickname = "1234567890123456789"
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid email
		invalidEntry = entry
		invalidEntry.EntrantEmail = "not_a_valid_email"
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid status
		invalidEntry = entry
		invalidEntry.Status = "not_a_valid_status"
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid payment method
		invalidEntry = entry
		invalidPaymentMethod := "not_a_valid_payment_method"
		invalidEntry.PaymentMethod = sqltypes.ToNullString(&invalidPaymentMethod)
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})
}

func TestEntryAgent_UpdateEntryPaymentDetails(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	ctxWithPIN := domain.NewContext()
	ctxWithPIN.Realm.Name = "TEST_REALM"
	ctxWithPIN.Realm.PIN = "5678"
	ctxWithPIN.Guard.SetAttempt("5678")

	// seed initial entry
	entry := models.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",
	}
	season := models.Season{
		ID:          "12345",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	}
	entry, err := agent.CreateEntry(ctxWithPIN, entry, &season)
	if err != nil {
		t.Fatal(err)
	}

	// override guard value so that context can be re-used for UpdateEntryPaymentDetails
	ctx := ctxWithPIN
	ctx.Guard.SetAttempt(entry.ID.String())

	paymentRef := "ABCD1234"

	t.Run("update payment details for an existent entry with valid credentials must succeed", func(t *testing.T) {
		entryWithPaymentDetails, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			models.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if err != nil {
			t.Fatal(err)
		}

		if models.EntryPaymentMethodPayPal != entryWithPaymentDetails.PaymentMethod.String {
			expectedGot(t, models.EntryPaymentMethodPayPal, entryWithPaymentDetails.PaymentMethod.String)
		}

		if paymentRef != entryWithPaymentDetails.PaymentRef.String {
			expectedGot(t, paymentRef, entryWithPaymentDetails.PaymentRef.String)
		}
	})

	t.Run("update invalid payment method for an existent entry must fail", func(t *testing.T) {
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			"not_a_valid_payment_method",
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("update missing payment ref for an existent entry must fail", func(t *testing.T) {
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			models.EntryPaymentMethodPayPal,
			"",
		)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("update payment details for a non-existent entry must fail", func(t *testing.T) {
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			"not_an_existing_entry_id",
			models.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid realm must fail", func(t *testing.T) {
		ctxWithMismatchedRealm := ctx
		ctxWithMismatchedRealm.Realm.Name = "DIFFERENT_REALM"
		_, err := agent.UpdateEntryPaymentDetails(
			ctxWithMismatchedRealm,
			entry.ID.String(),
			models.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid lookup ref must fail", func(t *testing.T) {
		ctxWithMismatchedGuardValue := ctx
		ctxWithMismatchedGuardValue.Guard.SetAttempt("not_the_correct_entry_short_code")
		_, err := agent.UpdateEntryPaymentDetails(
			ctxWithMismatchedGuardValue,
			entry.ID.String(),
			models.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid status must fail", func(t *testing.T) {
		// seed an additional entry
		entryWithInvalidStatus := models.Entry{
			EntrantName:     "Jamie Redknapp",
			EntrantNickname: "MrJamieR",
			EntrantEmail:    "jamie.redknapp@football.net",
		}
		entryWithInvalidStatus, err := agent.CreateEntry(ctxWithPIN, entryWithInvalidStatus, &season)
		if err != nil {
			t.Fatal(err)
		}

		// change its status so it can no longer accept payments and re-save
		entryWithInvalidStatus.Status = models.EntryStatusPaid
		entryWithInvalidStatus, err = agent.UpdateEntry(ctxWithPIN, entryWithInvalidStatus)
		if err != nil {
			t.Fatal(err)
		}

		// now running the operation we're testing should fail
		_, err = agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			models.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_ApproveEntryByShortCode(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	ctxWithPIN := domain.NewContext()
	ctxWithPIN.Realm.Name = "TEST_REALM"
	ctxWithPIN.Realm.PIN = "5678"
	ctxWithPIN.Guard.SetAttempt("5678")

	// seed initial entries
	season := models.Season{
		ID:          "12345",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	}

	entry, err := agent.CreateEntry(ctxWithPIN, models.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",
	}, &season)
	if err != nil {
		t.Fatal(err)
	}

	entryWithPaidStatus, err := agent.CreateEntry(ctxWithPIN, models.Entry{
		EntrantName:     "Jamie Redknapp",
		EntrantNickname: "MrJamieR",
		EntrantEmail:    "jamie.redknapp@football.net",
	}, &season)
	if err != nil {
		t.Fatal(err)
	}
	entryWithPaidStatus.Status = models.EntryStatusPaid
	entryWithPaidStatus, err = agent.UpdateEntry(ctxWithPIN, entryWithPaidStatus)
	if err != nil {
		t.Fatal(err)
	}

	entryWithReadyStatus, err := agent.CreateEntry(ctxWithPIN, models.Entry{
		EntrantName:     "Frank Lampard",
		EntrantNickname: "FrankieLamps",
		EntrantEmail:    "frank.lampard@football.net",
	}, &season)
	if err != nil {
		t.Fatal(err)
	}
	entryWithReadyStatus.Status = models.EntryStatusReady
	entryWithReadyStatus, err = agent.UpdateEntry(ctxWithPIN, entryWithReadyStatus)
	if err != nil {
		t.Fatal(err)
	}

	// set basic auth successful value to true so that it can be used by ApproveEntryByShortCode
	ctx := ctxWithPIN
	ctx.BasicAuthSuccessful = true

	t.Run("approve existent entry short code with valid credentials must succeed", func(t *testing.T) {
		// attempt to approve entry with paid status
		approvedEntry, err := agent.ApproveEntryByShortCode(ctx, entryWithPaidStatus.ShortCode)
		if err != nil {
			t.Fatal(err)
		}
		if !approvedEntry.IsApproved() {
			expectedGot(t, "approved entry true", "approved entry false")
		}
		if !approvedEntry.ApprovedAt.Valid {
			expectedNonEmpty(t, "Entry.ApprovedAt")
		}

		// attempt to approve entry with ready status
		approvedEntry, err = agent.ApproveEntryByShortCode(ctx, entryWithReadyStatus.ShortCode)
		if err != nil {
			t.Fatal(err)
		}
		if !approvedEntry.IsApproved() {
			expectedGot(t, "approved entry true", "approved entry false")
		}
		if !approvedEntry.ApprovedAt.Valid {
			expectedNonEmpty(t, "Entry.ApprovedAt")
		}
	})

	t.Run("approve existent entry with invalid credentials must fail", func(t *testing.T) {
		ctxWithInvalidCredentials := ctx
		ctxWithInvalidCredentials.BasicAuthSuccessful = false
		_, err := agent.ApproveEntryByShortCode(ctxWithInvalidCredentials, entry.ShortCode)
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("approve non-existent entry with valid credentials must fail", func(t *testing.T) {
		_, err := agent.ApproveEntryByShortCode(ctx, "non_existent_short_code")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("approve existent entry with invalid realm must fail", func(t *testing.T) {
		ctxWithInvalidRealm := ctx
		ctxWithInvalidRealm.Realm.Name = "DIFFERENT_REALM"
		_, err := agent.ApproveEntryByShortCode(ctxWithInvalidRealm, entry.ShortCode)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("approve existent entry with pending status must fail", func(t *testing.T) {
		// initial entry object should still have default "pending" status so just attempt to approve this
		_, err := agent.ApproveEntryByShortCode(ctx, entry.ShortCode)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("approve existent entry that has already been approved must fail", func(t *testing.T) {
		// just try to approve the same entry again
		_, err := agent.ApproveEntryByShortCode(ctx, entryWithPaidStatus.ShortCode)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}
