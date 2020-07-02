package domain

import (
	"errors"
	"fmt"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/models"
)

// ValidateSeason returns an error if validation rules are not satisfied for the provided Season
func ValidateSeason(s models.Season) error {
	if s.ID == datastore.FakeSeasonID {
		// don't validate our faked season
		return nil
	}

	// ensure strings are not empty
	for k, v := range map[string]string{
		"id":       s.ID,
		"clientID": s.ClientID.Value(),
		"name":     s.Name,
	} {
		if v == "" {
			return fmt.Errorf("%s must not be empty", k)
		}
	}

	// ensure timeframes are valid
	if !s.Active.Valid() {
		return errors.New("active timeframe must be valid")
	}
	if !s.EntriesAccepted.Valid() {
		return errors.New("entries accepted timeframe must be valid")
	}
	if s.EntriesAccepted.OverlapsWith(s.Active) {
		return errors.New("entries accepted timeframe must have elapsed before active timeframe begins")
	}
	switch {
	case len(s.PredictionsAccepted) < 1:
		return errors.New("at least 1 predictions accepted timeframe must exist")
	default:
		if !s.PredictionsAccepted[0].From.Equal(s.EntriesAccepted.From) || !s.PredictionsAccepted[0].Until.Equal(s.EntriesAccepted.Until) {
			return errors.New("first predictions accepted timeframe must be identical to entries accepted timeframe")
		}
	}

	for idx := 0; idx < len(s.PredictionsAccepted)-1; idx++ {
		nextIdx := idx + 1
		thisTimeframe := s.PredictionsAccepted[idx]
		nextTimeframe := s.PredictionsAccepted[nextIdx]

		if !thisTimeframe.Valid() {
			return fmt.Errorf("predictions accepted timeframe idx %d must be valid", idx)
		}
		if thisTimeframe.OverlapsWith(nextTimeframe) {
			return fmt.Errorf("predictions accepted timeframe idx %d must not overlap with idx %d", idx, nextIdx)
		}
		if !thisTimeframe.Until.Before(nextTimeframe.From) {
			return fmt.Errorf("predictions accepted timeframes idx %d and idx %d must be chronological", idx, nextIdx)
		}
	}

	// verify that each team exists and is not duplicated
	if _, err := FilterTeamsByIDs(s.TeamIDs, datastore.Teams); err != nil {
		return err
	}

	return nil
}
