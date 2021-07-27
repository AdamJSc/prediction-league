package mysqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"

	"golang.org/x/net/context"
)

// tokenDBFields defines the fields used regularly in Token-related transactions
var tokenDBFields = []string{
	"type",
	"value",
	"issued_at",
	"redeemed_at",
	"expires_at",
}

// TokenRepo defines our DB-backed Token data store
type TokenRepo struct {
	db *sql.DB
}

// Insert inserts a new Token into the database
func (t *TokenRepo) Insert(ctx context.Context, token *domain.Token) error {
	stmt := `INSERT INTO token (id, ` + getDBFieldsStringFromFields(tokenDBFields) + `)
					VALUES (?, ?, ?, ?, ?, ?)`

	rows, err := t.db.QueryContext(
		ctx,
		stmt,
		token.ID,
		token.Type,
		token.Value,
		token.IssuedAt,
		token.RedeemedAt,
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

	rows, err := t.db.QueryContext(ctx, stmt, params...)
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
			&token.RedeemedAt,
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

// Update updates an existing Token in the database
func (t *TokenRepo) Update(ctx context.Context, token *domain.Token) error {
	stmt := `UPDATE token
			SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(tokenDBFields) + `
			WHERE id = ?`

	res, err := t.db.Exec(
		stmt,
		token.Type,
		token.Value,
		token.IssuedAt,
		token.RedeemedAt,
		token.ExpiresAt,
		token.ID,
	)
	if err != nil {
		return wrapDBError(err)
	}

	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return domain.MissingDBRecordError{Err: errors.New("token does not exist")}
	}

	return nil
}

// Delete deletes Tokens from our database based on the provided criteria
func (t *TokenRepo) Delete(ctx context.Context, criteria map[string]interface{}, matchAny bool) (int64, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `DELETE FROM token ` + whereStmt

	res, err := t.db.Exec(stmt, params...)
	if err != nil {
		return 0, wrapDBError(err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return 0, wrapDBError(err)
	}

	return cnt, nil
}

// ExistsByID determines whether a Token with the provided ID exists in the database
func (t *TokenRepo) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM token WHERE id = ?`

	row := t.db.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return domain.MissingDBRecordError{Err: fmt.Errorf("token id %s: not found", id)}
	}

	return nil
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
func NewTokenRepo(db *sql.DB) (*TokenRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db: %w", domain.ErrIsNil)
	}
	return &TokenRepo{db: db}, nil
}
