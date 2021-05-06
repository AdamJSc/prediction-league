package sqltypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

// TODO - decide what to do with these types
// They are ported verbatim from the previous `coresql` dependency.
// They are types used on the domain model entities which conflates the responsibility of each application layer
// so they will need to remain here until this can be resolved.

// NullTime aliases sql.NullTime
type NullTime sql.NullTime

// MarshalJSON for NullTime
func (n NullTime) MarshalJSON() ([]byte, error) {
	var a *time.Time
	if n.Valid {
		a = &n.Time
	}
	return json.Marshal(a)
}

// Value for NullTime
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

// UnmarshalJSON for NullTime
func (n *NullTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = strings.Trim(s, `"`)

	var (
		zeroTime time.Time
		tim      time.Time
		err      error
	)

	if strings.EqualFold(s, "null") {
		return nil
	}

	if tim, err = time.Parse(time.RFC3339, s); err != nil {
		n.Valid = false
		return err
	}

	if tim == zeroTime {
		return nil
	}

	n.Time = tim
	n.Valid = true
	return nil
}

// Scan for NullTime
func (n *NullTime) Scan(src interface{}) error {
	// Set initial state for subsequent scans.
	n.Valid = false

	var a sql.NullTime
	if err := a.Scan(src); err != nil {
		return err
	}
	n.Time = a.Time
	if reflect.TypeOf(src) != nil {
		n.Valid = true
	}
	return nil
}

// ToNullTime creates a new NullTime
func ToNullTime(t time.Time) NullTime {
	if t == emptyTime {
		return NullTime(sql.NullTime{Valid: false})
	}
	return NullTime(sql.NullTime{Time: t, Valid: true})
}

// emptyTime allows default times to be considered
// null for insertion into the database.
var emptyTime = time.Time{}

// nullLiteral is helpful for checking
// for nulls, as they won't cause errors,
// yet we need the content of the file to change anyway
var nullLiteral = []byte("null")
