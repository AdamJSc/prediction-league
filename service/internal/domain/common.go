package domain

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	ctxKeyRealm         = "REALM"
	ctxKeyRealmPIN      = "REALM_PIN"
	ctxKeyRealmSeasonID = "REALM_SEASON_ID"

	envKeyAdminBasicAuth = "ADMIN_BASIC_AUTH"
)

// BadRequestError translates to a 400 Bad Request response status code
type BadRequestError struct{ Err error }

func (e BadRequestError) Error() string {
	return e.Err.Error()
}

// UnauthorizedError translates to a 401 Unauthorized response status code
type UnauthorizedError struct{ error }

// NotFoundError translates to a 404 Not Found response status code
type NotFoundError struct{ error }

// ConflictError translates to a 409 Conflict response status code
type ConflictError struct{ error }

// ValidationError translates to a 422 Unprocessable Entity response status code
type ValidationError struct {
	Reasons []string `json:"reasons"`
	Fields  []string `json:"fields"`
}

func (e ValidationError) Error() string {
	reasons := strings.Join(e.Reasons, " | ")
	fields := strings.Join(e.Fields, " | ")
	return fmt.Sprintf("reasons: %s, with fields: %v", strings.ToLower(reasons), strings.ToLower(fields))
}

// InternalError translates to a 500 Internal Server Error response status code
type InternalError struct{ error }

// vPackageErrorToValidationError extracts field names and messages from an error returned by the `v` validation package
// and transforms them into a ValidationError
func vPackageErrorToValidationError(err error, structure interface{}) ValidationError {
	var reasons, fields []string

	for _, validationPart := range strings.Split(err.Error(), "[validation]") {
		// reason is the full validation text
		reason := strings.Trim(validationPart, " ")
		if reason == "" {
			continue
		}

		reasons = append(reasons, reason)

		// structFieldName is the first part of the validation text, up to the colon delimiter
		structFieldName := strings.Trim(strings.Split(reason, ":")[0], " ")
		if structFieldName == "" {
			continue
		}

		// fallback to the original struct field name
		field := structFieldName
		if structField, ok := reflect.TypeOf(structure).FieldByName(structFieldName); ok {
			// try to get a db annotation for the struct field
			if tag := structField.Tag.Get("db"); tag != "" {
				field = tag
			}
			// try to get a json annotation
			if tag := structField.Tag.Get("json"); tag != "" {
				field = tag
			}
		}
		fields = append(fields, field)
	}

	return ValidationError{
		Reasons: reasons,
		Fields:  fields,
	}
}

// dbMissingRecordError represents an error from an SQL agent that pertains to a missing record
type dbMissingRecordError struct {
	error
}

// dbDuplicateRecordError represents an error from an SQL agent that pertains to a unique constraint violation
type dbDuplicateRecordError struct {
	error
}

