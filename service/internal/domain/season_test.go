package domain_test

import (
	"context"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"gotest.tools/assert/cmp"
	"os"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestSeasonAgent_CreateSeason(t *testing.T) {
	defer truncate(t)

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("ADMIN_BASIC_AUTH", "user:123456789")
	ctx := context.WithValue(context.Background(), "ADMIN_BASIC_AUTH", "user:123456789")

	agent := domain.SeasonAgent{SeasonAgentInjector: injector{db: db}}

	t.Run("creating a valid season with valid credentials must succeed", func(t *testing.T) {
		name := "My Season"
		entriesFrom := time.Date(1992, 7, 1, 0, 0, 0, 0, loc)
		startDate := time.Date(1992, 8, 15, 15, 0, 0, 0, loc)
		endDate := time.Date(1993, 5, 11, 23, 59, 59, 0, loc)

		s := domain.Season{
			Name:        name,
			EntriesFrom: entriesFrom,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		// should succeed
		if err := agent.CreateSeason(ctx, &s, 0); err != nil {
			t.Fatal(err)
		}

		// check raw values that shouldn't have changed
		if !cmp.Equal(name, s.Name)().Success() {
			expectedGot(t, name, s.Name)
		}
		if !cmp.Equal(entriesFrom, s.EntriesFrom)().Success() {
			expectedGot(t, entriesFrom, s.EntriesFrom)
		}
		if !cmp.Equal(startDate, s.StartDate)().Success() {
			expectedGot(t, startDate, s.StartDate)
		}
		if !cmp.Equal(endDate, s.EndDate)().Success() {
			expectedGot(t, endDate, s.EndDate)
		}

		// check sanitised values
		expectedID := "199293_1"
		expectedEntriesUntil := sqltypes.ToNullTime(s.StartDate)

		if !cmp.Equal(expectedID, s.ID)().Success() {
			expectedGot(t, expectedID, s.ID)
		}
		if !cmp.Equal(expectedEntriesUntil, s.EntriesUntil)().Success() {
			expectedGot(t, expectedEntriesUntil, s.EntriesUntil)
		}
		if cmp.Equal(time.Time{}, s.CreatedAt)().Success() {
			expectedNonEmpty(t, "time")
		}
		if !cmp.Equal(sqltypes.NullTime{}, s.UpdatedAt)().Success() {
			expectedEmpty(t, "nulltime", s.UpdatedAt)
		}
	})

	t.Run("creating a season with invalid credentials must fail", func(t *testing.T) {
		ctxWithInvalidCredentials := context.WithValue(context.Background(), "ADMIN_BASIC_AUTH", "not_valid_basic_auth_credentials")

		err = agent.CreateSeason(ctxWithInvalidCredentials, &domain.Season{}, 0)
		if !cmp.ErrorType(err, domain.UnauthorizedError{})().Success() {
			expectedTypeOfGot(t, domain.UnauthorizedError{}, err)
		}
	})

	t.Run("creating a season with missing required values must fail", func(t *testing.T) {
		var s domain.Season
		var err error

		// missing name
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing entries_from
		s = domain.Season{
			Name: "My Season",
		}
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing start_date
		s = domain.Season{
			Name:        "My Season",
			EntriesFrom: time.Now(),
		}
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		// missing end_date
		s = domain.Season{
			Name:        "My Season",
			EntriesFrom: time.Now(),
			StartDate:   time.Now(),
		}
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("creating a season with an end date before start date must fail", func(t *testing.T) {
		startDate := time.Now()
		s := domain.Season{
			Name:        "My Season",
			EntriesFrom: time.Now(),
			StartDate:   startDate,
			EndDate:     startDate.Add(-24 * time.Hour),
		}
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("creating a season with an entries until date before an entries from date must fail", func(t *testing.T) {
		entriesFrom := time.Now()
		s := domain.Season{
			Name:         "My Season",
			EntriesFrom:  entriesFrom,
			EntriesUntil: sqltypes.ToNullTime(entriesFrom.Add(-24 * time.Hour)),
			StartDate:    time.Now(),
			EndDate:      time.Now(),
		}
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("creating a season with a start date before entries until date must fail", func(t *testing.T) {
		entriesUntil := sqltypes.ToNullTime(time.Now())
		s := domain.Season{
			Name:         "My Season",
			EntriesFrom:  time.Now(),
			EntriesUntil: entriesUntil,
			StartDate:    entriesUntil.Time.Add(-24 * time.Hour),
			EndDate:      time.Now(),
		}
		err = agent.CreateSeason(ctx, &s, 0)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})
}
