package mysqldb

import (
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"golang.org/x/net/context"
	"prediction-league/service/internal/domain"
)

// tokenDBFields defines the fields used regularly in Token-related transactions
var tokenDBFields = []string{
	"type",
	"value",
	"issued_at",
	"expires_at",
}

// TokenRepo defines our DB-backed Token data store
type TokenRepo struct {
	Agent coresql.Agent
}

// Insert inserts a new Token into the database
func (t *TokenRepo) Insert(ctx context.Context, token *domain.Token) error {
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
func (t *TokenRepo) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]domain.Token, error) {
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
func (t *TokenRepo) ExistsByID(ctx context.Context, id string) error {
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
func (t *TokenRepo) DeleteByID(ctx context.Context, id string) error {
	stmt := `DELETE FROM token WHERE id = ?`

	rows, err := t.Agent.QueryContext(ctx, stmt, id)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	return rows.Err()
}

// GenerateUniqueTokenID returns a string representing a unique token ID
func (t *TokenRepo) GenerateUniqueTokenID(ctx context.Context) (string, error) {
	id := generateAlphaNumericString(domain.TokenLength)

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

// NewTokenRepo instantiates a new TokenRepo with the provided DB agent
func NewTokenRepo(db coresql.Agent) *TokenRepo {
	return &TokenRepo{Agent: db}
}
