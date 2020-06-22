package repositories

import (
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
)

// tokenDBFields defines the fields used regularly in Token-related transactions
var tokenDBFields = []string{
	"type",
	"value",
	"issued_at",
	"expires_at",
}

// TokenRepository defines the interface for transacting with our Token data source
type TokenRepository interface {
	Insert(ctx context.Context, token *models.Token) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.Token, error)
	ExistsByID(ctx context.Context, id string) error
	DeleteByID(ctx context.Context, id string) error
}

// TokenDatabaseRepository defines our DB-backed Token data store
type TokenDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new Token into the database
func (t TokenDatabaseRepository) Insert(ctx context.Context, token *models.Token) error {
	stmt := `INSERT INTO token (id, ` + getDBFieldsStringFromFields(tokenDBFields) + `)
					VALUES (?, ?, ?, ?, ?)`

	if _, err := t.agent.QueryContext(
		ctx,
		stmt,
		token.ID,
		token.Type,
		token.Value,
		token.IssuedAt,
		token.ExpiresAt,
	); err != nil {
		return wrapDBError(err)
	}

	return nil
}

// Select retrieves Tokens from our database based on the provided criteria
func (t TokenDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.Token, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(tokenDBFields) + ` FROM token ` + whereStmt

	rows, err := t.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}

	var tokens []models.Token
	for rows.Next() {
		token := models.Token{}

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
		return nil, MissingDBRecordError{errors.New("no tokens found")}
	}

	return tokens, nil
}

// ExistsByID determines whether a Token with the provided ID exists in the database
func (t TokenDatabaseRepository) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM token WHERE id = ?`

	row := t.agent.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{errors.New("token not found")}
	}

	return nil
}

// DeleteByID removes the Token with the provided ID from the database
func (t TokenDatabaseRepository) DeleteByID(ctx context.Context, id string) error {
	stmt := `DELETE FROM token WHERE id = ?`

	_, err := t.agent.QueryContext(ctx, stmt, id)
	if err != nil {
		return wrapDBError(err)
	}

	return nil
}

// NewTokenDatabaseRepository instantiates a new TokenDatabaseRepository with the provided DB agent
func NewTokenDatabaseRepository(db coresql.Agent) TokenDatabaseRepository {
	return TokenDatabaseRepository{agent: db}
}
