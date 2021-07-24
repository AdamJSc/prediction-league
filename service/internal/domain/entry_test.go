package domain_test

import (
	"context"
	"errors"
	"fmt"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"sort"
	"testing"
	"time"
)

func TestNewEntryAgent(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		cl := &mockClock{}

		tt := []struct {
			er      domain.EntryRepository
			epr     domain.EntryPredictionRepository
			sr      domain.StandingsRepository
			sc      domain.SeasonCollection
			cl      domain.Clock
			wantErr error
		}{
			{nil, epr, sr, sc, cl, domain.ErrIsNil},
			{er, nil, sr, sc, cl, domain.ErrIsNil},
			{er, epr, nil, sc, cl, domain.ErrIsNil},
			{er, epr, sr, nil, cl, domain.ErrIsNil},
			{er, epr, sr, sc, nil, domain.ErrIsNil},
			{er, epr, sr, sc, cl, nil},
		}

		for idx, tc := range tt {
			agent, gotErr := domain.NewEntryAgent(tc.er, tc.epr, tc.sr, tc.sc, tc.cl)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && agent == nil {
				t.Fatalf("tc #%d: want non-empty agent, got nil", idx)
			}
		}
	})
}

func TestEntryAgent_CreateEntry(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: dt})
	if err != nil {
		t.Fatal(err)
	}

	season := domain.Season{
		ID: "199293_1",
		EntriesAccepted: domain.TimeFrame{
			From:  dt.Add(-12 * time.Hour),
			Until: dt.Add(12 * time.Hour),
		},
		Live: domain.TimeFrame{
			From:  dt.Add(12 * time.Hour),
			Until: dt.Add(24 * time.Hour),
		},
	}

	paymentMethod := "entry_payment_method"
	paymentRef := "entry_payment_ref"

	entryID, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	entry := domain.Entry{
		// these values should be populated
		EntrantName:     "Harry Redknapp",
		EntrantNickname: "MrHarryR",
		EntrantEmail:    "harry.redknapp@football.net",

		// these values should be overridden
		ID:            entryID,
		ShortCode:     "entry_short_code",
		SeasonID:      "entry_season_id",
		RealmName:     "entry_realm_name",
		Status:        "entry_status",
		PaymentMethod: &paymentMethod,
		PaymentRef:    &paymentRef,
		EntryPredictions: []domain.EntryPrediction{
			domain.NewEntryPrediction([]string{"entry_team_id_1", "entry_team_id_2"}),
		},
		ApprovedAt: &dt,
		CreatedAt:  time.Time{},
		UpdatedAt:  &dt,
	}

	t.Run("create a valid entry with a valid guard value must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

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
		expectedRealm := domain.RealmFromContext(ctx).Name
		expectedStatus := domain.EntryStatusPending

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
		if createdEntry.PaymentMethod != nil {
			expectedEmpty(t, "Entry.PaymentMethod", createdEntry.PaymentMethod)
		}
		if createdEntry.PaymentRef != nil {
			expectedEmpty(t, "Entry.PaymentRef", createdEntry.PaymentRef)
		}
		if len(createdEntry.EntryPredictions) != 0 {
			expectedEmpty(t, "Entry.EntryPredictions", createdEntry.EntryPredictions)
		}
		if createdEntry.ApprovedAt != nil {
			expectedEmpty(t, "Entry.ApprovedAt", *createdEntry.ApprovedAt)
		}
		if createdEntry.CreatedAt.Equal(time.Time{}) {
			expectedNonEmpty(t, "Entry.CreatedAt")
		}
		if createdEntry.UpdatedAt != nil {
			expectedEmpty(t, "Entry.UpdatedAt", *createdEntry.UpdatedAt)
		}

		// inserting same entry a second time should fail
		_, err = agent.CreateEntry(ctx, entry, &season)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create an entry with a nil season pointer must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.CreateEntry(ctx, entry, nil)
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("create an entry with an invalid guard value must fail", func(t *testing.T) {
		ctx, cancel := testContextWithGuardAttempt(t, "not_the_correct_realm_pin")
		defer cancel()

		_, err := agent.CreateEntry(ctx, entry, &season)
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("create an entry for a season that isn't accepting entries must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		seasonNotAcceptingEntries := season

		// entry window doesn't begin until tomorrow
		seasonNotAcceptingEntries.EntriesAccepted.From = time.Now().Add(24 * time.Hour)
		seasonNotAcceptingEntries.Live.From = time.Now().Add(48 * time.Hour)

		_, err := agent.CreateEntry(ctx, entry, &seasonNotAcceptingEntries)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}

		// entry window has already elapsed
		seasonNotAcceptingEntries.EntriesAccepted.From = time.Now().Add(-48 * time.Hour)
		seasonNotAcceptingEntries.Live.From = time.Now().Add(-24 * time.Hour)

		_, err = agent.CreateEntry(ctx, entry, &seasonNotAcceptingEntries)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create an entry with missing required fields must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

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

	t.Run("create an entry with invalid nickname must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

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
		ctx, cancel := testContextDefault(t)
		defer cancel()

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

func TestEntryAgent_AddEntryPredictionToEntry(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	t.Run("add an entry prediction to an existing entry with valid guard value must succeed", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		teamIDs := domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)

		// reverse order of team IDs to ensure this is still an acceptable set of rankings when adding an entry prediction
		var rankings domain.RankingCollection
		for i := len(teamIDs) - 1; i >= 0; i-- {
			rankings = append(rankings, teamIDs[i])
		}

		entryPrediction := domain.EntryPrediction{Rankings: rankings}

		entryWithPrediction, err := agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)
		if err != nil {
			t.Fatal(err)
		}

		if len(entryWithPrediction.EntryPredictions) != 1 {
			expectedGot(t, 1, len(entryWithPrediction.EntryPredictions))
		}

		if !gocmp.Equal(entryWithPrediction.EntryPredictions[0].Rankings, rankings) {
			t.Fatal(gocmp.Diff(rankings, entryWithPrediction.EntryPredictions[0].Rankings))
		}
	})

	t.Run("add an entry prediction to an existing entry with invalid guard attempt must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, "not_the_same_as_entry.ShortCode")
		defer cancel()

		entryPrediction := domain.EntryPrediction{Rankings: domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)}

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("add an entry prediction to an existing entry with invalid realm name must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		domain.RealmFromContext(ctx).Name = "NOT_TEST_REALM"

		entryPrediction := domain.EntryPrediction{Rankings: domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)}

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("add an entry prediction to a non-existing entry must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		entryPrediction := domain.EntryPrediction{Rankings: domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)}

		nonExistentEntryID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		nonExistentEntry := entry
		nonExistentEntry.ID = nonExistentEntryID

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, nonExistentEntry)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("add an entry prediction to an entry with an invalid season must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		entryPrediction := domain.EntryPrediction{Rankings: domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)}

		entryWithInvalidSeason := entry
		entryWithInvalidSeason.SeasonID = "not_a_valid_season"

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entryWithInvalidSeason)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("add an entry prediction to an entry whose season is not currently accepting entries must fail", func(t *testing.T) {
		// predictions are NOT accepted
		ts := testSeason.PredictionsAccepted.From.Add(-time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		entryPrediction := domain.EntryPrediction{Rankings: domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)}

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("add an entry prediction with rankings that include an invalid team ID must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		rankings := domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)
		rankings = append(rankings, domain.Ranking{ID: "not_a_valid_team_id"})
		expectedMessage := "Invalid Team ID: not_a_valid_team_id"

		entryPrediction := domain.EntryPrediction{Rankings: rankings}

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)

		verr, ok := err.(domain.ValidationError)
		if !ok {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
			return
		}
		if len(verr.Reasons) != 1 || verr.Reasons[0] != expectedMessage {
			expectedGot(t, domain.ValidationError{Reasons: []string{expectedMessage}}, verr)
		}
	})

	t.Run("add an entry prediction with rankings that include a missing team ID must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		rankings := domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)
		if len(rankings) < 1 {
			t.Fatalf("expected rankings length of at least 1, got %d", len(rankings))
		}

		uIdx := len(rankings) - 1
		missingID := rankings[uIdx].ID
		expectedMessage := fmt.Sprintf("Missing Team ID: %s", missingID)

		// trim final ranking
		rankings = rankings[:uIdx]

		entryPrediction := domain.EntryPrediction{Rankings: rankings}

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)

		verr, ok := err.(domain.ValidationError)
		if !ok {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
			return
		}
		if len(verr.Reasons) != 1 || verr.Reasons[0] != expectedMessage {
			expectedGot(t, domain.ValidationError{Reasons: []string{expectedMessage}}, verr)
		}
	})

	t.Run("add an entry prediction with rankings that include a duplicate team ID must fail", func(t *testing.T) {
		// predictions are accepted
		ts := testSeason.PredictionsAccepted.From.Add(time.Nanosecond)

		agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: ts})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextWithGuardAttempt(t, entry.ShortCode)
		defer cancel()

		rankings := domain.NewRankingCollectionFromIDs(testSeason.TeamIDs)
		if len(rankings) < 2 {
			t.Fatalf("expected rankings length of at least 2, got %d", len(rankings))
		}

		uIdx := len(rankings) - 1
		duplicateID := rankings[uIdx].ID
		expectedMessage := fmt.Sprintf("Duplicate Team ID: %s", duplicateID)

		// append duplicate ID to rankings
		rankings = append(rankings, domain.Ranking{ID: duplicateID})

		entryPrediction := domain.EntryPrediction{Rankings: rankings}

		_, err = agent.AddEntryPredictionToEntry(ctx, entryPrediction, entry)

		verr, ok := err.(domain.ValidationError)
		if !ok {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
			return
		}
		if len(verr.Reasons) != 1 || verr.Reasons[0] != expectedMessage {
			expectedGot(t, domain.ValidationError{Reasons: []string{expectedMessage}}, verr)
		}
	})
}

