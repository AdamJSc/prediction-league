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
	ctx.SetRealm("TEST_REALM")
	ctx.SetRealmPIN("5678")

	season := domain.Season{
		ID:          "199293_1",
		EntriesFrom: time.Now().Add(-24 * time.Hour),
		StartDate:   time.Now().Add(24 * time.Hour),
	}

	entry := domain.Entry{
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "Mr Harry R",
		EntrantEmail:    "harry.redknapp@football.net",
	}

	t.Run("create a valid entry with a valid PIN must succeed", func(t *testing.T) {
		// should succeed
		createdEntry, err := agent.CreateEntry(ctx, entry, &season, "5678")
		if err != nil {
			t.Fatal(err)
		}

		// check raw values that shouldn't have changed
		if !cmp.Equal(entry.EntrantName, createdEntry.EntrantName)().Success() {
			expectedGot(t, entry.EntrantName, createdEntry.EntrantName)
		}
		if !cmp.Equal(entry.EntrantEmail, createdEntry.EntrantEmail)().Success() {
			expectedGot(t, entry.EntrantEmail, createdEntry.EntrantEmail)
		}
		if !cmp.Equal(entry.PaymentRef, createdEntry.PaymentRef)().Success() {
			expectedGot(t, entry.PaymentRef, createdEntry.PaymentRef)
		}
		if !cmp.Equal(entry.UpdatedAt, createdEntry.UpdatedAt)().Success() {
			expectedGot(t, entry.UpdatedAt, createdEntry.UpdatedAt)
		}

		// check sanitised values
		expectedSeasonID := season.ID
		expectedRealm := ctx.GetRealm()
		expectedNickname := "MrHarryR"
		expectedStatus := domain.EntryStatusPending

		if cmp.Equal("", createdEntry.ID)().Success() {
			expectedNonEmpty(t, "Entry.ID")
		}
		if cmp.Equal("", createdEntry.LookupRef)().Success() {
			expectedNonEmpty(t, "Entry.LookupRef")
		}
		if !cmp.Equal(expectedSeasonID, createdEntry.SeasonID)().Success() {
			expectedGot(t, expectedSeasonID, createdEntry.SeasonID)
		}
		if !cmp.Equal(expectedRealm, createdEntry.Realm)().Success() {
			expectedGot(t, expectedRealm, createdEntry.Realm)
		}
		if !cmp.Equal(expectedNickname, createdEntry.EntrantNickname)().Success() {
			expectedGot(t, expectedNickname, createdEntry.EntrantNickname)
		}
		if !cmp.Equal(expectedStatus, createdEntry.Status)().Success() {
			expectedGot(t, expectedStatus, createdEntry.Status)
		}
		if !cmp.DeepEqual([]string{}, createdEntry.TeamIDSequence)().Success() {
			expectedEmpty(t, "Entry.TeamIDSequence", createdEntry.TeamIDSequence)
		}
		if cmp.Equal(time.Time{}, createdEntry.CreatedAt)().Success() {
			expectedNonEmpty(t, "Entry.CreatedAt")
		}
		if !cmp.Equal(sqltypes.NullTime{}, createdEntry.UpdatedAt)().Success() {
			expectedEmpty(t, "Entry.UpdatedAt", createdEntry.UpdatedAt)
		}

		// inserting same entry a second time should fail
		_, err = agent.CreateEntry(ctx, entry, &season, "5678")
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create an entry with a nil season pointer must fail", func(t *testing.T) {
		_, err := agent.CreateEntry(ctx, entry, nil, "5678")
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("create an entry with an invalid PIN must fail", func(t *testing.T) {
		_, err := agent.CreateEntry(ctx, entry, &season, "not_the_correct_realm_pin")
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("create an entry for a season that isn't accepting entries must fail", func(t *testing.T) {
		seasonNotAcceptingEntries := season

		// entry window doesn't begin until tomorrow
		seasonNotAcceptingEntries.EntriesFrom = time.Now().Add(24 * time.Hour)
		seasonNotAcceptingEntries.StartDate = time.Now().Add(48 * time.Hour)

		_, err := agent.CreateEntry(ctx, entry, &seasonNotAcceptingEntries, "5678")
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}

		// entry window has already elapsed
		seasonNotAcceptingEntries.EntriesFrom = time.Now().Add(-48 * time.Hour)
		seasonNotAcceptingEntries.StartDate = time.Now().Add(-24 * time.Hour)

		_, err = agent.CreateEntry(ctx, entry, &seasonNotAcceptingEntries, "5678")
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
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant nickname
		invalidEntry = entry
		invalidEntry.EntrantNickname = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant email
		invalidEntry = entry
		invalidEntry.EntrantEmail = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid entrant email
		invalidEntry = entry
		invalidEntry.EntrantEmail = "not_a_valid_email@"
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("create an entry with existing entrant data must fail", func(t *testing.T) {
		// don't change nickname so this already exists
		invalidEntry := entry
		invalidEntry.EntrantName = "Not Harry Redknapp"
		invalidEntry.EntrantEmail = "not.harry.redknapp@football.net"
		_, err := agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}

		// don't change email so this already exists
		invalidEntry = entry
		invalidEntry.EntrantName = "Not Harry Redknapp"
		invalidEntry.EntrantNickname = "Not Harry R"
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_UpdateEntry(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	ctx := domain.NewContext()
	ctx.SetRealm("TEST_REALM")

	entryUUID, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}

	entryPaymentRef := "initial_payment_ref"

	entry := domain.Entry{
		ID:              entryUUID,
		LookupRef:       "initial_lookup_ref",
		SeasonID:        "12345",
		Realm:           "TEST_REALM",
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",
		Status:          domain.EntryStatusPending,
		PaymentRef:      sqltypes.ToNullString(&entryPaymentRef),
		TeamIDSequence:  []string{"initial_team_id_1", "initial_team_id_2", "initial_team_id_3"},
	}

	err = domain.DBInsertEntry(db, &entry)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("update an existent entry with a valid alternative entry must succeed", func(t *testing.T) {
		// define changed entry values
		changedEntryPaymentRef := "changed_entry_payment_ref"

		changedEntry := domain.Entry{
			ID:              entry.ID,
			LookupRef:       "changed_entry_lookup_ref",
			SeasonID:        "67890",
			Realm:           entry.Realm,
			EntrantName:     "Jamie Redknapp",
			EntrantNickname: "MrJamieR",
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
		if !cmp.Equal(entry.Realm, updatedEntry.Realm)().Success() {
			expectedGot(t, entry.Realm, updatedEntry.Realm)
		}
		if !cmp.Equal(entry.CreatedAt, updatedEntry.CreatedAt)().Success() {
			expectedGot(t, entry.CreatedAt, updatedEntry.CreatedAt)
		}

		// check values that should have changed
		if !cmp.Equal(changedEntry.LookupRef, updatedEntry.LookupRef)().Success() {
			expectedGot(t, changedEntry.LookupRef, updatedEntry.LookupRef)
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
		entryID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.UpdateEntry(ctx, domain.Entry{ID: entryID, Realm: "NOT_THE_ORIGINAL_REALM"})
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update a non-existent entry must fail", func(t *testing.T) {
		entryID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.UpdateEntry(ctx, domain.Entry{ID: entryID, Realm: "TEST_REALM"})
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
}
