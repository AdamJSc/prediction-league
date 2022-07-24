package domain

import (
	"errors"
	"fmt"
	"time"
)

const (
	// SeasonStatePending represents a Season timeframe whose status is PENDING
	SeasonStatePending = "pending"
	// SeasonStateActive represents a Season timeframe whose status is ACTIVE
	SeasonStateActive = "active"
	// SeasonStateElapsed represents a Season timeframe whose status is ELAPSED
	SeasonStateElapsed = "elapsed"
	// elapsingGracePeriod represents the grace period prior to a timeframe elapsing
	elapsingGracePeriod = 6 * time.Hour
)

// Season defines the structure of a Season against which Entries are played
type Season struct {
	ID                  string             // representation of season's start/end year along with instance number, e.g. 202021_1
	ClientID            ResourceIdentifier // identifier within the football data source
	Name                string             // season name, e.g. Premier League 2022/23
	ShortName           string             // short name, e.g. Prem 22/23
	Live                TimeFrame          // timeframe for which the season is live (real-world standings will be consumed during this timeframe)
	EntriesAccepted     TimeFrame          // timeframe within which new entries will be accepted
	PredictionsAccepted TimeFrame          // timeframe within which changes to entry predictions will be accepted
	TeamIDs             []string           // slice of strings representing valid team IDs that exist within TeamsCollection
	MaxRounds           int                // number of rounds after which season is considered completed (maximum number of games to be played by each team)
}

// GetState determines a Season's state based on a supplied timestamp
func (s Season) GetState(ts time.Time) SeasonState {
	var getCurrentStatus = func(tf TimeFrame) string {
		var status string
		switch {
		case tf.HasBegunBy(ts) && !tf.HasElapsedBy(ts):
			status = SeasonStateActive
		case tf.HasElapsedBy(ts):
			status = SeasonStateElapsed
		default:
			status = SeasonStatePending
		}
		return status
	}

	var getIsClosing = func(tf TimeFrame) bool {
		graceTs := ts.Add(elapsingGracePeriod)
		return tf.HasElapsedBy(graceTs) && !tf.HasElapsedBy(ts)
	}

	return SeasonState{
		LiveStatus:         getCurrentStatus(s.Live),
		EntriesStatus:      getCurrentStatus(s.EntriesAccepted),
		PredictionsStatus:  getCurrentStatus(s.PredictionsAccepted),
		PredictionsClosing: getIsClosing(s.PredictionsAccepted),
	}
}

// IsCompletedByStandings returns true if the provided standings represents a completed final round, otherwise false
func (s Season) IsCompletedByStandings(stnd Standings) bool {
	if stnd.SeasonID != s.ID {
		// standings pertain to a different season
		return false
	}

	for _, rwm := range stnd.Rankings {
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
	if s.PredictionsAccepted.BeginsWithin(tf) {
		stf := SequencedTimeFrame{
			Count:   1,
			Total:   1,
			Current: &s.PredictionsAccepted,
		}

		return stf, nil
	}

	return SequencedTimeFrame{}, ErrNoMatchingPredictionWindow
}

// GetPredictionWindowEndsWithin returns the Prediction Window that ends within the provided TimeFrame,
// or an error if no match is found
func (s Season) GetPredictionWindowEndsWithin(tf TimeFrame) (SequencedTimeFrame, error) {
	if s.PredictionsAccepted.EndsWithin(tf) {
		stf := SequencedTimeFrame{
			Count:   1,
			Total:   1,
			Current: &s.PredictionsAccepted,
		}

		return stf, nil
	}

	return SequencedTimeFrame{}, ErrNoMatchingPredictionWindow
}

// SeasonState defines the state of a Season
type SeasonState struct {
	LiveStatus         string
	EntriesStatus      string
	PredictionsStatus  string
	PredictionsClosing bool
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

	return Season{}, NotFoundError{fmt.Errorf("season id %s: not found", seasonID)}
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
func ValidateSeason(s Season, tc TeamCollection) error {
	if s.ID == FakeSeasonID {
		// don't validate our faked season
		return nil
	}

	if s.ClientID == nil {
		return errors.New("client id must not be nil")
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
	if !s.Live.Valid() {
		return errors.New("live timeframe must be valid")
	}
	if !s.EntriesAccepted.Valid() {
		return errors.New("entries accepted timeframe must be valid")
	}
	if !s.PredictionsAccepted.Valid() {
		return errors.New("predictions accepted timeframe must be valid")
	}
	if !s.PredictionsAccepted.From.Equal(s.EntriesAccepted.From) {
		return errors.New("predictions must be accepted from the same time as entries")
	}
	if !s.PredictionsAccepted.Until.After(s.EntriesAccepted.Until) {
		return errors.New("predictions must be accepted for a longer duration than entries")
	}

	// verify that each team exists and is not duplicated
	if _, err := FilterTeamsByIDs(s.TeamIDs, tc); err != nil {
		return err
	}

	return nil
}