// wrapDBError wraps an error from an SQL agent according to its nature as per the representations above
func wrapDBError(err error) error {
	if e, ok := err.(*mysql.MySQLError); ok {
		switch e.Number {
		case 1060:
		case 1061:
		case 1062:
			return dbDuplicateRecordError{err}
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return dbMissingRecordError{err}
	}

	return err
}

// domainErrorFromDBError returns the appropriate domain-level error from a database-specific error
func domainErrorFromDBError(err error) error {
	switch err.(type) {
	case dbDuplicateRecordError:
		return ConflictError{err}
	case dbMissingRecordError:
		return NotFoundError{err}
	}

	return InternalError{err}
}

// dbWhereStmt returns the WHERE clause portion of an SQL statement as a string, plus the parameters to
// pass to the operation, from a given map of criteria to query on
func dbWhereStmt(criteria map[string]interface{}, matchAny bool) (stmt string, params []interface{}) {
	var conditions []string

	for field, value := range criteria {
		conditions = append(conditions, fmt.Sprintf("(%s = ?)", field))
		params = append(params, value)
	}

	comparison := " AND "
	if matchAny {
		comparison = " OR "
	}

	if len(conditions) > 0 {
		stmt = `WHERE ` + strings.Join(conditions, comparison)
	}

	return stmt, params
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

// Context wraps a standard context for the purpose of additional helper methods
type Context struct {
	context.Context
}

// setString sets a context value whose type is a string
func (c *Context) setString(key string, value string) {
	c.Context = context.WithValue(c.Context, key, value)
}

// getString retrieves a context value whose type is a string
func (c *Context) getString(key string) string {
	var value string
	ctxValue := c.Context.Value(key)

	if ctxValue != nil {
		value = ctxValue.(string)
	}

	return value
}

// SetRealm sets the context's Realm value
func (c *Context) SetRealm(realm string) {
	c.setString(ctxKeyRealm, realm)
}

// GetRealm retrieves the context's Realm value
func (c *Context) GetRealm() string {
	return c.getString(ctxKeyRealm)
}

// SetRealmPIN sets the context's Realm PIN value
func (c *Context) SetRealmPIN(pin string) {
	c.setString(ctxKeyRealmPIN, pin)
}

// GetRealmPIN retrieves the context's Realm PIN value
func (c *Context) GetRealmPIN() string {
	return c.getString(ctxKeyRealmPIN)
}

// SetRealmSeasonID sets the context's Realm Season ID value
func (c *Context) SetRealmSeasonID(seasonID string) {
	c.setString(ctxKeyRealmSeasonID, seasonID)
}

// GetRealmSeasonID retrieves the context's Realm Season ID value
func (c *Context) GetRealmSeasonID() string {
	return c.getString(ctxKeyRealmSeasonID)
}

// NewContext returns a new Context
func NewContext() Context {
	return Context{Context: context.Background()}
}

// ContextFromRequest extracts data from a given request object and returns a domain object Context
func ContextFromRequest(r *http.Request) Context {
	ctx := NewContext()

	// get realm from host (strip port)
	realm := strings.Trim(strings.Split(r.Host, ":")[0], " ")
	realmFormattedForEnvKey := strings.ToUpper(strings.Replace(realm, ".", "_", -1))
	ctx.SetRealm(realm)

	// get realm PIN from env
	envKeyForRealmPIN := strings.ToUpper(fmt.Sprintf("%s_%s", realmFormattedForEnvKey, ctxKeyRealmPIN))
	realmPIN := os.Getenv(envKeyForRealmPIN)
	ctx.SetRealmPIN(realmPIN)

	// get realm Season ID from env
	envKeyForRealmSeasonID := strings.ToUpper(fmt.Sprintf("%s_%s", realmFormattedForEnvKey, ctxKeyRealmSeasonID))
	realmSeasonID := os.Getenv(envKeyForRealmSeasonID)
	ctx.SetRealmSeasonID(realmSeasonID)

	// set basic auth username/password requirements
	var userPass []byte
	authHeader := r.Header.Get("Authorization")
	headerParts := strings.Split(authHeader, "Basic ")
	if len(headerParts) == 2 {
		userPass, _ = base64.StdEncoding.DecodeString(headerParts[1])
	}
	ctx.Context = context.WithValue(ctx.Context, envKeyAdminBasicAuth, string(userPass))

	return ctx
}

// validateRealmPIN checks that the supplied PIN matches the Realm PIN added to the Context
func validateRealmPIN(ctx Context, pin string) error {
	realmPIN := ctx.GetRealmPIN()
	if realmPIN == "" || realmPIN != pin {
		return UnauthorizedError{errors.New("unauthorized")}
	}

	return nil
}

// getDBFieldsStringFromFields returns a statement-ready string of fields names
func getDBFieldsStringFromFields(fields []string) string {
	return strings.Join(fields, ", ")
}

// getDBFieldsWithEqualsPlaceholdersStringFromFields returns a statement-ready string of fields names with "equals value" placeholders
func getDBFieldsWithEqualsPlaceholdersStringFromFields(fields []string) string {
	var fieldsWithEqualsPlaceholders []string

	for _, field := range fields {
		fieldsWithEqualsPlaceholders = append(fieldsWithEqualsPlaceholders, fmt.Sprintf("%s = ?", field))
	}

	return strings.Join(fieldsWithEqualsPlaceholders, ", ")
}
