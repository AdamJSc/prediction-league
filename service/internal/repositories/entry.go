package repositories

import (
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"golang.org/x/net/context"
	"prediction-league/service/internal/models"
	"time"
)

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

// EntryRepository defines the interface for transacting with our Entry data source
type EntryRepository interface {
	Insert(ctx context.Context, entry *models.Entry) error
	Update(ctx context.Context, entry *models.Entry) error
	Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.Entry, error)
	ExistsByID(ctx context.Context, id string) error
}

// EntryDatabaseRepository defines our DB-backed Entry data store
type EntryDatabaseRepository struct {
	agent coresql.Agent
}

// Insert inserts a new Entry into the database
func (e EntryDatabaseRepository) Insert(ctx context.Context, entry *models.Entry) error {
	stmt := `INSERT INTO entry (id, ` + getDBFieldsStringFromFields(entryDBFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	rows, err := e.agent.QueryContext(
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
func (e EntryDatabaseRepository) Update(ctx context.Context, entry *models.Entry) error {
	stmt := `UPDATE entry
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(entryDBFields) + `, updated_at = ?
				WHERE id = ?`

	now := sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	rows, err := e.agent.QueryContext(
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

	entry.UpdatedAt = now

	return nil
}

// Select retrieves Entries from our database based on the provided criteria
func (e EntryDatabaseRepository) Select(ctx context.Context, criteria map[string]interface{}, matchAny bool) ([]models.Entry, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(entryDBFields) + `, created_at, updated_at FROM entry ` + whereStmt

	rows, err := e.agent.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var entries []models.Entry
	for rows.Next() {
		entry := models.Entry{}

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
		return nil, MissingDBRecordError{errors.New("no entries found")}
	}

	return entries, nil
}

// ExistsByID determines whether an Entry with the provided ID exists in the database
func (e EntryDatabaseRepository) ExistsByID(ctx context.Context, id string) error {
	stmt := `SELECT COUNT(*) FROM entry WHERE id = ?`

	row := e.agent.QueryRowContext(ctx, stmt, id)

	var count int
	if err := row.Scan(&count); err != nil {
		return wrapDBError(err)
	}

	if count == 0 {
		return MissingDBRecordError{errors.New("entry not found")}
	}

	return nil
}

// NewEntryDatabaseRepository instantiates a new EntryDatabaseRepository with the provided DB agent
func NewEntryDatabaseRepository(db coresql.Agent) EntryDatabaseRepository {
	return EntryDatabaseRepository{agent: db}
}