func TestEntryAgent_RetrieveEntryByID(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	for i := 0; i < 3; i++ {
		entry.EntryPredictions = append(entry.EntryPredictions, insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID)))
	}

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve an existent entry with valid credentials must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

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

		checkStringPtrMatch(t, entry.PaymentMethod, retrievedEntry.PaymentMethod)
		checkStringPtrMatch(t, entry.PaymentRef, retrievedEntry.PaymentRef)

		if len(entry.EntryPredictions) != len(retrievedEntry.EntryPredictions) {
			t.Fatal(gocmp.Diff(entry.EntryPredictions, retrievedEntry.EntryPredictions))
		}
		if entry.CreatedAt.In(utc) != retrievedEntry.CreatedAt.In(utc) {
			expectedGot(t, entry.CreatedAt, retrievedEntry.CreatedAt)
		}

		checkTimePtrMatch(t, entry.ApprovedAt, retrievedEntry.ApprovedAt)
		checkTimePtrMatch(t, entry.UpdatedAt, retrievedEntry.UpdatedAt)
	})

	t.Run("retrieve a non-existent entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveEntryByID(ctx, "not_a_valid_id")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve an entry with a mismatched realm must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		domain.RealmFromContext(ctx).Name = "DIFFERENT_REALM"

		_, err := agent.RetrieveEntryByID(ctx, entry.ID.String())
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_RetrieveEntryByEntrantEmail(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	for i := 0; i < 3; i++ {
		entry.EntryPredictions = append(entry.EntryPredictions, insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID)))
	}

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve an existent entry with valid credentials must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// should succeed
		retrievedEntry, err := agent.RetrieveEntryByEntrantEmail(ctx, entry.EntrantEmail, entry.SeasonID, entry.RealmName)
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

		checkStringPtrMatch(t, entry.PaymentMethod, retrievedEntry.PaymentMethod)
		checkStringPtrMatch(t, entry.PaymentRef, retrievedEntry.PaymentRef)

		if len(entry.EntryPredictions) != len(retrievedEntry.EntryPredictions) {
			t.Fatal(gocmp.Diff(entry.EntryPredictions, retrievedEntry.EntryPredictions))
		}
		if entry.CreatedAt.In(utc) != retrievedEntry.CreatedAt.In(utc) {
			expectedGot(t, entry.CreatedAt, retrievedEntry.CreatedAt)
		}

		checkTimePtrMatch(t, entry.ApprovedAt, retrievedEntry.ApprovedAt)
		checkTimePtrMatch(t, entry.UpdatedAt, retrievedEntry.UpdatedAt)
	})

	t.Run("retrieve a non-existent entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveEntryByEntrantEmail(ctx, "not_an_existent_email", entry.SeasonID, entry.RealmName)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		_, err = agent.RetrieveEntryByEntrantEmail(ctx, entry.EntrantEmail, "not_an_existent_season_id", entry.RealmName)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		_, err = agent.RetrieveEntryByEntrantEmail(ctx, entry.EntrantEmail, entry.SeasonID, "not_an_existent_realm_name")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve an entry with a mismatched realm must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		domain.RealmFromContext(ctx).Name = "DIFFERENT_REALM"

		_, err := agent.RetrieveEntryByEntrantEmail(ctx, entry.EntrantEmail, entry.SeasonID, entry.RealmName)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_RetrieveEntryByEntrantNickname(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	for i := 0; i < 3; i++ {
		entry.EntryPredictions = append(entry.EntryPredictions, insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID)))
	}

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve an existent entry with valid credentials must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// should succeed
		retrievedEntry, err := agent.RetrieveEntryByEntrantNickname(ctx, entry.EntrantNickname, entry.SeasonID, entry.RealmName)
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

		checkStringPtrMatch(t, entry.PaymentMethod, retrievedEntry.PaymentMethod)
		checkStringPtrMatch(t, entry.PaymentRef, retrievedEntry.PaymentRef)

		if len(entry.EntryPredictions) != len(retrievedEntry.EntryPredictions) {
			t.Fatal(gocmp.Diff(entry.EntryPredictions, retrievedEntry.EntryPredictions))
		}
		if entry.CreatedAt.In(utc) != retrievedEntry.CreatedAt.In(utc) {
			expectedGot(t, entry.CreatedAt, retrievedEntry.CreatedAt)
		}

		checkTimePtrMatch(t, entry.ApprovedAt, retrievedEntry.ApprovedAt)
		checkTimePtrMatch(t, entry.UpdatedAt, retrievedEntry.UpdatedAt)
	})

	t.Run("retrieve a non-existent entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveEntryByEntrantNickname(ctx, "not_an_existent_nickname", entry.SeasonID, entry.RealmName)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		_, err = agent.RetrieveEntryByEntrantNickname(ctx, entry.EntrantNickname, "not_an_existent_season_id", entry.RealmName)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		_, err = agent.RetrieveEntryByEntrantNickname(ctx, entry.EntrantNickname, entry.SeasonID, "not_an_existent_realm_name")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve an entry with a mismatched realm must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		domain.RealmFromContext(ctx).Name = "DIFFERENT_REALM"

		_, err := agent.RetrieveEntryByEntrantNickname(ctx, entry.EntrantNickname, entry.SeasonID, entry.RealmName)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_RetrieveEntriesBySeasonID(t *testing.T) {
	defer truncate(t)

	// generate entries
	var generatedEntries = []domain.Entry{
		generateTestEntry(t,
			"Harry Redknapp",
			"MrHarryR",
			"harry.redknapp@football.net",
		),
		generateTestEntry(t,
			"Jamie Redknapp",
			"MrJamieR",
			"jamie.redknapp@football.net",
		),
		generateTestEntry(t,
			"Frank Lampard",
			"FrankieLamps",
			"frank.lampard@football.net",
		),
	}

	// set our second entry to an invalid season ID, so that it won't be retrieved by our agent method
	generatedEntries[1].SeasonID = "nnnnnn"

	// set an approved date on our third entry so that we can verify retrieval of approved entries only
	now := time.Now()
	generatedEntries[2].ApprovedAt = &now

	// insert entries
	var entries []domain.Entry
	for _, entry := range generatedEntries {
		entries = append(entries, insertEntry(t, entry))
	}

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve existing entries with valid credentials must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// should succeed
		retrievedEntries, err := agent.RetrieveEntriesBySeasonID(ctx, testSeason.ID, false)
		if err != nil {
			t.Fatal(err)
		}

		sort.SliceStable(retrievedEntries, func(i, j int) bool {
			return retrievedEntries[i].EntrantNickname > retrievedEntries[j].EntrantNickname
		})

		if len(retrievedEntries) != 2 {
			t.Fatalf("expected length of 2, got %d", len(retrievedEntries))
		}

		if retrievedEntries[0].EntrantNickname != "MrHarryR" {
			expectedGot(t, "MrHarryR", retrievedEntries[0].EntrantNickname)
		}

		if retrievedEntries[1].EntrantNickname != "FrankieLamps" {
			expectedGot(t, "FrankieLamps", retrievedEntries[1].EntrantNickname)
		}
	})

	t.Run("retrieve only those existing entries that have been approved, with valid credentials must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// should succeed
		retrievedEntries, err := agent.RetrieveEntriesBySeasonID(ctx, testSeason.ID, true)
		if err != nil {
			t.Fatal(err)
		}

		if len(retrievedEntries) != 1 {
			t.Fatalf("expected length of 1, got %d", len(retrievedEntries))
		}

		if retrievedEntries[0].EntrantNickname != "FrankieLamps" {
			expectedGot(t, "FrankieLamps", retrievedEntries[0].EntrantNickname)
		}
	})

	t.Run("retrieve non-existent entries must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveEntriesBySeasonID(ctx, "not_a_valid_season_id", false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestEntryAgent_UpdateEntry(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("update an existent entry with a valid alternative entry must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		// define changed entry values
		changedEntryPaymentRef := "changed_entry_payment_ref"

		changedEntry := domain.Entry{
			ID:              entry.ID,
			ShortCode:       "changed_entry_short_code",
			SeasonID:        "67890",
			RealmName:       entry.RealmName,
			EntrantName:     "Jamie Redknapp",
			EntrantNickname: "MrJamieR",
			EntrantEmail:    "jamie.redknapp@football.net",
			Status:          domain.EntryStatusReady,
			PaymentRef:      &changedEntryPaymentRef,
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
		if entry.CreatedAt.In(utc) != updatedEntry.CreatedAt.In(utc) {
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
		if updatedEntry.UpdatedAt == nil {
			expectedNonEmpty(t, "Entry.UpdatedAt")
		}
	})

	t.Run("update an existent entry with a changed realm must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.UpdateEntry(ctx, domain.Entry{ID: entry.ID, RealmName: "NOT_THE_ORIGINAL_REALM"})
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update a non-existent entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entryID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.UpdateEntry(ctx, domain.Entry{ID: entryID, RealmName: entry.RealmName})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("update an existing entry with missing required fields must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

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
		ctx, cancel := testContextDefault(t)
		defer cancel()

		var invalidEntry domain.Entry
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
		invalidEntry.PaymentMethod = &invalidPaymentMethod
		_, err = agent.UpdateEntry(ctx, invalidEntry)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})
}

