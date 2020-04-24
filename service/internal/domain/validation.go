package domain

import (
	"errors"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"github.com/ladydascalie/v"
	"reflect"
	"regexp"
	"time"
)

// RegisterCustomValidators provides a bootstrap for registering customer validation functions with the `v` package
func RegisterCustomValidators() {
	v.Set("notEmpty", func(args string, value, structure interface{}) error {
		var empty interface{}

		switch value.(type) {
		case string:
			empty = ""
		case time.Time:
			empty = time.Time{}
		case uuid.UUID:
			empty = uuid.UUID{}
		}

		if reflect.TypeOf(value) == reflect.TypeOf(empty) && !reflect.DeepEqual(value, empty) {
			return nil
		}

		return errors.New("must not be empty")
	})

	v.Set("isEntryStatus", func(args string, value, structure interface{}) error {
		switch value {
		case EntryStatusPending, EntryStatusPaid, EntryStatusReady:
			return nil
		}

		return errors.New("invalid entry status")
	})

	v.Set("isEntryPaymentMethod", func(args string, value, structure interface{}) error {
		valueAsNullString, ok := value.(sqltypes.NullString)
		if !ok {
			return errors.New("couldn't convert to null string")
		}

		if valueAsNullString.String == "" {
			// can be empty
			return nil
		}

		switch valueAsNullString.String {
		case EntryPaymentMethodPayPal, EntryPaymentMethodOther:
			return nil
		}

		return errors.New("invalid entry payment method")
	})

	v.Set("email", func(args string, value, structure interface{}) error {
		if value.(string) == "" {
			return errors.New("missing email")
		}

		var pattern, err = regexp.Compile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		if err != nil {
			return err
		}

		if !pattern.MatchString(value.(string)) {
			return errors.New("invalid email")
		}

		return nil
	})
}
