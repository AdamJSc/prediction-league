package models

import (
	"fmt"
	"time"
)

const (
	SeasonStatusPending = "pending"
	SeasonStatusActive  = "active"
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
func (s Season) IsCompletedByStandings(standings Standings) bool {
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

// SeasonState defines the state of a Season
type SeasonState struct {
	Status                 string
	IsAcceptingEntries     bool
	IsAcceptingPredictions bool
	NextPredictionsWindow  *TimeFrame
}

// SeasonCollection is map of Seasons
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
