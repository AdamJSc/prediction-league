package models

import (
	"fmt"
)

// Team defines the structure of a Team that belongs to a Season
type Team struct {
	ID        string             `json:"id"`         // team's initials, e.g. AFCB (must be unique)
	ClientID  ResourceIdentifier `json:"-"`          // identifier within the football data source
	Name      string             `json:"name"`       // team's long name, e.g. AFC Bournemouth
	ShortName string             `json:"short_name"` // team's short name, e.g. Bournemouth
	CrestURL  string             `json:"crest_url"`  // absolute URL representing image of team's crest
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
