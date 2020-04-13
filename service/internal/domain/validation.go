package domain

import (
	"errors"
	"github.com/ladydascalie/v"
	"reflect"
	"time"
)

func RegisterCustomValidators() {
	v.Set("notempty", func(args string, value, structure interface{}) error {
		var empty interface{}

		switch value.(type) {
		case string:
			empty = ""
		case time.Time:
			empty = time.Time{}
		}

		if reflect.DeepEqual(value, empty) {
			return errors.New("must not be empty")
		}

		return nil
	})
}
