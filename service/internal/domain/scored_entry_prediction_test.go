package domain_test

import (
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert/cmp"
)

func TestNewScoredEntryPredictionAgent(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		tt := []struct {
			er      domain.EntryRepository
			epr     domain.EntryPredictionRepository
			sr      domain.StandingsRepository
			sepr    domain.ScoredEntryPredictionRepository
			wantErr error
		}{
			{nil, epr, sr, sepr, domain.ErrIsNil},
			{er, nil, sr, sepr, domain.ErrIsNil},
			{er, epr, nil, sepr, domain.ErrIsNil},
			{er, epr, sr, nil, domain.ErrIsNil},
			{er, epr, sr, sepr, nil},
		}

		for idx, tc := range tt {
			agent, gotErr := domain.NewScoredEntryPredictionAgent(tc.er, tc.epr, tc.sr, tc.sepr)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && agent == nil {
				t.Fatalf("tc #%d: want non-empty agent, got nil", idx)
			}
		}
	})
}

func TestScoredEntryPredictionAgent_CreateScoredEntryPrediction(t *testing.T) {
	t.Cleanup(truncate)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))
	scoredEntryPrediction := generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID)

	t.Run("create valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		createdScoredEntryPrediction, err := agent.CreateScoredEntryPrediction(ctx, scoredEntryPrediction)
		if err != nil {
			t.Fatal(err)
		}

		var emptyTime time.Time
		if createdScoredEntryPrediction.EntryPredictionID != scoredEntryPrediction.EntryPredictionID {
			expectedGot(t, scoredEntryPrediction.EntryPredictionID, createdScoredEntryPrediction.EntryPredictionID)
		}
		if createdScoredEntryPrediction.StandingsID != scoredEntryPrediction.StandingsID {
			expectedGot(t, scoredEntryPrediction.StandingsID, createdScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(createdScoredEntryPrediction.Rankings, scoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(scoredEntryPrediction.Rankings, createdScoredEntryPrediction.Rankings))
		}
		if createdScoredEntryPrediction.Score != scoredEntryPrediction.Score {
			expectedGot(t, scoredEntryPrediction.Score, createdScoredEntryPrediction.Score)
		}
		if createdScoredEntryPrediction.CreatedAt.Equal(emptyTime) {
			expectedNonEmpty(t, "CreatedAt")
		}
		if createdScoredEntryPrediction.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", *createdScoredEntryPrediction.UpdatedAt)
		}

		// inserting same scored entry prediction a second time should fail
		_, err = agent.CreateScoredEntryPrediction(ctx, scoredEntryPrediction)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create invalid scored entry prediction must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		missingEntryPredictionID := scoredEntryPrediction
		missingEntryPredictionID.EntryPredictionID = uuid.UUID{}
		_, err := agent.CreateScoredEntryPrediction(ctx, missingEntryPredictionID)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingStandingsID := scoredEntryPrediction
		missingStandingsID.StandingsID = uuid.UUID{}
		_, err = agent.CreateScoredEntryPrediction(ctx, missingStandingsID)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("create scored entry prediction for non-existent entry prediction or standings must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		invalidEntryPredictionID := scoredEntryPrediction
		invalidEntryPredictionID.EntryPredictionID = nonExistentID
		_, err = agent.CreateScoredEntryPrediction(ctx, invalidEntryPredictionID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		invalidStandingsID := scoredEntryPrediction
		invalidStandingsID.StandingsID = nonExistentID
		_, err = agent.CreateScoredEntryPrediction(ctx, invalidStandingsID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestScoredEntryPredictionAgent_UpdateScoredEntryPrediction(t *testing.T) {
	t.Cleanup(truncate)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))
	scoredEntryPrediction := insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID))

	t.Run("update valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		changedScoredEntryPrediction := scoredEntryPrediction
		changedScoredEntryPrediction.Score = 456
		changedScoredEntryPrediction.Rankings = domain.NewRankingWithScoreCollectionFromIDs([]string{"changedID_1", "changedID_2", "changedID_3"})

		updatedScoredEntryPrediction, err := agent.UpdateScoredEntryPrediction(ctx, changedScoredEntryPrediction)
		if err != nil {
			t.Fatal(err)
		}

		if updatedScoredEntryPrediction.EntryPredictionID != changedScoredEntryPrediction.EntryPredictionID {
			expectedGot(t, changedScoredEntryPrediction.EntryPredictionID, updatedScoredEntryPrediction.EntryPredictionID)
		}
		if updatedScoredEntryPrediction.StandingsID != changedScoredEntryPrediction.StandingsID {
			expectedGot(t, changedScoredEntryPrediction.StandingsID, updatedScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(updatedScoredEntryPrediction.Rankings, changedScoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(changedScoredEntryPrediction.Rankings, updatedScoredEntryPrediction.Rankings))
		}
		if updatedScoredEntryPrediction.Score != changedScoredEntryPrediction.Score {
			expectedGot(t, changedScoredEntryPrediction.Score, updatedScoredEntryPrediction.Score)
		}
		if !updatedScoredEntryPrediction.CreatedAt.Equal(changedScoredEntryPrediction.CreatedAt) {
			expectedGot(t, changedScoredEntryPrediction.CreatedAt, updatedScoredEntryPrediction.CreatedAt)
		}
		if updatedScoredEntryPrediction.UpdatedAt == nil {
			expectedNonEmpty(t, "UpdatedAt")
		}
	})

	t.Run("update non-existent scored entry prediction must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		nonExistentEntryPredictionID := scoredEntryPrediction
		nonExistentEntryPredictionID.EntryPredictionID = nonExistentID
		_, err = agent.UpdateScoredEntryPrediction(ctx, nonExistentEntryPredictionID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		nonExistentStandingsID := scoredEntryPrediction
		nonExistentStandingsID.StandingsID = nonExistentID
		_, err = agent.UpdateScoredEntryPrediction(ctx, nonExistentStandingsID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

// TODO - tests for GenerateScoredEntryPrediction

func TestScoredEntryPredictionAgent_RetrieveScoredEntryPredictionByIDs(t *testing.T) {
	t.Cleanup(truncate)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	e := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	ep := insertEntryPrediction(t, generateTestEntryPrediction(t, e.ID))
	st := insertStandings(t, generateTestStandings(t))

	now := time.Now().Truncate(time.Second)
	sep := generateTestScoredEntryPrediction(t, ep.ID, st.ID)
	sep.CreatedAt = now
	sep = insertScoredEntryPrediction(t, sep)

	t.Run("retrieve existent scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		rtrSep, err := agent.RetrieveScoredEntryPredictionByIDs(ctx, ep.ID.String(), st.ID.String())
		if err != nil {
			t.Fatal(err)
		}

		if rtrSep.EntryPredictionID != sep.EntryPredictionID {
			expectedGot(t, sep.EntryPredictionID, rtrSep.EntryPredictionID)
		}
		if rtrSep.StandingsID != sep.StandingsID {
			expectedGot(t, sep.StandingsID, rtrSep.StandingsID)
		}
		if !gocmp.Equal(rtrSep.Rankings, sep.Rankings) {
			t.Fatal(gocmp.Diff(sep.Rankings, rtrSep.Rankings))
		}
		if rtrSep.Score != sep.Score {
			expectedGot(t, sep.Score, rtrSep.Score)
		}
		if !rtrSep.CreatedAt.In(utc).Equal(now.In(utc)) {
			expectedGot(t, now, rtrSep.CreatedAt)
		}
		if rtrSep.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", *rtrSep.UpdatedAt)
		}
	})

	t.Run("retrieve non-existent scored entry prediction must produce expected error", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		newUUID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		nID := newUUID.String()

		if _, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, nID, sep.StandingsID.String()); !errors.As(err, &domain.NotFoundError{}) {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
		if _, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, sep.EntryPredictionID.String(), nID); !errors.As(err, &domain.NotFoundError{}) {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestScoredEntryPredictionAgent_RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(t *testing.T) {
	t.Cleanup(truncate)

	agent, err := domain.NewScoredEntryPredictionAgent(er, epr, sr, sepr)
	if err != nil {
		t.Fatal(err)
	}

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	standings := insertStandings(t, generateTestStandings(t))
	now := time.Now().Truncate(time.Second)

	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	scoredEntryPrediction := generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID)
	scoredEntryPrediction.CreatedAt = now
	scoredEntryPrediction = insertScoredEntryPrediction(t, scoredEntryPrediction)

	differentEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	mostRecentScoredEntryPrediction := generateTestScoredEntryPrediction(t, differentEntryPrediction.ID, standings.ID)
	mostRecentScoredEntryPrediction.CreatedAt = now.Add(time.Second) // created most recently
	mostRecentScoredEntryPrediction = insertScoredEntryPrediction(t, mostRecentScoredEntryPrediction)

	t.Run("retrieve latest scored entry prediction by existent entry id and round number must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedScoredEntryPrediction, err := agent.RetrieveLatestScoredEntryPredictionByEntryIDAndRoundNumber(
			ctx,
			entry.ID.String(),
			standings.RoundNumber,
		)
		if err != nil {
			t.Fatal(err)
		}

		if retrievedScoredEntryPrediction.EntryPredictionID != mostRecentScoredEntryPrediction.EntryPredictionID {
			expectedGot(t, mostRecentScoredEntryPrediction.EntryPredictionID, retrievedScoredEntryPrediction.EntryPredictionID)
		}
		if retrievedScoredEntryPrediction.StandingsID != mostRecentScoredEntryPrediction.StandingsID {
			expectedGot(t, mostRecentScoredEntryPrediction.StandingsID, retrievedScoredEntryPrediction.StandingsID)
		}
		if !gocmp.Equal(retrievedScoredEntryPrediction.Rankings, mostRecentScoredEntryPrediction.Rankings) {
			t.Fatal(gocmp.Diff(mostRecentScoredEntryPrediction.Rankings, retrievedScoredEntryPrediction.Rankings))
		}
		if retrievedScoredEntryPrediction.Score != mostRecentScoredEntryPrediction.Score {
			expectedGot(t, mostRecentScoredEntryPrediction.Score, retrievedScoredEntryPrediction.Score)
		}
		if !retrievedScoredEntryPrediction.CreatedAt.In(utc).Equal(mostRecentScoredEntryPrediction.CreatedAt.In(utc)) {
			expectedGot(t, mostRecentScoredEntryPrediction.CreatedAt, retrievedScoredEntryPrediction.CreatedAt)
		}
		if retrievedScoredEntryPrediction.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", *retrievedScoredEntryPrediction.UpdatedAt)
		}
	})

	t.Run("retrieve latest scored entry prediction by non-existent entry id must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, nonExistentID.String(), standings.ID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve latest scored entry prediction by non-existent entry id must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveScoredEntryPredictionByIDs(ctx, entry.ID.String(), nonExistentID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}
