package models

import "strconv"

// ResourceIdentifier defines a generic interface for
// retrieving the value that identifies a resource
type ResourceIdentifier interface {
	Value() string
}

// SeasonIdentifier defines a season identifier for use with the football-data.org API
type SeasonIdentifier struct {
	SeasonID string
}

func (f SeasonIdentifier) Value() string {
	return f.SeasonID
}

// TeamIdentifier defines a team identifier for use with the football-data.org API
type TeamIdentifier struct {
	TeamID int
}

func (t TeamIdentifier) Value() string {
	return strconv.Itoa(t.TeamID)
}
