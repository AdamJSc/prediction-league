package domain

import (
	"fmt"
	"prediction-league/service/internal/models"
	"time"
)

// ValidateSeason returns an error if validation rules are not satisfied for the provided Season
func ValidateSeason(s models.Season) error {
	var validationMsgs []string

	// validate strings
	for k, v := range map[string]string{
		"ID":       s.ID,
		"ClientID": s.ClientID.Value(),
		"Name":     s.Name,
	} {
		if v == "" {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}

	// validate timestamps
	emptyTime := time.Time{}.Format(time.RFC3339Nano)
	if s.StartDate.Format(time.RFC3339Nano) == emptyTime {
		validationMsgs = append(validationMsgs, "Start Date must not be empty")
	}
	if s.EndDate.Format(time.RFC3339Nano) == emptyTime {
		validationMsgs = append(validationMsgs, "End Date must not be empty")
	}
	if s.EntriesFrom.Format(time.RFC3339Nano) == emptyTime {
		validationMsgs = append(validationMsgs, "Entries From Date must not be empty")
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	if !s.EntriesFrom.Before(s.StartDate) {
		return ValidationError{
			Reasons: []string{"EntriesFrom date cannot occur before Start date"},
		}
	}

	if !s.StartDate.Before(s.EndDate) {
		return ValidationError{
			Reasons: []string{"End date cannot occur before Start date"},
		}
	}

	return nil
}
