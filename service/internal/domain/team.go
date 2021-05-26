package domain

import (
	"fmt"
	"strconv"
)

// Team defines the structure of a Team that belongs to a Season
type Team struct {
	ID        string             `json:"id"`         // team's initials, e.g. AFCB (must be unique)
	ClientID  ResourceIdentifier `json:"-"`          // identifier within the football data source
	Name      string             `json:"name"`       // team's long name, e.g. AFC Bournemouth
	ShortName string             `json:"short_name"` // team's short name, e.g. Bournemouth
	CrestURL  string             `json:"crest_url"`  // absolute URL representing image of team's crest
}

// TeamCollection is map of Team
type TeamCollection map[string]Team

// GetByID retrieves a matching Team from the collection by its ID
func (t TeamCollection) GetByID(teamID string) (Team, error) {
	for id, team := range t {
		if id == teamID {
			return team, nil
		}
	}

	return Team{}, NotFoundError{fmt.Errorf("team id %s: not found", teamID)}
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

// TeamIdentifier defines a team identifier for use with the football-data.org API
type TeamIdentifier struct {
	TeamID int
}

func (t TeamIdentifier) Value() string {
	return strconv.Itoa(t.TeamID)
}

// ValidateTeam returns an error if validation rules are not satisfied for the provided Team
func ValidateTeam(t Team) error {
	// validate values
	for k, v := range map[string]struct {
		actual  string
		invalid string
	}{
		"id":        {actual: t.ID, invalid: ""},
		"name":      {actual: t.Name, invalid: ""},
		"shortName": {actual: t.ShortName, invalid: ""},
		"crestURL":  {actual: t.CrestURL, invalid: ""},
		"clientID":  {actual: t.ClientID.Value(), invalid: "0"},
	} {
		if v.actual == v.invalid {
			return fmt.Errorf("%s must not be empty", k)
		}
	}

	return nil
}

// FilterTeamsByIDs returns the provided TeamCollection filtered by the provided IDs
func FilterTeamsByIDs(ids []string, collection TeamCollection) ([]Team, error) {
	var teams []Team

	seen := make(map[string]struct{})

	for _, id := range ids {
		team, err := collection.GetByID(id)
		if err != nil {
			return nil, NotFoundError{fmt.Errorf("missing team id: %s", id)}
		}

		if _, ok := seen[id]; ok {
			return nil, ConflictError{fmt.Errorf("team id exists multiple times: %s", id)}
		}

		teams = append(teams, team)
		seen[id] = struct{}{}
	}

	return teams, nil
}
