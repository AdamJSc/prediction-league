package domain

import (
	"errors"
	"github.com/LUSHDigital/uuid"
	"github.com/ladydascalie/v"
	"reflect"
	"regexp"
	"time"
)

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
		case EntryStatusPending, EntryStatusPaid, EntryStatusComplete:
			return nil
		}

		return errors.New("invalid entry status")
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
