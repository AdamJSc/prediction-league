package domain_test

import (
	"context"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"testing"
	"time"
)

func TestTokenAgent_GenerateToken(t *testing.T) {
	defer truncate(t)

	agent := domain.TokenAgent{
		TokenAgentInjector: injector{db: db},
	}

	// arbitrary timestamp that isn't the current moment
	ts := time.Now().Add(-24 * time.Hour)

	t.Run("generate a token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetTimestampOnContext(ctx, ts)
		defer cancel()

		expectedType := 123
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
		if !token.IssuedAt.Equal(ts) {
			expectedGot(t, ts, token.IssuedAt)
		}
		expectedExpires := ts.Add(domain.TokenDurationInMinutes * time.Minute)
		if !token.ExpiresAt.Equal(expectedExpires) {
			expectedGot(t, expectedExpires, token.ExpiresAt)
		}

		// inserting same token a second time must fail
		err = repositories.NewTokenDatabaseRepository(db).Insert(ctx, token)
		if !cmp.ErrorType(err, repositories.DuplicateDBRecordError{})().Success() {
			expectedTypeOfGot(t, repositories.DuplicateDBRecordError{}, err)
		}
	})
}

func TestTokenAgent_RetrieveTokenByID(t *testing.T) {
	defer truncate(t)

	token := generateTestToken()
	tokenRepo := repositories.NewTokenDatabaseRepository(db)
	if err := tokenRepo.Insert(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	agent := domain.TokenAgent{
		TokenAgentInjector: injector{db: db},
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

	token := generateTestToken()
	tokenRepo := repositories.NewTokenDatabaseRepository(db)
	if err := tokenRepo.Insert(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	agent := domain.TokenAgent{
		TokenAgentInjector: injector{db: db},
	}

	t.Run("delete an existing token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		if err := agent.DeleteToken(ctx, *token); err != nil {
			t.Fatal(err)
		}

		// token must have been deleted
		err := tokenRepo.ExistsByID(ctx, token.ID)
		if !cmp.ErrorType(err, repositories.MissingDBRecordError{})().Success() {
			expectedTypeOfGot(t, repositories.MissingDBRecordError{}, err)
		}
	})
}

func TestTokenAgent_DeleteTokensExpiredAfter(t *testing.T) {
	defer truncate(t)

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

	var tokens = []*models.Token{
		token1,
		token2,
		token3,
	}

	tokenRepo := repositories.NewTokenDatabaseRepository(db)
	for _, token := range tokens {
		if err := tokenRepo.Insert(context.Background(), token); err != nil {
			t.Fatal(err)
		}
	}

	agent := domain.TokenAgent{
		TokenAgentInjector: injector{db: db},
	}

	t.Run("delete tokens expired after valid timestamp must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		if err := agent.DeleteTokensExpiredAfter(ctx, now); err != nil {
			t.Fatal(err)
		}

		// token 1 must have been deleted
		err := tokenRepo.ExistsByID(ctx, token1.ID)
		if !cmp.ErrorType(err, repositories.MissingDBRecordError{})().Success() {
			expectedTypeOfGot(t, repositories.MissingDBRecordError{}, err)
		}

		// token 2 must have been deleted
		err = tokenRepo.ExistsByID(ctx, token2.ID)
		if !cmp.ErrorType(err, repositories.MissingDBRecordError{})().Success() {
			expectedTypeOfGot(t, repositories.MissingDBRecordError{}, err)
		}

		// token 3 must have been deleted
		if err := tokenRepo.ExistsByID(ctx, token3.ID); err != nil {
			t.Fatal(err)
		}
	})
}

func generateTestToken() *models.Token {
	// arbitrary timestamp that isn't the current moment
	ts := time.Now().Truncate(time.Second).Add(-24 * time.Hour)

	return &models.Token{
		ID:        "token_id",
		Type:      123,
		Value:     "Hello World",
		IssuedAt:  ts,
		ExpiresAt: ts.Add(time.Minute),
	}
}
