package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	TokenTypeAuth = iota
	TokenTypeEntryRegistration
	TokenTypeMagicLogin
	TokenTypePrediction
	TokenLength = 32
)

var TokenValidityDuration = map[int]time.Duration{
	TokenTypeAuth:              time.Minute * 60,
	TokenTypeEntryRegistration: time.Minute * 10,
	TokenTypeMagicLogin:        time.Minute * 10,
	TokenTypePrediction:        time.Minute * 10,
}

// Token defines a token model
type Token struct {
	ID         string     `db:"id"`
	Type       int        `db:"type"`
	Value      string     `db:"value"`
	IssuedAt   time.Time  `db:"issued_at"`
	RedeemedAt *time.Time `db:"redeemed_at"`
	ExpiresAt  time.Time  `db:"expires_at"`
}

// TokenRepository defines the interface for transacting with our Token data source
type TokenRepository interface {
	Insert(ctx context.Context, token *Token) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]Token, error)
	Update(ctx context.Context, token *Token) error
	Delete(ctx context.Context, criteria map[string]interface{}, matchAny bool) (int64, error)
	ExistsByID(ctx context.Context, id string) error
	GenerateUniqueTokenID(ctx context.Context) (string, error)
}

// TokenAgent defines the behaviours for handling Tokens
type TokenAgent struct {
	tr TokenRepository
	cl Clock
	l  Logger
}

// GenerateToken generates a new unique token
func (t *TokenAgent) GenerateToken(ctx context.Context, typ int, value string) (*Token, error) {
	id, err := t.tr.GenerateUniqueTokenID(ctx)
	if err != nil {
		return nil, err
	}

	tokenDur, ok := TokenValidityDuration[typ]
	if !ok {
		return nil, NotFoundError{fmt.Errorf("token type %d has no validity duration", typ)}
	}

	now := t.cl.Now()

	token := Token{
		ID:        id,
		Type:      typ,
		Value:     value,
		IssuedAt:  now,
		ExpiresAt: now.Add(tokenDur),
	}

	if err := t.tr.Insert(ctx, &token); err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}

	return &token, nil
}

// RetrieveTokenByID retrieves an existing token by the provided ID
func (t *TokenAgent) RetrieveTokenByID(ctx context.Context, id string) (*Token, error) {
	tokens, err := t.tr.Select(ctx, map[string]interface{}{
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

// RedeemToken sets a RedeemedAt value on the provided Token and updates it
func (t *TokenAgent) RedeemToken(ctx context.Context, token Token) error {
	if token.RedeemedAt != nil {
		return ConflictError{errors.New("token has already been redeemed")}
	}

	redeemed := t.cl.Now()
	token.RedeemedAt = &redeemed

	err := t.tr.Update(ctx, &token)
	if err != nil {
		return domainErrorFromRepositoryError(err)
	}

	return nil
}

// DeleteToken removes the provided token
func (t *TokenAgent) DeleteToken(ctx context.Context, token Token) error {
	cnt, err := t.tr.Delete(ctx, map[string]interface{}{
		"id": token.ID,
	}, false)
	if err != nil {
		return domainErrorFromRepositoryError(err)
	}

	if cnt == 0 {
		return NotFoundError{errors.New("token not found")}
	}

	if cnt > 1 {
		return InternalError{fmt.Errorf("deleted %d tokens by id '%s'", cnt, token.ID)}
	}

	return nil
}

// DeleteTokensExpiredAfter removes tokens that have expired since the provide timestamp
func (t *TokenAgent) DeleteTokensExpiredAfter(ctx context.Context, timestamp time.Time) (int64, error) {
	cnt, err := t.tr.Delete(ctx, map[string]interface{}{
		"expires_at": DBQueryCondition{
			Operator: "<=",
			Operand:  timestamp,
		},
	}, false)
	if err != nil {
		return 0, domainErrorFromRepositoryError(err)
	}

	return cnt, nil
}

// DeleteInFlightTokens removes tokens that meet the provided criteria and have not yet been redeemed
func (t *TokenAgent) DeleteInFlightTokens(ctx context.Context, typ int, val string) (int64, error) {
	cnt, err := t.tr.Delete(ctx, map[string]interface{}{
		"type":        typ,
		"value":       val,
		"redeemed_at": DBQueryCondition{"IS NULL", nil},
	}, false)
	if err != nil {
		return 0, domainErrorFromRepositoryError(err)
	}

	return cnt, nil
}

// IsTokenValid determines whether the provided Token is valid
func (t *TokenAgent) IsTokenValid(tkn *Token, typ int, val string) bool {
	now := t.cl.Now()
	switch {
	case tkn.Type != typ:
		t.l.Errorf("token id '%s': token type %d is not %d", tkn.ID, tkn.Type, typ)
		return false
	case tkn.Value != val:
		t.l.Errorf("token id '%s': token value '%s' is not '%s'", tkn.ID, tkn.Value, val)
		return false
	case now.After(tkn.ExpiresAt):
		t.l.Errorf("token id '%s': expired", tkn.ID)
		return false
	}
	return true
}

// NewTokenAgent returns a new TokenAgent using the provided repository
func NewTokenAgent(tr TokenRepository, cl Clock, l Logger) (*TokenAgent, error) {
	switch {
	case tr == nil:
		return nil, fmt.Errorf("token repository: %w", ErrIsNil)
	case cl == nil:
		return nil, fmt.Errorf("clock: %w", ErrIsNil)
	case l == nil:
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	return &TokenAgent{tr, cl, l}, nil
}
