package domain

import (
	"encoding/json"
	coresql "github.com/LUSHDigital/core-sql"
	"time"
)

var dbEntryFields = "season_id, realm, entrant_name, entrant_nickname, entrant_email, team_id_sequence, status, payment_ref,"

func dbInsertEntry(db coresql.Agent, e *Entry) error {
	stmt := `INSERT INTO entry (id, ` + dbEntryFields + ` created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Truncate(time.Second)

	teamIDSequence, err := json.Marshal(e.TeamIDSequence)
	if err != nil {
		return wrapDBError(err)
	}

	if _, err := db.Query(
		stmt,
		e.ID,
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

func dbSelectEntries(db coresql.Agent, criteria map[string]interface{}, matchAny bool) ([]Entry, error) {
	whereStmt, params := dbWhereStmt(criteria, matchAny)

	stmt := `SELECT id, ` + dbEntryFields + ` created_at, updated_at FROM entry ` + whereStmt

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
			&entry.SeasonID,
			&entry.Realm,
			&entry.EntrantName,
			&entry.EntrantNickname,
			&entry.EntrantEmail,
			&teamIDSequence,
			&entry.Status,
			&entry.PaymentRef,
			&entry.CreatedAt,
			&entry.UpdateAt,
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
