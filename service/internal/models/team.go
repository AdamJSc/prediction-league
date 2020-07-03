package models

import (
	"fmt"
)

// Team defines the structure of a Team that belongs to a Season
type Team struct {
	ID        string             `json:"id"`
	ClientID  ResourceIdentifier `json:"-"`
	Name      string             `json:"name"`
	ShortName string             `json:"short_name"`
	CrestURL  string             `json:"crest_url"`
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

	return Team{}, fmt.Errorf("team id %s: not found", teamID)
}

// GetByResourceID retrieves a matching Team from the collection by its ID
func (t TeamCollection) GetByResourceID(clientResourceID ResourceIdentifier) (Team, error) {
	for _, team := range t {
		if team.ClientID.Value() == clientResourceID.Value() {
			return team, nil
		}
	}

	return Team{}, fmt.Errorf("team client resource id %s: not found", clientResourceID.Value())
}
