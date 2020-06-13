package domain_test

import (
	"github.com/LUSHDigital/uuid"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"testing"
	"time"
)

func TestScoredEntrySelectionAgent_CreateScoredEntrySelection(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entrySelection := insertEntrySelection(t, generateTestEntrySelection(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))

	scoredEntrySelection := generateTestScoredEntrySelection(t, entrySelection.ID, standings.ID)

	agent := domain.ScoredEntrySelectionAgent{
		ScoredEntrySelectionAgentInjector: injector{db: db},
	}

	t.Run("create valid scored entry selection must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		createdScoredEntrySelection, err := agent.CreateScoredEntrySelection(ctx, scoredEntrySelection)
		if err != nil {
			t.Fatal(err)
		}

		var emptyTime time.Time
		if createdScoredEntrySelection.EntrySelectionID != scoredEntrySelection.EntrySelectionID {
			expectedGot(t, scoredEntrySelection.EntrySelectionID, createdScoredEntrySelection.EntrySelectionID)
		}
		if createdScoredEntrySelection.StandingsID != scoredEntrySelection.StandingsID {
			expectedGot(t, scoredEntrySelection.StandingsID, createdScoredEntrySelection.StandingsID)
		}
		if !gocmp.Equal(createdScoredEntrySelection.Rankings, scoredEntrySelection.Rankings) {
			t.Fatal(gocmp.Diff(scoredEntrySelection.Rankings, createdScoredEntrySelection.Rankings))
		}
		if createdScoredEntrySelection.Score != scoredEntrySelection.Score {
			expectedGot(t, scoredEntrySelection.Score, createdScoredEntrySelection.Score)
		}
		if createdScoredEntrySelection.CreatedAt.Equal(emptyTime) {
			expectedNonEmpty(t, "CreatedAt")
		}
		if createdScoredEntrySelection.UpdatedAt.Valid {
			expectedEmpty(t, "UpdatedAt", createdScoredEntrySelection.UpdatedAt)
		}

		// inserting same scored entry selection a second time should fail
		_, err = agent.CreateScoredEntrySelection(ctx, scoredEntrySelection)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})

	t.Run("create invalid scored entry selection must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		missingEntrySelectionID := scoredEntrySelection
		missingEntrySelectionID.EntrySelectionID = uuid.UUID{}
		_, err := agent.CreateScoredEntrySelection(ctx, missingEntrySelectionID)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingStandingsID := scoredEntrySelection
		missingStandingsID.StandingsID = uuid.UUID{}
		_, err = agent.CreateScoredEntrySelection(ctx, missingStandingsID)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("create scored entry selection for non-existent entry selection or standings must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		invalidEntrySelectionID := scoredEntrySelection
		invalidEntrySelectionID.EntrySelectionID = nonExistentID
		_, err = agent.CreateScoredEntrySelection(ctx, invalidEntrySelectionID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		invalidStandingsID := scoredEntrySelection
		invalidStandingsID.StandingsID = nonExistentID
		_, err = agent.CreateScoredEntrySelection(ctx, invalidStandingsID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestScoredEntrySelectionAgent_UpdateScoredEntrySelection(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entrySelection := insertEntrySelection(t, generateTestEntrySelection(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))

	scoredEntrySelection := insertScoredEntrySelection(t, generateTestScoredEntrySelection(t, entrySelection.ID, standings.ID))

	agent := domain.ScoredEntrySelectionAgent{
		ScoredEntrySelectionAgentInjector: injector{db: db},
	}

	t.Run("update valid scored entry selection must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		changedScoredEntrySelection := scoredEntrySelection
		changedScoredEntrySelection.Score = 456
		changedScoredEntrySelection.Rankings = models.NewRankingWithScoreCollectionFromIDs([]string{"changedID_1", "changedID_2", "changedID_3"})

		updatedScoredEntrySelection, err := agent.UpdateScoredEntrySelection(ctx, changedScoredEntrySelection)
		if err != nil {
			t.Fatal(err)
		}

		if updatedScoredEntrySelection.EntrySelectionID != changedScoredEntrySelection.EntrySelectionID {
			expectedGot(t, changedScoredEntrySelection.EntrySelectionID, updatedScoredEntrySelection.EntrySelectionID)
		}
		if updatedScoredEntrySelection.StandingsID != changedScoredEntrySelection.StandingsID {
			expectedGot(t, changedScoredEntrySelection.StandingsID, updatedScoredEntrySelection.StandingsID)
		}
		if !gocmp.Equal(updatedScoredEntrySelection.Rankings, changedScoredEntrySelection.Rankings) {
			t.Fatal(gocmp.Diff(changedScoredEntrySelection.Rankings, updatedScoredEntrySelection.Rankings))
		}
		if updatedScoredEntrySelection.Score != changedScoredEntrySelection.Score {
			expectedGot(t, changedScoredEntrySelection.Score, updatedScoredEntrySelection.Score)
		}
		if !updatedScoredEntrySelection.CreatedAt.Equal(changedScoredEntrySelection.CreatedAt) {
			expectedGot(t, changedScoredEntrySelection.CreatedAt, updatedScoredEntrySelection.CreatedAt)
		}
		if !updatedScoredEntrySelection.UpdatedAt.Valid {
			expectedNonEmpty(t, "UpdatedAt")
		}
	})

	t.Run("update non-existent scored entry selection must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		nonExistentEntrySelectionID := scoredEntrySelection
		nonExistentEntrySelectionID.EntrySelectionID = nonExistentID
		_, err = agent.UpdateScoredEntrySelection(ctx, nonExistentEntrySelectionID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		nonExistentStandingsID := scoredEntrySelection
		nonExistentStandingsID.StandingsID = nonExistentID
		_, err = agent.UpdateScoredEntrySelection(ctx, nonExistentStandingsID)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestScoredEntrySelectionAgent_RetrieveScoredEntrySelectionByIDs(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t, "Harry Redknapp", "MrHarryR", "harry.redknapp@football.net"))
	entrySelection := insertEntrySelection(t, generateTestEntrySelection(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))

	now := time.Now().Truncate(time.Second)
	scoredEntrySelection := generateTestScoredEntrySelection(t, entrySelection.ID, standings.ID)
	scoredEntrySelection.CreatedAt = now
	scoredEntrySelection = insertScoredEntrySelection(t, scoredEntrySelection)

	agent := domain.ScoredEntrySelectionAgent{
		ScoredEntrySelectionAgentInjector: injector{db: db},
	}

	t.Run("retrieve existent scored entry selection must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedScoredEntrySelection, err := agent.RetrieveScoredEntrySelectionByIDs(ctx, entrySelection.ID.String(), standings.ID.String())
		if err != nil {
			t.Fatal(err)
		}

		if retrievedScoredEntrySelection.EntrySelectionID != scoredEntrySelection.EntrySelectionID {
			expectedGot(t, scoredEntrySelection.EntrySelectionID, retrievedScoredEntrySelection.EntrySelectionID)
		}
		if retrievedScoredEntrySelection.StandingsID != scoredEntrySelection.StandingsID {
			expectedGot(t, scoredEntrySelection.StandingsID, retrievedScoredEntrySelection.StandingsID)
		}
		if !gocmp.Equal(retrievedScoredEntrySelection.Rankings, scoredEntrySelection.Rankings) {
			t.Fatal(gocmp.Diff(scoredEntrySelection.Rankings, retrievedScoredEntrySelection.Rankings))
		}
		if retrievedScoredEntrySelection.Score != scoredEntrySelection.Score {
			expectedGot(t, scoredEntrySelection.Score, retrievedScoredEntrySelection.Score)
		}
		if !retrievedScoredEntrySelection.CreatedAt.In(utc).Equal(now.In(utc)) {
			expectedGot(t, now, retrievedScoredEntrySelection.CreatedAt)
		}
		if retrievedScoredEntrySelection.UpdatedAt.Valid {
			expectedEmpty(t, "UpdatedAt", retrievedScoredEntrySelection.UpdatedAt)
		}
	})

	t.Run("retrieve non-existent scored entry selection must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveScoredEntrySelectionByIDs(ctx, nonExistentID.String(), scoredEntrySelection.StandingsID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}

		_, err = agent.RetrieveScoredEntrySelectionByIDs(ctx, scoredEntrySelection.EntrySelectionID.String(), nonExistentID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}
