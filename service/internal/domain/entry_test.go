package domain_test

import (
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
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

	season := domain.Season{
		ID:          "199293_1",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	}

	paymentMethod := "entry_payment_method"
	paymentRef := "entry_payment_ref"

	entry := domain.Entry{
		// these values should be populated
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "Mr Harry R",
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
		if !cmp.Equal(entry.EntrantName, createdEntry.EntrantName)().Success() {
			expectedGot(t, entry.EntrantName, createdEntry.EntrantName)
		}
		if !cmp.Equal(entry.EntrantNickname, createdEntry.EntrantNickname)().Success() {
			expectedGot(t, entry.EntrantName, createdEntry.EntrantName)
		}
		if !cmp.Equal(entry.EntrantEmail, createdEntry.EntrantEmail)().Success() {
			expectedGot(t, entry.EntrantEmail, createdEntry.EntrantEmail)
		}

		// check sanitised values
		expectedSeasonID := season.ID
		expectedRealm := ctx.Realm.Name
		expectedStatus := domain.EntryStatusPending

		if cmp.Equal("", createdEntry.ID)().Success() {
			expectedNonEmpty(t, "Entry.ID")
		}
		if cmp.Equal("", createdEntry.ShortCode)().Success() {
			expectedNonEmpty(t, "Entry.ShortCode")
		}
		if !cmp.Equal(expectedSeasonID, createdEntry.SeasonID)().Success() {
			expectedGot(t, expectedSeasonID, createdEntry.SeasonID)
		}
		if !cmp.Equal(expectedRealm, createdEntry.RealmName)().Success() {
			expectedGot(t, expectedRealm, createdEntry.RealmName)
		}
		if !cmp.Equal(expectedStatus, createdEntry.Status)().Success() {
			expectedGot(t, expectedStatus, createdEntry.Status)
		}
		if !cmp.Equal(sqltypes.NullString{}, createdEntry.PaymentMethod)().Success() {
			expectedEmpty(t, "Entry.PaymentMethod", createdEntry.PaymentMethod)
		}
		if !cmp.Equal(sqltypes.NullString{}, createdEntry.PaymentRef)().Success() {
			expectedEmpty(t, "Entry.PaymentRef", createdEntry.PaymentRef)
		}
		if !cmp.DeepEqual([]string{}, createdEntry.TeamIDSequence)().Success() {
			expectedEmpty(t, "Entry.TeamIDSequence", createdEntry.TeamIDSequence)
		}
		if !cmp.Equal(sqltypes.NullTime{}, createdEntry.ApprovedAt)().Success() {
			expectedEmpty(t, "Entry.ApprovedAt", createdEntry.ApprovedAt)
		}
		if cmp.Equal(time.Time{}, createdEntry.CreatedAt)().Success() {
			expectedNonEmpty(t, "Entry.CreatedAt")
		}
		if !cmp.Equal(sqltypes.NullTime{}, createdEntry.UpdatedAt)().Success() {
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
		var invalidEntry domain.Entry
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
		invalidEntry.EntrantNickname = "Not Harry R"
		// email doesn't change so it already exists
		_, err = agent.CreateEntry(ctx, invalidEntry, &season)
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
	entry, err := agent.CreateEntry(ctx, domain.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "Mr Harry R",
		EntrantEmail:    "harry.redknapp@football.net",
	}, &domain.Season{
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

		changedEntry := domain.Entry{
			ID:              entry.ID,
			ShortCode:       "changed_entry_short_code",
			SeasonID:        "67890",
			RealmName:       entry.RealmName,
			EntrantName:     "Jamie Redknapp",
			EntrantNickname: "Mr Jamie R",
			EntrantEmail:    "jamie.redknapp@football.net",
			Status:          domain.EntryStatusReady,
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
		if !cmp.Equal(entry.ID, updatedEntry.ID)().Success() {
			expectedGot(t, entry.ID, updatedEntry.ID)
		}
		if !cmp.Equal(entry.RealmName, updatedEntry.RealmName)().Success() {
			expectedGot(t, entry.RealmName, updatedEntry.RealmName)
		}
		if !cmp.Equal(entry.CreatedAt, updatedEntry.CreatedAt)().Success() {
			expectedGot(t, entry.CreatedAt, updatedEntry.CreatedAt)
		}

		// check values that should have changed
		if !cmp.Equal(changedEntry.ShortCode, updatedEntry.ShortCode)().Success() {
			expectedGot(t, changedEntry.ShortCode, updatedEntry.ShortCode)
		}
		if !cmp.Equal(changedEntry.SeasonID, updatedEntry.SeasonID)().Success() {
			expectedGot(t, changedEntry.SeasonID, updatedEntry.SeasonID)
		}
		if !cmp.Equal(changedEntry.EntrantName, updatedEntry.EntrantName)().Success() {
			expectedGot(t, changedEntry.EntrantName, updatedEntry.EntrantName)
		}
		if !cmp.Equal(changedEntry.EntrantNickname, updatedEntry.EntrantNickname)().Success() {
			expectedGot(t, changedEntry.EntrantNickname, updatedEntry.EntrantNickname)
		}
		if !cmp.Equal(changedEntry.EntrantEmail, updatedEntry.EntrantEmail)().Success() {
			expectedGot(t, changedEntry.EntrantEmail, updatedEntry.EntrantEmail)
		}
		if !cmp.Equal(changedEntry.Status, updatedEntry.Status)().Success() {
			expectedGot(t, changedEntry.Status, updatedEntry.Status)
		}
		if !cmp.Equal(changedEntry.PaymentRef, updatedEntry.PaymentRef)().Success() {
			expectedGot(t, changedEntry.PaymentRef, updatedEntry.PaymentRef)
		}
		if !reflect.DeepEqual(changedEntry.TeamIDSequence, updatedEntry.TeamIDSequence) {
			expectedGot(t, changedEntry.PaymentRef, updatedEntry.PaymentRef)
		}
		if cmp.Equal(time.Time{}, updatedEntry.UpdatedAt)().Success() {
			expectedNonEmpty(t, "Entry.UpdatedAt")
		}
	})

	t.Run("update an existent entry with a changed realm must fail", func(t *testing.T) {
		_, err = agent.UpdateEntry(ctx, domain.Entry{ID: entry.ID, RealmName: "NOT_THE_ORIGINAL_REALM"})
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update a non-existent entry must fail", func(t *testing.T) {
		_, err = agent.UpdateEntry(ctx, domain.Entry{ID: uuid.Must(uuid.NewV4()), RealmName: entry.RealmName})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("update an existing entry with missing required fields must fail", func(t *testing.T) {
		var invalidEntry domain.Entry
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
		var invalidEntry domain.Entry
		var err error

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
	entry := domain.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "Mr Harry R",
		EntrantEmail:    "harry.redknapp@football.net",
	}
	season := domain.Season{
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
	ctx.Guard.SetAttempt(entry.ShortCode)

	paymentRef := "ABCD1234"

	t.Run("update payment details for an existent entry with valid credentials must succeed", func(t *testing.T) {
		entryWithPaymentDetails, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(domain.EntryPaymentMethodPayPal, entryWithPaymentDetails.PaymentMethod.String)().Success() {
			expectedGot(t, domain.EntryPaymentMethodPayPal, entryWithPaymentDetails.PaymentMethod.String)
		}

		if !cmp.Equal(paymentRef, entryWithPaymentDetails.PaymentRef.String)().Success() {
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
			domain.EntryPaymentMethodPayPal,
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
			domain.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid realm must fail", func(t *testing.T) {
		ctxWithMismatchedtRealm := ctx
		ctxWithMismatchedtRealm.Realm.Name = "DIFFERENT_REALM"
		_, err := agent.UpdateEntryPaymentDetails(
			ctxWithMismatchedtRealm,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
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
			domain.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid status must fail", func(t *testing.T) {
		// seed an additional entry
		entryWithInvalidStatus := domain.Entry{
			EntrantName:     "Jamie Redknapp",
			EntrantNickname: "Mr Jamie R",
			EntrantEmail:    "jamie.redknapp@football.net",
		}
		entryWithInvalidStatus, err := agent.CreateEntry(ctxWithPIN, entryWithInvalidStatus, &season)
		if err != nil {
			t.Fatal(err)
		}

		// change its status so it can no longer accept payments and re-save
		entryWithInvalidStatus.Status = domain.EntryStatusPaid
		entryWithInvalidStatus, err = agent.UpdateEntry(ctxWithPIN, entryWithInvalidStatus)
		if err != nil {
			t.Fatal(err)
		}

		// now running the operation we're testing should fail
		_, err = agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
			paymentRef,
		)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}
