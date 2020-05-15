package models

import (
	"errors"
)

// Team defines the structure of a Team that belongs to a Season
type Team struct {
	ID        string
	ClientID  ResourceIdentifier
	Name      string
	ShortName string
	CrestURL  string
}

// TeamCollection is map of Teams
type TeamCollection map[string]Team

// GetByID retrieves a matching Team from the collection by its ID
func (t TeamCollection) GetByID(teamID string) (Team, error) {
	for id, team := range t {
		if id == teamID {
			return team, nil
		}
	}

	return Team{}, errors.New("not found")
}

// GetByResourceID retrieves a matching Team from the collection by its ID
func (t TeamCollection) GetByResourceID(clientResourceID ResourceIdentifier) (Team, error) {
	for _, team := range t {
		if team.ClientID.Value() == clientResourceID.Value() {
			return team, nil
		}
	}

	return Team{}, errors.New("not found")
}
