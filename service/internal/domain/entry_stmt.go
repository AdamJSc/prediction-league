package domain

import (
	"encoding/json"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"time"
)

// dbEntryFields defines the fields used regularly in Entry-related transactions
var dbEntryFields = []string{
	"lookup_ref",
	"season_id",
	"realm",
	"entrant_name",
	"entrant_nickname",
	"entrant_email",
	"team_id_sequence",
	"status",
	"payment_ref",
}

// DBInsertEntry insert an Entry to the database
func DBInsertEntry(db coresql.Agent, e *Entry) error {
	stmt := `INSERT INTO entry (id, ` + getDBFieldsStringFromFields(dbEntryFields) + `, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	teamIDSequence, err := json.Marshal(e.TeamIDSequence)
	if err != nil {
		return wrapDBError(err)
	}

	if _, err := db.Query(
		stmt,
		e.ID,
		e.LookupRef,
		e.SeasonID,
		e.Realm,
		e.EntrantName,
		e.EntrantNickname,
		e.EntrantEmail,
		teamIDSequence,
		e.Status,
		e.PaymentRef,
		now,
	); err != nil {
		return wrapDBError(err)
	}

	e.CreatedAt = now

	return nil
}

// dbUpdateEntry update an existing Entry in the database
func dbUpdateEntry(db coresql.Agent, e *Entry) error {
	stmt := `UPDATE entry
				SET ` + getDBFieldsWithEqualsPlaceholdersStringFromFields(dbEntryFields) + `, updated_at = ?
				WHERE id = ?`

	now := sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	teamIDSequence, err := json.Marshal(e.TeamIDSequence)
	if err != nil {
		return wrapDBError(err)
	}

	if _, err := db.Query(
		stmt,
		e.LookupRef,
		e.SeasonID,
		e.Realm,
		e.EntrantName,
		e.EntrantNickname,
		e.EntrantEmail,
		teamIDSequence,
		e.Status,
		e.PaymentRef,
		now,
		e.ID,
	); err != nil {
		return wrapDBError(err)
	}

	e.UpdatedAt = now

	return nil
}

// dbSelectEntries retrieves Entries from the database
func dbSelectEntries(db coresql.Agent, criteria map[string]interface{}, matchAny bool) ([]Entry, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + getDBFieldsStringFromFields(dbEntryFields) + `, created_at, updated_at FROM entry ` + whereStmt

	rows, err := db.Query(stmt, params...)
	if err != nil {
		return []Entry{}, wrapDBError(err)
	}

	var entries []Entry
	for rows.Next() {
		entry := Entry{}

		var teamIDSequence []byte

		if err := rows.Scan(
			&entry.ID,
			&entry.LookupRef,
			&entry.SeasonID,
			&entry.Realm,
			&entry.EntrantName,
			&entry.EntrantNickname,
			&entry.EntrantEmail,
			&teamIDSequence,
			&entry.Status,
			&entry.PaymentRef,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return []Entry{}, err
		}

		if err := json.Unmarshal(teamIDSequence, &entry.TeamIDSequence); err != nil {
			return []Entry{}, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
