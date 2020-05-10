package models

import (
	"errors"
	"strconv"
)

// FootballDataOrgTeamIdentifier defines a team identifier for use with the
type FootballDataOrgTeamIdentifier struct {
	ClientResourceIdentifier
	TeamID int
}

func (f FootballDataOrgTeamIdentifier) Value() string {
	return strconv.Itoa(f.TeamID)
}

// Team defines the structure of a Team that belongs to a Season
type Team struct {
	ID        string
	ClientID  FootballDataOrgTeamIdentifier
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
