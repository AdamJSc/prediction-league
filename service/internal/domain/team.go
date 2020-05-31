package domain

import (
	"fmt"
	"prediction-league/service/internal/models"
)

// ValidateTeam returns an error if validation rules are not satisfied for the provided Team
func ValidateTeam(t models.Team) error {
	var validationMsgs []string

	// validate values
	for k, v := range map[string]struct {
		actual  string
		invalid string
	}{
		"ID":        {actual: t.ID, invalid: ""},
		"Name":      {actual: t.Name, invalid: ""},
		"ShortName": {actual: t.ShortName, invalid: ""},
		"CrestURL":  {actual: t.CrestURL, invalid: ""},
		"ClientID":  {actual: t.ClientID.Value(), invalid: "0"},
	} {
		if v.actual == v.invalid {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	return nil
}
