package domain_test

import (
	"github.com/LUSHDigital/core-sql/sqltypes"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
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

	t.Run("creating a valid entry with a valid PIN must succeed", func(t *testing.T) {
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
		if !cmp.DeepEqual(entry.TeamIDSequence, createdEntry.TeamIDSequence)().Success() {
			expectedGot(t, entry.TeamIDSequence, createdEntry.TeamIDSequence)
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

	t.Run("creating an entry with a nil pointer must fail", func(t *testing.T) {
		_, err := agent.CreateEntry(ctx, entry, nil, "5678")
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("creating an entry with an invalid PIN must fail", func(t *testing.T) {
		_, err := agent.CreateEntry(ctx, entry, &season, "not_the_correct_realm_pin")
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("creating an entry for a season that isn't accepting entries must fail", func(t *testing.T) {
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

	t.Run("creating an entry with missing required fields must fail", func(t *testing.T) {
		invalidEntry := entry

		// missing entrant name
		invalidEntry.EntrantName = ""
		_, err := agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant nickname
		invalidEntry.EntrantNickname = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entrant email
		invalidEntry.EntrantEmail = ""
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// invalid entrant email
		invalidEntry.EntrantEmail = "not_a_valid_email@"
		_, err = agent.CreateEntry(ctx, invalidEntry, &season, "5678")
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("creating an entry with existing entrant data must fail", func(t *testing.T) {
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
