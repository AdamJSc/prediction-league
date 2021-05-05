package domain

import (
	"errors"
	"fmt"
	"time"
)

const (
	// SeasonStatusPending represents a Season whose status is PENDING
	SeasonStatusPending = "pending"
	// SeasonStatusActive represents a Season whose status is ACTIVE
	SeasonStatusActive = "active"
	// SeasonStatusElapsed represents a Season whose status is ELAPSED
	SeasonStatusElapsed = "elapsed"
)

// Season defines the structure of a Season against which Entries are played
type Season struct {
	ID                  string             // representation of season's start/end year along with instance number, e.g. 202021_1
	ClientID            ResourceIdentifier // identifier within the football data source
	Name                string             // season name, e.g. Premier League 2020/21
	Active              TimeFrame          // timeframe for which the season is active (real-world standings will be consumed during this timeframe)
	EntriesAccepted     TimeFrame          // timeframe within which new entries will be accepted
	PredictionsAccepted []TimeFrame        // series of timeframes within which changes to entry predictions will be accepted
	TeamIDs             []string           // slice of strings representing valid team IDs that exist within TeamsCollection
	MaxRounds           int                // number of rounds after which season is considered completed (maximum number of games to be played by each team)
}

// GetState determines a Season's state based on a supplied timestamp
func (s Season) GetState(ts time.Time) SeasonState {
	var state SeasonState

	// determine season's current status
	state.Status = SeasonStatusPending
	switch {
	case s.Active.HasBegunBy(ts) && !s.Active.HasElapsedBy(ts):
		state.Status = SeasonStatusActive
	case s.Active.HasElapsedBy(ts):
		state.Status = SeasonStatusElapsed
	}

	// is season currently accepting entries?
	if s.EntriesAccepted.HasBegunBy(ts) && !s.EntriesAccepted.HasElapsedBy(ts) {
		state.IsAcceptingEntries = true
	}

	// is season currently accepting predictions?
	for _, tf := range s.PredictionsAccepted {
		thisTf := tf

		if tf.HasBegunBy(ts) && !tf.HasElapsedBy(ts) {
			// next predictions window should be the current timeframe if predictions are currently being accepted
			state.IsAcceptingPredictions = true
			state.NextPredictionsWindow = &thisTf
			break
		}

		// if we aren't currently accepting predictions, does this tf represent the next time that we are?
		if !state.IsAcceptingPredictions && !tf.HasBegunBy(ts) {
			nextTf := tf
			state.NextPredictionsWindow = &nextTf
			break
		}
	}

	return state
}

// IsCompletedByStandings returns true if the provided standings represents a completed final round, otherwise false
func (s Season) IsCompletedByStandings(standings *Standings) bool {
	if standings.SeasonID != s.ID {
		// standings pertain to a different season
		return false
	}

	for _, rwm := range standings.Rankings {
		played, ok := rwm.MetaData[MetaKeyPlayedGames]
		if !ok || played != s.MaxRounds {
			// this ranked team has not played the maximum number of games, so season is not complete
			return false
		}
	}

	// season is complete
	return true
}

// GetPredictionWindowBeginsWithin returns the Prediction Window that begins within the provided TimeFrame,
// or an error if no match is found
func (s Season) GetPredictionWindowBeginsWithin(tf TimeFrame) (SequencedTimeFrame, error) {
	total := len(s.PredictionsAccepted)
	count := 0

	for idx, window := range s.PredictionsAccepted {
		count++
		if window.BeginsWithin(tf) {
			stf := SequencedTimeFrame{
				Count:   count,
				Total:   total,
				Current: &window,
			}

			if count < total {
				stf.Next = &s.PredictionsAccepted[idx+1]
			}

			return stf, nil
		}
	}

	return SequencedTimeFrame{}, ErrNoMatchingPredictionWindow
}

// GetPredictionWindowEndsWithin returns the Prediction Window that ends within the provided TimeFrame,
// or an error if no match is found
func (s Season) GetPredictionWindowEndsWithin(tf TimeFrame) (SequencedTimeFrame, error) {
	total := len(s.PredictionsAccepted)
	count := 0

	for idx, window := range s.PredictionsAccepted {
		count++
		if window.EndsWithin(tf) {
			stf := SequencedTimeFrame{
				Count:   count,
				Total:   total,
				Current: &window,
			}

			if count < total {
				stf.Next = &s.PredictionsAccepted[idx+1]
			}

			return stf, nil
		}
	}

	return SequencedTimeFrame{}, ErrNoMatchingPredictionWindow
}

// SeasonState defines the state of a Season
type SeasonState struct {
	Status                 string
	IsAcceptingEntries     bool
	IsAcceptingPredictions bool
	NextPredictionsWindow  *TimeFrame
}

// SeasonCollection is map of Season
type SeasonCollection map[string]Season

// GetByID retrieves a matching Season from the collection by its ID
func (s SeasonCollection) GetByID(seasonID string) (Season, error) {
	for id, season := range s {
		if id == seasonID {
			return season, nil
		}
	}

	return Season{}, fmt.Errorf("season id %s: not found", seasonID)
}

// SeasonIdentifier defines a season identifier for use with the football-data.org API
type SeasonIdentifier struct {
	SeasonID string
}

func (f SeasonIdentifier) Value() string {
	return f.SeasonID
}

// ResourceIdentifier defines a generic interface for
// retrieving the value that identifies a resource
type ResourceIdentifier interface {
	Value() string
}

// ValidateSeason returns an error if validation rules are not satisfied for the provided Season
func ValidateSeason(s Season) error {
	if s.ID == fakeSeasonID {
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
	if _, err := FilterTeamsByIDs(s.TeamIDs, TeamsDataStore); err != nil {
		return err
	}

	return nil
}
