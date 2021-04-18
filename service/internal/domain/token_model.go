package domain

import (
	"time"
)

// Token defines a token model
type Token struct {
	ID        string    `db:"id"`
	Type      int       `db:"type"`
	Value     string    `db:"value"`
	IssuedAt  time.Time `db:"issued_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
