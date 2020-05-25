package domain_test

import (
	"github.com/LUSHDigital/uuid"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/domain"
	"reflect"
	"testing"
	"time"
)

func TestStandingsAgent_CreateStandings(t *testing.T) {
	defer truncate(t)

	agent := domain.StandingsAgent{
		StandingsAgentInjector: injector{db: db},
	}

	standings := generateTestStandings(t)

	t.Run("create valid standings must succeed", func(t *testing.T) {
		ctx, cancel := testContext(t)
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
		if createdStandings.CreatedAt.Equal(emptyTime) {
			expectedNonEmpty(t, "CreatedAt")
		}
		if createdStandings.UpdatedAt.Valid {
			expectedEmpty(t, "UpdatedAt", createdStandings.UpdatedAt)
		}
	})
}

func TestStandingsAgent_UpdateStandings(t *testing.T) {
	defer truncate(t)

	standings := insertStandings(t, generateTestStandings(t))

	agent := domain.StandingsAgent{
		StandingsAgentInjector: injector{db: db},
	}

	t.Run("update valid standings must succeed", func(t *testing.T) {
		ctx, cancel := testContext(t)
		defer cancel()

		changedStandings := standings
		changedStandings.RoundNumber = 2
		changedStandings.Rankings[0].Ranking.ID = "bonjour"
		changedStandings.Rankings[1].Ranking.ID = "monde"

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
		if !changedStandings.CreatedAt.Equal(updatedStandings.CreatedAt) {
			expectedGot(t, changedStandings.CreatedAt, updatedStandings.CreatedAt)
		}
		if !updatedStandings.UpdatedAt.Valid {
			expectedNonEmpty(t, "UpdatedAt")
		}
	})

	t.Run("update non-existent standings must fail", func(t *testing.T) {
		ctx, cancel := testContext(t)
		defer cancel()

		changedStandings := standings

		id, err := uuid.NewV4()
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

func TestStandingsAgent_RetrieveStandingsByID(t *testing.T) {
	defer truncate(t)

	standings := insertStandings(t, generateTestStandings(t))

	agent := domain.StandingsAgent{
		StandingsAgentInjector: injector{db: db},
	}

	t.Run("retrieve existent standings must succeed", func(t *testing.T) {
		ctx, cancel := testContext(t)
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
		if !standings.CreatedAt.Equal(retrievedStandings.CreatedAt) {
			expectedGot(t, standings.CreatedAt, retrievedStandings.CreatedAt)
		}
		if !standings.UpdatedAt.Time.Equal(retrievedStandings.UpdatedAt.Time) {
			expectedGot(t, standings.UpdatedAt.Time, retrievedStandings.UpdatedAt.Time)
		}
	})

	t.Run("retrieve non-existent standings must fail", func(t *testing.T) {
		ctx, cancel := testContext(t)
		defer cancel()

		nonExistentID, err := uuid.NewV4()
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

	agent := domain.StandingsAgent{
		StandingsAgentInjector: injector{db: db},
	}

	t.Run("retrieve existent standings must succeed", func(t *testing.T) {
		ctx, cancel := testContext(t)
		defer cancel()

		seasonID := reflect.ValueOf(datastore.Seasons).MapKeys()[0].String()

		retrievedStandings, err := agent.RetrieveStandingsBySeasonAndRoundNumber(ctx, seasonID, 1)
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
		if !standings2.CreatedAt.Equal(retrievedStandings.CreatedAt) {
			expectedGot(t, standings2.CreatedAt, retrievedStandings.CreatedAt)
		}
		if !standings2.UpdatedAt.Time.Equal(retrievedStandings.UpdatedAt.Time) {
			expectedGot(t, standings2.UpdatedAt.Time, retrievedStandings.UpdatedAt.Time)
		}
	})
}
