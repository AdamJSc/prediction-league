package domain

import (
	"context"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"math/rand"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"time"
)

const (
	TokenLength            = 32
	TokenDurationInMinutes = 20
)

// TokenAgentInjector defines the dependencies required by our TokenAgent
type TokenAgentInjector interface {
	MySQL() coresql.Agent
}

// TokenAgent defines the behaviours for handling Tokens
type TokenAgent struct{ TokenAgentInjector }

// GenerateToken generates a new unique token
func (t TokenAgent) GenerateToken(ctx context.Context, typ int, value string) (*models.Token, error) {
	tokenRepo := repositories.NewTokenDatabaseRepository(t.MySQL())

	id, err := generateUniqueTokenID(ctx, tokenRepo)
	if err != nil {
		return nil, err
	}

	now := TimestampFromContext(ctx)

	token := models.Token{
		ID:        id,
		Type:      typ,
		Value:     value,
		IssuedAt:  now,
		ExpiresAt: now.Add(TokenDurationInMinutes * time.Minute),
	}

	if err := tokenRepo.Insert(ctx, &token); err != nil {
		return nil, domainErrorFromDBError(err)
	}

	return &token, nil
}

// RetrieveTokenByID retrieves an existing token by the provided ID
func (t TokenAgent) RetrieveTokenByID(ctx context.Context, id string) (*models.Token, error) {
	tokenRepo := repositories.NewTokenDatabaseRepository(t.MySQL())

	tokens, err := tokenRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return nil, domainErrorFromDBError(err)
	}

	if len(tokens) != 1 {
		return nil, NotFoundError{fmt.Errorf("token id: %s not found", id)}
	}

	return &tokens[0], nil
}

// DeleteToken removes the provided token
func (t TokenAgent) DeleteToken(ctx context.Context, token models.Token) error {
	tokenRepo := repositories.NewTokenDatabaseRepository(t.MySQL())

	err := tokenRepo.DeleteByID(ctx, token.ID)
	if err != nil {
		return domainErrorFromDBError(err)
	}

	return nil
}

// DeleteTokensExpiredAfter removes tokens that have expired since the provide timestamp
func (t TokenAgent) DeleteTokensExpiredAfter(ctx context.Context, timestamp time.Time) error {
	tokenRepo := repositories.NewTokenDatabaseRepository(t.MySQL())

	tokens, err := tokenRepo.Select(ctx, map[string]interface{}{
		"expires_at": repositories.Condition{
			Operator: "<=",
			Operand:  timestamp,
		},
	}, false)
	if err != nil {
		return domainErrorFromDBError(err)
	}

	for _, token := range tokens {
		if err := tokenRepo.DeleteByID(ctx, token.ID); err != nil {
			return domainErrorFromDBError(err)
		}
	}

	return nil
}

// generateUniqueTokenID returns a string representing a unique token ID
func generateUniqueTokenID(ctx context.Context, tokenRepo repositories.TokenRepository) (string, error) {
	id := generateAlphaNumericString(TokenLength)

	if err := tokenRepo.ExistsByID(ctx, id); err != nil {
		switch err.(type) {
		case repositories.MissingDBRecordError:
			return id, nil
		default:
			return "", err
		}
	}

	return generateUniqueTokenID(ctx, tokenRepo)
}

// generateAlphaNumericString generates an alphanumeric string to the provided length
func generateAlphaNumericString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	source := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	var generated string

	sourceLen := len(source)

	for i := 0; i < length; i++ {
		randInt := r.Int63n(int64(sourceLen))
		randByte := []byte(source)[randInt]
		generated += string(randByte)
	}

	return generated
}
