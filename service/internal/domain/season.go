package domain

import (
	"fmt"
	"prediction-league/service/internal/datastore"
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
	var emptyTime time.Time
	if s.Active.From.Equal(emptyTime) {
		validationMsgs = append(validationMsgs, "Active From Date must not be empty")
	}
	if s.Active.Until.Equal(emptyTime) {
		validationMsgs = append(validationMsgs, "Active Until Date must not be empty")
	}
	if s.EntriesAccepted.From.Equal(emptyTime) {
		validationMsgs = append(validationMsgs, "Entries From Date must not be empty")
	}

	// validate teams
	var teams = make(map[string]struct{})
	for _, id := range s.TeamIDs {
		if _, err := datastore.Teams.GetByID(id); err != nil {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Invalid Team ID '%s'", id))
		}
		if _, ok := teams[id]; ok {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Team ID '%s' exists multiple times", id))
		}
		teams[id] = struct{}{}
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	if !s.EntriesAccepted.From.Before(s.Active.From) {
		return ValidationError{
			Reasons: []string{"Entries Accepted From date cannot occur before Active From date"},
		}
	}

	if !s.Active.From.Before(s.Active.Until) {
		return ValidationError{
			Reasons: []string{"Active Until date cannot occur before Active From date"},
		}
	}

	return nil
}
