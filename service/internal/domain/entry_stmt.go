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
