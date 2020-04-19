package domain_test

import (
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"reflect"
	"testing"
	"time"
)

func TestEntryAgent_CreateEntry(t *testing.T) {
	defer truncate(t)

	agent := domain.EntryAgent{EntryAgentInjector: injector{db: db}}

	t.Run("creating a valid entry with a valid PIN must succeed", func(t *testing.T) {
		ctx := domain.NewContext()
		ctx.SetRealm("TEST_REALM")
		ctx.SetRealmPIN("5678")

		season := domain.Season{
			ID:          "199293_1",
			EntriesFrom: time.Now().Add(-24 * time.Hour),
			StartDate:   time.Now().Add(24 * time.Hour),
		}

		sourceEntry := domain.Entry{
			EntrantName:     "John Doe",
			EntrantNickname: "JohnD",
			EntrantEmail:    "john.doe@hello.net",
		}

		createdEntry := sourceEntry

		// should succeed
		if err := agent.CreateEntry(ctx, &createdEntry, &season, "5678"); err != nil {
			t.Fatal(err)
		}

		// check raw values that shouldn't have changed
		if !cmp.Equal(sourceEntry.EntrantName, createdEntry.EntrantName)().Success() {
			expectedGot(t, sourceEntry.EntrantName, createdEntry.EntrantName)
		}
		if !cmp.Equal(sourceEntry.EntrantNickname, createdEntry.EntrantNickname)().Success() {
			expectedGot(t, sourceEntry.EntrantNickname, createdEntry.EntrantNickname)
		}
		if !cmp.Equal(sourceEntry.EntrantEmail, createdEntry.EntrantEmail)().Success() {
			expectedGot(t, sourceEntry.EntrantEmail, createdEntry.EntrantEmail)
		}
		if !cmp.Equal(sourceEntry.PaymentRef, createdEntry.PaymentRef)().Success() {
			expectedGot(t, sourceEntry.PaymentRef, createdEntry.PaymentRef)
		}
		if !cmp.DeepEqual(sourceEntry.TeamIDSequence, createdEntry.TeamIDSequence)().Success() {
			expectedGot(t, sourceEntry.TeamIDSequence, createdEntry.TeamIDSequence)
		}
		if !cmp.Equal(sourceEntry.UpdatedAt, createdEntry.UpdatedAt)().Success() {
			expectedGot(t, sourceEntry.UpdatedAt, createdEntry.UpdatedAt)
		}

		// check sanitised values
		expectedSeasonID := season.ID
		expectedRealm := ctx.GetRealm()
		expectedStatus := "pending"

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
		if !cmp.Equal(expectedStatus, createdEntry.Status)().Success() {
			expectedGot(t, expectedStatus, createdEntry.Status)
		}
		if cmp.Equal(time.Time{}, createdEntry.CreatedAt)().Success() {
			expectedNonEmpty(t, "Entry.CreatedAt")
		}

		// inserting same entry a second time should fail
		err := agent.CreateEntry(ctx, &createdEntry, &season, "5678")
		if reflect.TypeOf(err) != reflect.TypeOf(domain.ConflictError{}) {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}
