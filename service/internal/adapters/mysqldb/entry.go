package mysqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"math/rand"
	"prediction-league/service/internal/domain"
	"time"
)

// shortCodeLength represents the number of characters that a short code will contain
const shortCodeLength = 6

// entryDBFields defines the fields used regularly in Entry-related transactions
var entryDBFields = []string{
	"short_code",
	"season_id",
	"realm_name",
	"entrant_name",
	"entrant_nickname",
	"entrant_email",
	"status",
	"payment_method",
	"payment_ref",
	"approved_at",
}

// EntryRepo defines our DB-backed Entry data store
type EntryRepo struct {
	db *sql.DB
}

// Insert inserts a new Entry into the database
func (e *EntryRepo) Insert(ctx context.Context, entry *domain.Entry) error {
	stmt := `INSERT INTO entry (id, ` + getDBFieldsStringFromFields(entryDBFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	rows, err := e.db.QueryContext(
		ctx,
		stmt,
		entry.ID,
		entry.ShortCode,
		entry.SeasonID,
		entry.RealmName,
		entry.EntrantName,
		entry.EntrantNickname,
		entry.EntrantEmail,
		entry.Status,
		entry.PaymentMethod,
		entry.PaymentRef,
		entry.ApprovedAt,
		now,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	entry.CreatedAt = now

	return nil
}

// Update updates an existing Entry in the database
func (e *EntryRepo) Update(ctx context.Context, entry *domain.Entry) error {
	stmt := `UPDATE entry
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(entryDBFields) + `, updated_at = ?
				WHERE id = ?`

	now := time.Now().Truncate(time.Second)

	rows, err := e.db.QueryContext(
		ctx,
		stmt,
		entry.ShortCode,
		entry.SeasonID,
		entry.RealmName,
		entry.EntrantName,
		entry.EntrantNickname,
		entry.EntrantEmail,
		entry.Status,
		entry.PaymentMethod,
		entry.PaymentRef,
		entry.ApprovedAt,
		now,
		entry.ID,
	)
	if err != nil {
		return wrapDBError(err)
	}
	defer rows.Close()

	entry.UpdatedAt = &now

	return nil
}

// Select retrieves Entries from our database based on the provided criteria
func (e *EntryRepo) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]domain.Entry, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entryDBFields) + `, created_at, updated_at FROM entry ` + whereStmt

	rows, err := e.db.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var entries []domain.Entry
	for rows.Next() {
		entry := domain.Entry{}

		if err := rows.Scan(
			&entry.ID,
			&entry.ShortCode,
			&entry.SeasonID,
			&entry.RealmName,
			&entry.EntrantName,
			&entry.EntrantNickname,
			&entry.EntrantEmail,
			&entry.Status,
			&entry.PaymentMethod,
			&entry.PaymentRef,
			&entry.ApprovedAt,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, wrapDBError(err)
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, domain.MissingDBRecordError{Err: errors.New("no entries found")}
	}

	return entries, nil
}

// SelectBySeasonIDAndApproved retrieves Entries that match the provided criteria
func (e *EntryRepo) SelectBySeasonIDAndApproved(ctx context.Context, seasonID string, approved bool) ([]domain.Entry, error) {
	// retrieve entries by season and round number
	criteria := map[string]interface{}{
		"season_id": seasonID,
	}
	if approved {
		criteria["approved_at"] = domain.DBQueryCondition{
			Operator: "IS NOT NULL",
		}
	} else {
		criteria["approved_at"] = domain.DBQueryCondition{
			Operator: "IS NULL",
		}
	}
	return e.Select(ctx, criteria, false)
}

// ExistsByID determines whether an Entry with the provided ID exists in the database
func (e *EntryRepo) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM entry WHERE id = ?`

	row := e.db.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return domain.MissingDBRecordError{Err: fmt.Errorf("entry id %s: not found", id)}
	}

	return nil
}

// GenerateUniqueShortCode generates a string that does not already exist as a Lookup Ref
func (e *EntryRepo) GenerateUniqueShortCode(ctx context.Context) (string, error) {
	shortCode := generateRandomAlphaNumericString(shortCodeLength)

	_, err := e.Select(ctx, map[string]interface{}{
		"short_code": shortCode,
	}, false)
	switch err.(type) {
	case nil:
		// the short code already exists, so we need to generate a new one
		return e.GenerateUniqueShortCode(ctx)
	case domain.MissingDBRecordError:
		// the lookup ref we have generated is unique, we can return it
		return shortCode, nil
	}
	return "", err
}

// generateRandomAlphaNumericString returns a randomised string of given length
func generateRandomAlphaNumericString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}

	return string(b)
}

// NewEntryRepo instantiates a new EntryRepo with the provided DB agent
func NewEntryRepo(db *sql.DB) (*EntryRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db: %w", domain.ErrIsNil)
	}
	return &EntryRepo{db: db}, nil
}
