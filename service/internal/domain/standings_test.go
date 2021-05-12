package domain_test

import (
	"errors"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestNewStandingsAgent(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		_, gotErr := domain.NewStandingsAgent(nil)
		if !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestStandingsAgent_CreateStandings(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewStandingsAgent(sr)
	if err != nil {
		t.Fatal(err)
	}

	standings := generateTestStandings(t)

	t.Run("create valid standings must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		createdStandings, err := agent.CreateStandings(ctx, standings)
		if err != nil {
			t.Fatal(err)
		}

		var emptyID uuid.UUID
		var emptyTime time.Time
		if createdStandings.ID == emptyID {
			expectedNonEmpty(t, "ID")
		}
		if !gocmp.Equal(createdStandings.Rankings, standings.Rankings) {
			t.Fatal(gocmp.Diff(standings.Rankings, createdStandings.Rankings))
		}
		if createdStandings.RoundNumber != standings.RoundNumber {
			expectedGot(t, standings.RoundNumber, createdStandings.RoundNumber)
		}
		if createdStandings.SeasonID != standings.SeasonID {
			expectedGot(t, standings.SeasonID, createdStandings.SeasonID)
		}
		if !createdStandings.Finalised {
			t.Fatal("expected finalised to be true, but it was not")
		}
		if createdStandings.CreatedAt.Equal(emptyTime) {
			expectedNonEmpty(t, "CreatedAt")
		}
		if createdStandings.UpdatedAt != nil {
			expectedEmpty(t, "UpdatedAt", createdStandings.UpdatedAt)
		}

		// inserting same standings a second time should fail
		_, err = agent.CreateStandings(ctx, standings)
		if !cmp.ErrorType(err, domain.ConflictError{})().Success() {
			expectedTypeOfGot(t, domain.ConflictError{}, err)
		}
	})
}

func TestStandingsAgent_UpdateStandings(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewStandingsAgent(sr)
	if err != nil {
		t.Fatal(err)
	}

	standings := insertStandings(t, generateTestStandings(t))

	t.Run("update valid standings must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		changedStandings := standings
		changedStandings.RoundNumber = 2
		changedStandings.Rankings[0].Ranking.ID = "bonjour"
		changedStandings.Rankings[1].Ranking.ID = "monde"
		changedStandings.Finalised = false

		updatedStandings, err := agent.UpdateStandings(ctx, changedStandings)
		if err != nil {
			t.Fatal(err)
		}

		if updatedStandings.ID != standings.ID {
			expectedGot(t, standings.ID, updatedStandings.ID)
		}
		if !gocmp.Equal(updatedStandings.Rankings, changedStandings.Rankings) {
			t.Fatal(gocmp.Diff(changedStandings.Rankings, updatedStandings.Rankings))
		}
		if updatedStandings.RoundNumber != changedStandings.RoundNumber {
			expectedGot(t, changedStandings.RoundNumber, updatedStandings.RoundNumber)
		}
		if updatedStandings.SeasonID != changedStandings.SeasonID {
			expectedGot(t, changedStandings.SeasonID, updatedStandings.SeasonID)
		}
		if updatedStandings.Finalised != changedStandings.Finalised {
			expectedGot(t, changedStandings.Finalised, updatedStandings.Finalised)
		}
		if !changedStandings.CreatedAt.Equal(updatedStandings.CreatedAt) {
			expectedGot(t, changedStandings.CreatedAt, updatedStandings.CreatedAt)
		}
		if updatedStandings.UpdatedAt == nil {
			expectedNonEmpty(t, "UpdatedAt")
		}
	})

	t.Run("update non-existent standings must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		changedStandings := standings

		id, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		changedStandings.ID = id

		_, err = agent.UpdateStandings(ctx, changedStandings)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

// TODO - tests for StandingsAgent.RetrieveStandingsIfNotFinalised (copy RetrieveStandingsBySeasonAndRoundNumber)

// TODO - tests for ValidateAndSortStandings

func TestStandingsAgent_RetrieveStandingsByID(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewStandingsAgent(sr)
	if err != nil {
		t.Fatal(err)
	}

	standings := insertStandings(t, generateTestStandings(t))

	t.Run("retrieve existent standings must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedStandings, err := agent.RetrieveStandingsByID(ctx, standings.ID.String())
		if err != nil {
			t.Fatal(err)
		}

		if retrievedStandings.ID != standings.ID {
			expectedGot(t, standings.ID, retrievedStandings.ID)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, standings.Rankings) {
			t.Fatal(gocmp.Diff(standings.Rankings, retrievedStandings.Rankings))
		}
		if retrievedStandings.RoundNumber != standings.RoundNumber {
			expectedGot(t, standings.RoundNumber, retrievedStandings.RoundNumber)
		}
		if retrievedStandings.SeasonID != standings.SeasonID {
			expectedGot(t, standings.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.Finalised != standings.Finalised {
			expectedGot(t, standings.Finalised, retrievedStandings.Finalised)
		}
		if !standings.CreatedAt.Equal(retrievedStandings.CreatedAt) {
			expectedGot(t, standings.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, standings.UpdatedAt, retrievedStandings.UpdatedAt)
	})

	t.Run("retrieve non-existent standings must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		_, err = agent.RetrieveStandingsByID(ctx, nonExistentID.String())
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestStandingsAgent_RetrieveStandingsBySeasonAndRoundNumber(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewStandingsAgent(sr)
	if err != nil {
		t.Fatal(err)
	}

	// season ID won't match our method parameters, so this won't be returned
	standings1 := generateTestStandings(t)
	standings1.SeasonID = "nnnnnn"
	standings1 = insertStandings(t, standings1)

	// this will be returned by our agent method
	standings2 := generateTestStandings(t)
	standings2 = insertStandings(t, standings2)

	// round number won't match our method parameters, so this won't be returned
	standings3 := generateTestStandings(t)
	standings3.RoundNumber = 2
	standings3 = insertStandings(t, standings3)

	t.Run("retrieve existent standings must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		seasonID := testSeason.ID
		roundNumber := 1

		retrievedStandings, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, seasonID, roundNumber)
		if err != nil {
			t.Fatal(err)
		}

		if retrievedStandings.ID != standings2.ID {
			expectedGot(t, standings2.ID, retrievedStandings.ID)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, standings2.Rankings) {
			t.Fatal(gocmp.Diff(standings2.Rankings, retrievedStandings.Rankings))
		}
		if retrievedStandings.RoundNumber != standings2.RoundNumber {
			expectedGot(t, standings2.RoundNumber, retrievedStandings.RoundNumber)
		}
		if retrievedStandings.SeasonID != standings2.SeasonID {
			expectedGot(t, standings2.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.Finalised != standings2.Finalised {
			expectedGot(t, standings2.Finalised, retrievedStandings.Finalised)
		}
		if !standings2.CreatedAt.Equal(retrievedStandings.CreatedAt) {
			expectedGot(t, standings2.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, standings2.UpdatedAt, retrievedStandings.UpdatedAt)
	})

	t.Run("retrieving non-existent standings must provided the expected error", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		if _, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, "non-existent", 1); !errors.As(err, &domain.NotFoundError{}) {
			t.Fatalf("want not found error, got %s (%T)", err.Error(), err)
		}

		if _, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, testSeason.ID, 1234); !errors.As(err, &domain.NotFoundError{}) {
			t.Fatalf("want not found error, got %s (%T)", err.Error(), err)
		}
	})
}

func TestStandingsAgent_RetrieveLatestStandingsBySeasonIDAndTimestamp(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewStandingsAgent(sr)
	if err != nil {
		t.Fatal(err)
	}

	baseDate := time.Now().Truncate(time.Second)
	baseDateOneHour := baseDate.Add(time.Hour)

	// season ID won't match our method parameters, so this won't be returned
	standings1 := generateTestStandings(t)
	standings1.SeasonID = "nnnnnn"
	standings1 = insertStandings(t, standings1)
	standings1.UpdatedAt = &baseDateOneHour
	standings1 = updateStandings(t, standings1)

	// this will be returned by our agent method
	standings2 := generateTestStandings(t)
	standings2.RoundNumber = 2
	standings2.CreatedAt = baseDate
	standings2 = insertStandings(t, standings2)
	standings2.UpdatedAt = &baseDateOneHour
	standings2 = updateStandings(t, standings2)

	// this will be returned by our agent method
	standings3 := generateTestStandings(t)
	standings3.RoundNumber = 3
	standings3.CreatedAt = baseDate.Add(2 * time.Hour)
	standings3 = insertStandings(t, standings3)

	seasonID := standings2.SeasonID

	t.Run("retrieve latest standings by timestamp that occurs before first expected standings created date must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, standings2.CreatedAt.Add(-time.Nanosecond))
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("retrieve latest standings by timestamp that occurs on first expected standings created date must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedStandings, err := agent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, standings2.CreatedAt)
		if err != nil {
			t.Fatal(err)
		}

		expectedStandings := standings2

		if retrievedStandings.ID != expectedStandings.ID {
			expectedGot(t, expectedStandings.ID, retrievedStandings.ID)
		}
		if retrievedStandings.SeasonID != expectedStandings.SeasonID {
			expectedGot(t, expectedStandings.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.RoundNumber != expectedStandings.RoundNumber {
			expectedGot(t, expectedStandings.RoundNumber, retrievedStandings.RoundNumber)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, expectedStandings.Rankings) {
			expectedGot(t, expectedStandings.Rankings, retrievedStandings.Rankings)
		}
		if retrievedStandings.Finalised != expectedStandings.Finalised {
			expectedGot(t, expectedStandings.Finalised, retrievedStandings.Finalised)
		}
		if !retrievedStandings.CreatedAt.Equal(expectedStandings.CreatedAt) {
			expectedGot(t, expectedStandings.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, expectedStandings.UpdatedAt, retrievedStandings.UpdatedAt)
	})

	t.Run("retrieve latest standings by timestamp that occurs before second expected standings created date must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedStandings, err := agent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, standings3.CreatedAt.Add(-time.Nanosecond))
		if err != nil {
			t.Fatal(err)
		}

		expectedStandings := standings2

		if retrievedStandings.ID != expectedStandings.ID {
			expectedGot(t, expectedStandings.ID, retrievedStandings.ID)
		}
		if retrievedStandings.SeasonID != expectedStandings.SeasonID {
			expectedGot(t, expectedStandings.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.RoundNumber != expectedStandings.RoundNumber {
			expectedGot(t, expectedStandings.RoundNumber, retrievedStandings.RoundNumber)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, expectedStandings.Rankings) {
			expectedGot(t, expectedStandings.Rankings, retrievedStandings.Rankings)
		}
		if retrievedStandings.Finalised != expectedStandings.Finalised {
			expectedGot(t, expectedStandings.Finalised, retrievedStandings.Finalised)
		}
		if !retrievedStandings.CreatedAt.Equal(expectedStandings.CreatedAt) {
			expectedGot(t, expectedStandings.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, expectedStandings.UpdatedAt, retrievedStandings.UpdatedAt)
	})

	t.Run("retrieve latest standings by timestamp that occurs on second expected standings created date must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedStandings, err := agent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, standings3.CreatedAt)
		if err != nil {
			t.Fatal(err)
		}

		expectedStandings := standings3

		if retrievedStandings.ID != expectedStandings.ID {
			expectedGot(t, expectedStandings.ID, retrievedStandings.ID)
		}
		if retrievedStandings.SeasonID != expectedStandings.SeasonID {
			expectedGot(t, expectedStandings.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.RoundNumber != expectedStandings.RoundNumber {
			expectedGot(t, expectedStandings.RoundNumber, retrievedStandings.RoundNumber)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, expectedStandings.Rankings) {
			expectedGot(t, expectedStandings.Rankings, retrievedStandings.Rankings)
		}
		if retrievedStandings.Finalised != expectedStandings.Finalised {
			expectedGot(t, expectedStandings.Finalised, retrievedStandings.Finalised)
		}
		if !retrievedStandings.CreatedAt.Equal(expectedStandings.CreatedAt) {
			expectedGot(t, expectedStandings.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, expectedStandings.UpdatedAt, retrievedStandings.UpdatedAt)
	})

	t.Run("retrieve latest standings by timestamp that occurs just after second expected standings created date must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedStandings, err := agent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, standings3.CreatedAt.Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}

		expectedStandings := standings3

		if retrievedStandings.ID != expectedStandings.ID {
			expectedGot(t, expectedStandings.ID, retrievedStandings.ID)
		}
		if retrievedStandings.SeasonID != expectedStandings.SeasonID {
			expectedGot(t, expectedStandings.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.RoundNumber != expectedStandings.RoundNumber {
			expectedGot(t, expectedStandings.RoundNumber, retrievedStandings.RoundNumber)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, expectedStandings.Rankings) {
			expectedGot(t, expectedStandings.Rankings, retrievedStandings.Rankings)
		}
		if retrievedStandings.Finalised != expectedStandings.Finalised {
			expectedGot(t, expectedStandings.Finalised, retrievedStandings.Finalised)
		}
		if !retrievedStandings.CreatedAt.Equal(expectedStandings.CreatedAt) {
			expectedGot(t, expectedStandings.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, expectedStandings.UpdatedAt, retrievedStandings.UpdatedAt)
	})

	t.Run("retrieve latest standings by timestamp that occurs significantly after second expected standings created date must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedStandings, err := agent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, standings3.CreatedAt.Add(10000*time.Hour))
		if err != nil {
			t.Fatal(err)
		}

		expectedStandings := standings3

		if retrievedStandings.ID != expectedStandings.ID {
			expectedGot(t, expectedStandings.ID, retrievedStandings.ID)
		}
		if retrievedStandings.SeasonID != expectedStandings.SeasonID {
			expectedGot(t, expectedStandings.SeasonID, retrievedStandings.SeasonID)
		}
		if retrievedStandings.RoundNumber != expectedStandings.RoundNumber {
			expectedGot(t, expectedStandings.RoundNumber, retrievedStandings.RoundNumber)
		}
		if !gocmp.Equal(retrievedStandings.Rankings, expectedStandings.Rankings) {
			expectedGot(t, expectedStandings.Rankings, retrievedStandings.Rankings)
		}
		if retrievedStandings.Finalised != expectedStandings.Finalised {
			expectedGot(t, expectedStandings.Finalised, retrievedStandings.Finalised)
		}
		if !retrievedStandings.CreatedAt.Equal(expectedStandings.CreatedAt) {
			expectedGot(t, expectedStandings.CreatedAt, retrievedStandings.CreatedAt)
		}
		checkTimePtrMatch(t, expectedStandings.UpdatedAt, retrievedStandings.UpdatedAt)
	})
}
