package models

import (
	"time"
)

const (
	tokenDuration = 20 * time.Minute

	TokenTypeAuthToken = iota
)

// Token defines a token model
type Token struct {
	ID       string    `db:"id"`
	Type     int       `db:"type"`
	Value    string    `db:"value"`
	IssuedAt time.Time `db:"issued_at"`
	Expires  time.Time `db:"updated_at"`
}

// NewToken generates a new token
func NewToken(id string, typ int, value string) *Token {
	now := time.Now().Truncate(time.Second)
	expires := now.Add(tokenDuration)

	return &Token{
		ID:       id,
		Type:     typ,
		Value:    value,
		IssuedAt: now,
		Expires:  expires,
	}
}
