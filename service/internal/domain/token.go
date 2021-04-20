package domain

import (
	"context"
	"fmt"
	"time"
)

const (
	TokenTypeAuth = iota
	TokenTypeShortCodeResetToken
)

var TokenValidityDuration = map[int]time.Duration{
	TokenTypeAuth:                time.Minute * 60,
	TokenTypeShortCodeResetToken: time.Minute * 10,
}

// TokenRepository defines the interface for transacting with our Token data source
type TokenRepository interface {
	Insert(ctx context.Context, token *Token) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]Token, error)
	ExistsByID(ctx context.Context, id string) error
	DeleteByID(ctx context.Context, id string) error
	GenerateUniqueTokenID(ctx context.Context) (string, error)
}

// TokenAgentInjector defines the dependencies required by our TokenAgent
type TokenAgentInjector interface {
	TokenRepo() TokenRepository
}

// TokenAgent defines the behaviours for handling Tokens
type TokenAgent struct{ TokenAgentInjector }

// GenerateToken generates a new unique token
func (t *TokenAgent) GenerateToken(ctx context.Context, typ int, value string) (*Token, error) {
	tokenRepo := t.TokenRepo()

	id, err := tokenRepo.GenerateUniqueTokenID(ctx)
	if err != nil {
		return nil, err
	}

	tokenDur, ok := TokenValidityDuration[typ]
	if !ok {
		return nil, NotFoundError{fmt.Errorf("token type %d has no validity duration", typ)}
	}

	now := TimestampFromContext(ctx)

	token := Token{
		ID:        id,
		Type:      typ,
		Value:     value,
		IssuedAt:  now,
		ExpiresAt: now.Add(tokenDur),
	}

	if err := tokenRepo.Insert(ctx, &token); err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	return &token, nil
}

// RetrieveTokenByID retrieves an existing token by the provided ID
func (t *TokenAgent) RetrieveTokenByID(ctx context.Context, id string) (*Token, error) {
	tokenRepo := t.TokenRepo()

	tokens, err := tokenRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	if len(tokens) != 1 {
		return nil, NotFoundError{fmt.Errorf("token id: %s not found", id)}
	}

	return &tokens[0], nil
}

// DeleteToken removes the provided token
func (t *TokenAgent) DeleteToken(ctx context.Context, token Token) error {
	tokenRepo := t.TokenRepo()

	err := tokenRepo.DeleteByID(ctx, token.ID)
	if err != nil {
		return domainErrorFromRepositoryError(err)
	}

	return nil
}

// DeleteTokensExpiredAfter removes tokens that have expired since the provide timestamp
func (t *TokenAgent) DeleteTokensExpiredAfter(ctx context.Context, timestamp time.Time) error {
	tokenRepo := t.TokenRepo()

	tokens, err := tokenRepo.Select(ctx, map[string]interface{}{
		"expires_at": DBQueryCondition{
			Operator: "<=",
			Operand:  timestamp,
		},
	}, false)
	if err != nil {
		return domainErrorFromRepositoryError(err)
	}

	for _, token := range tokens {
		if err := tokenRepo.DeleteByID(ctx, token.ID); err != nil {
			return domainErrorFromRepositoryError(err)
		}
	}

	return nil
}