func TestEntryAgent_UpdateEntryPaymentDetails(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	paymentRef := "ABCD1234"

	ctx, cancel := testContextDefault(t)
	defer cancel()

	t.Run("update payment details for an existent entry with valid credentials must succeed", func(t *testing.T) {
		entryWithPaymentDetails, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
			paymentRef,
			true,
		)
		if err != nil {
			t.Fatal(err)
		}

		if domain.EntryPaymentMethodPayPal != *entryWithPaymentDetails.PaymentMethod {
			expectedGot(t, domain.EntryPaymentMethodPayPal, *entryWithPaymentDetails.PaymentMethod)
		}

		if paymentRef != *entryWithPaymentDetails.PaymentRef {
			expectedGot(t, paymentRef, *entryWithPaymentDetails.PaymentRef)
		}
	})

	t.Run("update invalid payment method for an existent entry must fail", func(t *testing.T) {
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			"not_a_valid_payment_method",
			paymentRef,
			true,
		)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("update entry with payment method 'other' when this is not accepted must fail", func(t *testing.T) {
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodOther,
			paymentRef,
			false,
		)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update missing payment ref for an existent entry must fail", func(t *testing.T) {
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
			"",
			true,
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
			true,
		)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid realm must fail", func(t *testing.T) {
		domain.RealmFromContext(ctx).Name = "DIFFERENT_REALM"

		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
			paymentRef,
			true,
		)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("update payment details for an existing entry with an invalid status must fail", func(t *testing.T) {
		entryWithInvalidStatus := generateTestEntry(t,
			"Jamie Redknapp",
			"MrJamieR",
			"jamie.redknapp@football.net",
		)
		entryWithInvalidStatus.Status = domain.EntryStatusPaid
		entryWithInvalidStatus = insertEntry(t, entryWithInvalidStatus)

		// now running the operation we're testing should fail
		_, err := agent.UpdateEntryPaymentDetails(
			ctx,
			entry.ID.String(),
			domain.EntryPaymentMethodPayPal,
			paymentRef,
			true,
		)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_ApproveEntryByShortCode(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	entryWithPaidStatus := generateTestEntry(t,
		"Jamie Redknapp",
		"MrJamieR",
		"jamie.redknapp@football.net",
	)
	entryWithPaidStatus.Status = domain.EntryStatusPaid
	entryWithPaidStatus = insertEntry(t, entryWithPaidStatus)

	entryWithReadyStatus := generateTestEntry(t,
		"Frank Lampard",
		"FrankieLamps",
		"frank.lampard@football.net",
	)
	entryWithReadyStatus.Status = domain.EntryStatusReady
	entryWithReadyStatus = insertEntry(t, entryWithReadyStatus)

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{t: dt})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("approve existent entry short code with valid credentials must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		defer cancel()

		// attempt to approve entry with paid status
		approvedEntry, err := agent.ApproveEntryByShortCode(ctx, entryWithPaidStatus.ShortCode)
		if err != nil {
			t.Fatal(err)
		}
		if !approvedEntry.IsApproved() {
			expectedGot(t, "approved entry true", "approved entry false")
		}
		if approvedEntry.ApprovedAt == nil {
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
		if approvedEntry.ApprovedAt == nil {
			expectedNonEmpty(t, "Entry.ApprovedAt")
		}
	})

	t.Run("approve existent entry with invalid credentials must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		// basic auth success on context defaults to false so this should fail naturally
		defer cancel()

		_, err := agent.ApproveEntryByShortCode(ctx, entry.ShortCode)
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("approve non-existent entry with valid credentials must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		defer cancel()

		_, err := agent.ApproveEntryByShortCode(ctx, "non_existent_short_code")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("approve existent entry with invalid realm must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		defer cancel()

		domain.RealmFromContext(ctx).Name = "DIFFERENT_REALM"
		_, err := agent.ApproveEntryByShortCode(ctx, entry.ShortCode)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("approve existent entry with pending status must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		defer cancel()

		// initial entry object should still have default "pending" status so just attempt to approve this
		_, err := agent.ApproveEntryByShortCode(ctx, entry.ShortCode)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("approve existent entry that has already been approved must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		defer cancel()

		// just try to approve the same entry again
		_, err := agent.ApproveEntryByShortCode(ctx, entryWithPaidStatus.ShortCode)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestEntryAgent_GetEntryPredictionByTimestamp(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	var entryPredictions []domain.EntryPrediction
	for i := 1; i <= 2; i++ {
		// ensure each entry prediction is 48 hours apart
		days := time.Duration(i) * 48 * time.Hour

		entryPrediction := generateTestEntryPrediction(t, entry.ID)
		entryPrediction.CreatedAt = time.Now().Add(days)
		entryPrediction = insertEntryPrediction(t, entryPrediction)

		entryPredictions = append(entryPredictions, entryPrediction)
	}

	agent, err := domain.NewEntryAgent(er, epr, sr, sc, &mockClock{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve entry prediction by timestamp that occurs before earliest prediction must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveEntryPredictionByTimestamp(ctx, entry, entryPredictions[0].CreatedAt.Add(-time.Second))
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve entry prediction by timestamp that equals earliest prediction must return earliest prediction", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		actualEntryPrediction, err := agent.RetrieveEntryPredictionByTimestamp(ctx, entry, entryPredictions[0].CreatedAt)
		if err != nil {
			t.Fatal(err)
		}

		expectedEntryPrediction := entryPredictions[0]

		if !actualEntryPrediction.CreatedAt.Equal(expectedEntryPrediction.CreatedAt) {
			expectedGot(t, expectedEntryPrediction.CreatedAt, actualEntryPrediction.CreatedAt)
		}
	})

	t.Run("retrieve entry prediction by timestamp that occurs before latest prediction must return previous prediction", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		actualEntryPrediction, err := agent.RetrieveEntryPredictionByTimestamp(ctx, entry, entryPredictions[1].CreatedAt.Add(-time.Second))
		if err != nil {
			t.Fatal(err)
		}

		expectedEntryPrediction := entryPredictions[0]

		if !actualEntryPrediction.CreatedAt.Equal(expectedEntryPrediction.CreatedAt) {
			expectedGot(t, expectedEntryPrediction.CreatedAt, actualEntryPrediction.CreatedAt)
		}
	})

	t.Run("retrieve entry prediction by timestamp that equals latest prediction must return latest prediction", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		actualEntryPrediction, err := agent.RetrieveEntryPredictionByTimestamp(ctx, entry, entryPredictions[1].CreatedAt)
		if err != nil {
			t.Fatal(err)
		}

		expectedEntryPrediction := entryPredictions[1]

		if !actualEntryPrediction.CreatedAt.Equal(expectedEntryPrediction.CreatedAt) {
			expectedGot(t, expectedEntryPrediction.CreatedAt, actualEntryPrediction.CreatedAt)
		}
	})

	t.Run("retrieve entry prediction by timestamp that occurs after latest prediction must return latest prediction", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		actualEntryPrediction, err := agent.RetrieveEntryPredictionByTimestamp(ctx, entry, entryPredictions[1].CreatedAt.Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}

		expectedEntryPrediction := entryPredictions[1]

		if !actualEntryPrediction.CreatedAt.Equal(expectedEntryPrediction.CreatedAt) {
			expectedGot(t, expectedEntryPrediction.CreatedAt, actualEntryPrediction.CreatedAt)
		}
	})
}

// TODO - tests for EntryAgent.RetrieveEntryPredictionsForActiveSeasonByTimestamp

func TestEntryAgent_GetPredictionRankingLimit(t *testing.T) {
	defer truncate(t)

	dt := time.Date(2018, 5, 26, 14, 0, 0, 0, time.FixedZone("Europe/London", 3600))

	e := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))
	e2 := insertEntry(t, generateTestEntry(t,
		"Jamie Redknapp",
		"MrJamieR",
		"jamie.redknapp@football.net",
	))

	// entry prediction 1 created one hour after base date
	ep1 := generateTestEntryPrediction(t, e.ID)
	ep1.CreatedAt = dt.Add(time.Hour)
	ep1 = insertEntryPrediction(t, ep1)

	// standings created two hours after base date
	st := generateTestStandings(t)
	st.CreatedAt = dt.Add(2 * time.Hour)
	st = insertStandings(t, st)

	// entry prediction 2 created three hours after base date
	ep2 := generateTestEntryPrediction(t, e.ID)
	ep2.CreatedAt = dt.Add(3 * time.Hour)
	ep2 = insertEntryPrediction(t, ep2)

	t.Run("getting prediction ranking limit for the provided timestamp must return the expected limit value", func(t *testing.T) {
		tt := []struct {
			name      string
			entry     domain.Entry
			ts        time.Time
			wantLimit int
		}{
			{
				name:      "ts before first entry prediction, before standings (no limit)",
				entry:     e,
				ts:        ep1.CreatedAt.Add(-time.Second),
				wantLimit: domain.RankingLimitNone,
			},
			{
				name:      "ts at time of first entry prediction, before standings (no limit)",
				entry:     e,
				ts:        ep1.CreatedAt,
				wantLimit: domain.RankingLimitNone,
			},
			{
				name:      "ts after first entry prediction, before standings (no limit)",
				entry:     e,
				ts:        ep1.CreatedAt.Add(time.Second),
				wantLimit: domain.RankingLimitNone,
			},
			{
				name:      "ts at time of standings, entry does not have a prediction already (no limit)",
				entry:     e2,
				ts:        st.CreatedAt,
				wantLimit: domain.RankingLimitNone,
			},
			{
				name:      "ts at time of standings, entry does not have a new prediction since standings (regular limit)",
				entry:     e,
				ts:        st.CreatedAt,
				wantLimit: domain.RankingLimitRegular,
			},
			{
				name:      "ts after standings, entry does not have a new prediction since standings (regular limit)",
				entry:     e,
				ts:        st.CreatedAt.Add(time.Second),
				wantLimit: domain.RankingLimitRegular,
			},
			{
				name:      "ts after standings, at time of new prediction (limit of 0)",
				entry:     e,
				ts:        ep2.CreatedAt,
				wantLimit: 0,
			},
			{
				name:      "ts after standings, after time of new prediction (limit of 0)",
				entry:     e,
				ts:        ep2.CreatedAt.Add(time.Second),
				wantLimit: 0,
			},
		}

		for _, tc := range tt {
			cl := &mockClock{t: tc.ts}
			agent, err := domain.NewEntryAgent(er, epr, sr, sc, cl)
			if err != nil {
				t.Fatal(err)
			}
			gotLimit, err := agent.GetPredictionRankingLimit(context.Background(), tc.entry)
			if err != nil {
				t.Fatal(err)
			}
			if gotLimit != tc.wantLimit {
				t.Fatalf("tc '%s': want limit %d, got %d", tc.name, tc.wantLimit, gotLimit)
			}
		}
	})
}

func TestEntryAgent_CheckRankingLimit(t *testing.T) {
	defer truncate(t)

	dt := time.Date(2018, 5, 26, 14, 0, 0, 0, time.FixedZone("Europe/London", 3600))
	cl := &mockClock{t: dt}
	pughID, fletchID, pitmanID := "marc", "steve", "brett"

	baseRankings := domain.RankingCollection{
		{pughID, 1},
		{fletchID, 2},
		{pitmanID, 3},
	}

	entryWithPred := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	// insert entry prediction for entryWithPred
	baseEntryPred := insertEntryPrediction(t, domain.EntryPrediction{
		ID:        uuid.New(),
		EntryID:   entryWithPred.ID,
		Rankings:  baseRankings,
		CreatedAt: dt.Add(-time.Second), // ensure this is retrieved as latest entry prediction relative to test date
	})

	entryNoPred := insertEntry(t, generateTestEntry(t,
		"Jamie Redknapp",
		"MrJamieR",
		"jamie.redknapp@football.net",
	))

	// changedRankings entails 2 changes compared to baseRankings
	changedRankings := domain.RankingCollection{
		{pughID, 1},
		{pitmanID, 2}, // swapped with fletch
		{fletchID, 3}, // swapped with pitman
	}

	entryPredChangedRank := domain.EntryPrediction{
		EntryID:  entryWithPred.ID,
		Rankings: changedRankings,
	}

	tt := []struct {
		name       string
		e          domain.Entry
		ep         domain.EntryPrediction
		lim        int
		wantErrMsg string
	}{
		{
			name: "no limit, must return no error",
			lim:  domain.RankingLimitNone,
			// want no error
		},
		{
			name: "if no existing prediction, must return no error",
			e:    entryNoPred,
			// want no error
		},
		{
			name:       "no changes, must return expected error",
			e:          entryWithPred,
			ep:         baseEntryPred,
			wantErrMsg: "no changes to rankings",
		},
		{
			name: "2 changes with limit of 3, must return no error",
			e:    entryWithPred,
			ep:   entryPredChangedRank,
			lim:  3,
			// want no error
		},
		{
			name: "2 changes with limit of 2, must return no error",
			e:    entryWithPred,
			ep:   entryPredChangedRank,
			lim:  2,
			// want no error
		},
		{
			name:       "2 changes with limit of 1, must return expected error",
			e:          entryWithPred,
			ep:         entryPredChangedRank,
			lim:        1,
			wantErrMsg: "cannot change 2 rankings (1 permitted)",
		},
		{
			name:       "2 changes with limit of 0, must return expected error",
			e:          entryWithPred,
			ep:         entryPredChangedRank,
			lim:        0,
			wantErrMsg: "cannot change 2 rankings (0 permitted)",
		},
	}

	for _, tc := range tt {
		agent, err := domain.NewEntryAgent(er, epr, sr, sc, cl)
		if err != nil {
			t.Fatal(err)
		}
		gotErr := agent.CheckRankingLimit(context.Background(), tc.lim, tc.ep, tc.e)
		if tc.wantErrMsg == "" && gotErr != nil {
			t.Fatalf("tc '%s': want no error, got %s (%T)", tc.name, gotErr, gotErr)
		}
		if tc.wantErrMsg != "" {
			if gotErr == nil || !errors.As(gotErr, &domain.ConflictError{}) || gotErr.Error() != tc.wantErrMsg {
				t.Fatalf("tc '%s': want conflict error %s, got %s (%T)", tc.name, tc.wantErrMsg, gotErr, gotErr)
			}
		}
		if tc.wantErrMsg != "" {
			if gotErr == nil || gotErr.Error() != tc.wantErrMsg {
				t.Fatalf("tc '%s': want error message '%s', got '%s' (%T)", tc.name, tc.wantErrMsg, gotErr, gotErr)
			}
		}
	}
}

func checkTimePtrMatch(t *testing.T, exp *time.Time, got *time.Time) {
	switch {
	case exp == nil && got == nil:
		return
	case exp == nil && got != nil:
		expectedGot(t, nil, *got)
	case exp != nil && got == nil:
		expectedGot(t, *exp, nil)
	}
	if !exp.In(time.UTC).Equal(got.In(time.UTC)) {
		expectedGot(t, *exp, *got)
	}
}
