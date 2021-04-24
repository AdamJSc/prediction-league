package sqltypes

import (
	"bytes"
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

// NullString aliases sql.NullString
type NullString sql.NullString

// MarshalJSON for NullString
func (n NullString) MarshalJSON() ([]byte, error) {
	var a *string
	if n.Valid {
		a = &n.String
	}
	return json.Marshal(a)
}

// UnmarshalJSON for NullString
func (n *NullString) UnmarshalJSON(b []byte) error {
	if bytes.EqualFold(b, nullLiteral) {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &n.String)
	n.Valid = err == nil
	return err
}

// Value for NullString
func (n NullString) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.String, nil
}

// Scan for NullString
func (n *NullString) Scan(src interface{}) error {
	var a sql.NullString
	if err := a.Scan(src); err != nil {
		return err
	}
	n.String = a.String
	if reflect.TypeOf(src) != nil {
		n.Valid = true
	}
	return nil
}

// ToNullString returns a new NullString
func ToNullString(s *string) NullString {
	if s == nil {
		return NullString(sql.NullString{Valid: false})
	}
	return NullString(sql.NullString{String: *s, Valid: true})
}

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
