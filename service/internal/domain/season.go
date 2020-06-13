package domain

import (
	"fmt"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/models"
)

// ValidateSeason returns an error if validation rules are not satisfied for the provided Season
func ValidateSeason(s models.Season) error {
	var validationMsgs []string

	// ensure strings are not empty
	for k, v := range map[string]string{
		"ID":       s.ID,
		"ClientID": s.ClientID.Value(),
		"Name":     s.Name,
	} {
		if v == "" {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}

	// ensure timeframes are valid
	if !s.Active.Valid() {
		validationMsgs = append(validationMsgs, "Active timeframe must be valid")
	}
	if !s.EntriesAccepted.Valid() {
		validationMsgs = append(validationMsgs, "Entries Accepted timeframe must be valid")
	}
	if s.EntriesAccepted.OverlapsWith(s.Active) {
		return ValidationError{
			Reasons: []string{"Entries Accepted timeframe must have elapsed before Active timeframe begins"},
		}
	}
	switch {
	case len(s.SelectionsAccepted) < 1:
		validationMsgs = append(validationMsgs, "At least 1 Selections Accepted timeframe must exist")
	default:
		if !s.SelectionsAccepted[0].From.Equal(s.EntriesAccepted.From) || !s.SelectionsAccepted[0].Until.Equal(s.EntriesAccepted.Until) {
			validationMsgs = append(validationMsgs, "First Selections Accepted timeframe must be identical to Entries Accepted timeframe")
		}
	}

	for idx := 1; idx < len(s.SelectionsAccepted); idx++ {
		count := idx + 1
		thisTimeframe := s.SelectionsAccepted[idx]
		nextTimeframe := s.SelectionsAccepted[count]

		if !thisTimeframe.Valid() {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Selections Accepted timeframe #%d must be valid", count))
		}
		if thisTimeframe.OverlapsWith(nextTimeframe) {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Selections Accepted timeframe #%d must not overlap with #%d", count, count+1))
		}
		if !thisTimeframe.Until.Before(nextTimeframe.From) {
			validationMsgs = append(validationMsgs, fmt.Sprintf("Selections Accepted timeframes #%d and #%d must be chronological", count, count+1))
		}

		count++
	}

	// verify that each team exists and is not duplicated
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

	return nil
}
