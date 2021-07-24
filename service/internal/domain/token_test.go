package domain_test

import (
	"context"
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	"gotest.tools/assert/cmp"
)

func TestNewTokenAgent(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		cl := &mockClock{}
		l := &mockLogger{}

		tt := []struct {
			tr      domain.TokenRepository
			cl      domain.Clock
			l       domain.Logger
			wantErr error
		}{
			{nil, cl, l, domain.ErrIsNil},
			{tr, nil, l, domain.ErrIsNil},
			{tr, cl, nil, domain.ErrIsNil},
			{tr, cl, l, nil},
		}
		for idx, tc := range tt {
			agent, gotErr := domain.NewTokenAgent(tc.tr, tc.cl, nil)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && agent == nil {
				t.Fatalf("tc #%d: want non-empty agent, got nil", idx)
			}
		}
	})
}

func TestTokenAgent_GenerateToken(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{t: dt}, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("generate an auth token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedType := domain.TokenTypeAuth
		expectedValue := "Hello World"

		token, err := agent.GenerateToken(ctx, expectedType, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		if len(token.ID) != domain.TokenLength {
			expectedGot(t, domain.TokenLength, len(token.ID))
		}
		if token.Type != expectedType {
			expectedGot(t, expectedType, token.Type)
		}
		if token.Value != expectedValue {
			expectedGot(t, expectedValue, token.Value)
		}
		if !token.IssuedAt.Equal(dt) {
			expectedGot(t, dt, token.IssuedAt)
		}
		expectedExpires := dt.Add(domain.TokenValidityDuration[expectedType])
		if !token.ExpiresAt.Equal(expectedExpires) {
			expectedGot(t, expectedExpires, token.ExpiresAt)
		}

		// inserting same token a second time must fail
		err = tr.Insert(ctx, token)
		if !cmp.ErrorType(err, domain.DuplicateDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.DuplicateDBRecordError{}, err)
		}
	})

	t.Run("generate a short code reset token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedType := domain.TokenTypeMagicLogin
		expectedValue := "Hello World"

		token, err := agent.GenerateToken(ctx, expectedType, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		if len(token.ID) != domain.TokenLength {
			expectedGot(t, domain.TokenLength, len(token.ID))
		}
		if token.Type != expectedType {
			expectedGot(t, expectedType, token.Type)
		}
		if token.Value != expectedValue {
			expectedGot(t, expectedValue, token.Value)
		}
		if !token.IssuedAt.Equal(dt) {
			expectedGot(t, dt, token.IssuedAt)
		}
		expectedExpires := dt.Add(domain.TokenValidityDuration[expectedType])
		if !token.ExpiresAt.Equal(expectedExpires) {
			expectedGot(t, expectedExpires, token.ExpiresAt)
		}

		// inserting same token a second time must fail
		err = tr.Insert(ctx, token)
		if !cmp.ErrorType(err, domain.DuplicateDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.DuplicateDBRecordError{}, err)
		}
	})

	t.Run("generate a token of a non-existent token type must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentTokenType := 123456

		_, err := agent.GenerateToken(ctx, nonExistentTokenType, "Hello World")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestTokenAgent_RetrieveTokenByID(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	token := generateTestToken()
	if err := tr.Insert(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve an existing token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedToken, err := agent.RetrieveTokenByID(ctx, token.ID)
		if err != nil {
			t.Fatal(err)
		}

		if retrievedToken.ID != token.ID {
			expectedGot(t, token.ID, retrievedToken.ID)
		}
		if retrievedToken.Type != token.Type {
			expectedGot(t, token.Type, retrievedToken.Type)
		}
		if retrievedToken.Value != token.Value {
			expectedGot(t, token.Value, retrievedToken.Value)
		}
		if !retrievedToken.IssuedAt.Equal(token.IssuedAt) {
			expectedGot(t, token.IssuedAt, retrievedToken.IssuedAt)
		}
		if !retrievedToken.ExpiresAt.Equal(token.ExpiresAt) {
			expectedGot(t, token.ExpiresAt, retrievedToken.ExpiresAt)
		}
	})

	t.Run("retrieve a non-existent token must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveTokenByID(ctx, "non_existent_id")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestTokenAgent_DeleteToken(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	token := generateTestToken()
	if err := tr.Insert(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	t.Run("delete an existing token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		if err := agent.DeleteToken(ctx, *token); err != nil {
			t.Fatal(err)
		}

		// token must have been deleted
		err := tr.ExistsByID(ctx, token.ID)
		if !cmp.ErrorType(err, domain.MissingDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.MissingDBRecordError{}, err)
		}
	})
}

func TestTokenAgent_DeleteTokensExpiredAfter(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Truncate(time.Second)

	// token 1 represents a token that expires 1 second in the past
	token1 := generateTestToken()
	token1.ID = "test_token_1"
	token1.ExpiresAt = now.Add(-time.Second)

	// token 2 represents a token that expires now
	token2 := generateTestToken()
	token2.ID = "test_token_2"
	token2.ExpiresAt = now

	// token 3 represents a token that expires 1 second in the future
	token3 := generateTestToken()
	token3.ID = "test_token_3"
	token3.ExpiresAt = now.Add(time.Second)

	var tokens = []*domain.Token{
		token1,
		token2,
		token3,
	}

	for _, token := range tokens {
		if err := tr.Insert(context.Background(), token); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("delete tokens expired after valid timestamp must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		if err := agent.DeleteTokensExpiredAfter(ctx, now); err != nil {
			t.Fatal(err)
		}

		// token 1 must have been deleted
		err := tr.ExistsByID(ctx, token1.ID)
		if !cmp.ErrorType(err, domain.MissingDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.MissingDBRecordError{}, err)
		}

		// token 2 must have been deleted
		err = tr.ExistsByID(ctx, token2.ID)
		if !cmp.ErrorType(err, domain.MissingDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.MissingDBRecordError{}, err)
		}

		// token 3 must have been deleted
		if err := tr.ExistsByID(ctx, token3.ID); err != nil {
			t.Fatal(err)
		}
	})
}

// TODO - ShortCode: tests for IsTokenValid

func generateTestToken() *domain.Token {
	// arbitrary timestamp that isn't the current moment
	ts := time.Now().Truncate(time.Second).Add(-24 * time.Hour)

	return &domain.Token{
		ID:        "token_id",
		Type:      123,
		Value:     "Hello World",
		IssuedAt:  ts,
		ExpiresAt: ts.Add(time.Minute),
	}
}
