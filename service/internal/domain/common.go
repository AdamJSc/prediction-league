package domain

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"os"
	"reflect"
	"strings"
)

const (
	ctxKeyRealm    = "REALM"
	ctxKeyRealmPIN = "REALM_PIN"

	envKeyAdminBasicAuth = "ADMIN_BASIC_AUTH"
)

type BadRequestError struct{ Err error }

func (e BadRequestError) Error() string {
	return e.Err.Error()
}

type ValidationError struct {
	Reasons []string `json:"reasons"`
	Fields  []string `json:"fields"`
}

func (e ValidationError) Error() string {
	reasons := strings.Join(e.Reasons, " | ")
	fields := strings.Join(e.Fields, " | ")
	return fmt.Sprintf("reasons: %s, with fields: %v", strings.ToLower(reasons), strings.ToLower(fields))
}

type UnauthorizedError struct{ error }

type ConflictError struct{ error }

type NotFoundError struct{ error }

type InternalError struct{ error }

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

type dbMissingRecordError struct {
	error
}

type dbDuplicateRecordError struct {
	error
}

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

func domainErrorFromDBError(err error) error {
	switch err.(type) {
	case dbDuplicateRecordError:
		return ConflictError{err}
	case dbMissingRecordError:
		return NotFoundError{err}
	}

	return InternalError{err}
}

func ContextFromRequest(r *http.Request) Context {
	ctx := Context{context.Background()}

	// get realm from host (strip port)
	realm := strings.Trim(strings.Split(r.Host, ":")[0], " ")
	ctx.setRealm(realm)

	// get realm PIN from env
	realmFormattedForEnvKey := strings.ToUpper(strings.Replace(realm, ".", "_", -1))
	envKeyForRealmPIN := strings.ToUpper(fmt.Sprintf("%s_%s", ctxKeyRealmPIN, realmFormattedForEnvKey))

	realmPIN := os.Getenv(envKeyForRealmPIN)
	ctx.setRealmPIN(realmPIN)

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

func validateBasicAuth(ctx context.Context) error {
	expected := os.Getenv(envKeyAdminBasicAuth)
	actual := ctx.Value(envKeyAdminBasicAuth).(string)

	if actual != expected {
		return UnauthorizedError{errors.New("unauthorized")}
	}

	return nil
}

type Context struct {
	context.Context
}

func (c *Context) setRealm(realm string) {
	c.setString(ctxKeyRealm, realm)
}

func (c *Context) getRealm() string {
	return c.getString(ctxKeyRealm)
}

func (c *Context) setRealmPIN(pin string) {
	c.setString(ctxKeyRealmPIN, pin)
}

func (c *Context) getRealmPIN() string {
	return c.getString(ctxKeyRealmPIN)
}

func (c *Context) setString(key string, value string) {
	c.Context = context.WithValue(c.Context, key, value)
}

func (c *Context) getString(key string) string {
	var value string
	ctxValue := c.Context.Value(key)

	if ctxValue != nil {
		value = ctxValue.(string)
	}

	return value
}
