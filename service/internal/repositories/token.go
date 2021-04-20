package repositories

import (
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"golang.org/x/net/context"
	"math/rand"
	"prediction-league/service/internal/domain"
	"time"
)

const TokenLength = 32

// tokenDBFields defines the fields used regularly in Token-related transactions
var tokenDBFields = []string{
	"type",
	"value",
	"issued_at",
	"expires_at",
}

// TokenDatabaseRepository defines our DB-backed Token data store
type TokenDatabaseRepository struct {
	Agent coresql.Agent
}

// Insert inserts a new Token into the database
func (t *TokenDatabaseRepository) Insert(ctx context.Context, token *domain.Token) error {
	stmt := `INSERT INTO token (id, ` + getDBFieldsStringFromFields(tokenDBFields) + `)
					VALUES (?, ?, ?, ?, ?)`

	rows, err := t.Agent.QueryContext(
		ctx,
		stmt,
		token.ID,
		token.Type,
		token.Value,
		token.IssuedAt,
		token.ExpiresAt,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return nil
}

// Select retrieves Tokens from our database based on the provided criteria
func (t *TokenDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]domain.Token, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(tokenDBFields) + ` FROM token ` + whereStmt

	rows, err := t.Agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var tokens []domain.Token
	for rows.Next() {
		token := domain.Token{}

		if err := rows.Scan(
			&token.ID,
			&token.Type,
			&token.Value,
			&token.IssuedAt,
			&token.ExpiresAt,
		); err != nil {
			return nil, wrapDBError(err)
		}

		tokens = append(tokens, token)
	}

	if len(tokens) == 0 {
		return nil, domain.MissingDBRecordError{Err: errors.New("no tokens found")}
	}

	return tokens, nil
}

// ExistsByID determines whether a Token with the provided ID exists in the database
func (t *TokenDatabaseRepository) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM token WHERE id = ?`

	row := t.Agent.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return domain.MissingDBRecordError{Err: fmt.Errorf("token id %s: not found", id)}
	}

	return nil
}

// DeleteByID removes the Token with the provided ID from the database
func (t *TokenDatabaseRepository) DeleteByID(ctx context.Context, id string) error {
	stmt := `DELETE FROM token WHERE id = ?`

	rows, err := t.Agent.QueryContext(ctx, stmt, id)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return rows.Err()
}

// GenerateUniqueTokenID returns a string representing a unique token ID
func (t *TokenDatabaseRepository) GenerateUniqueTokenID(ctx context.Context) (string, error) {
	id := generateAlphaNumericString(TokenLength)

	if err := t.ExistsByID(ctx, id); err != nil {
		switch err.(type) {
		case domain.MissingDBRecordError:
			return id, nil
		default:
			return "", err
		}
	}

	return t.GenerateUniqueTokenID(ctx)
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
