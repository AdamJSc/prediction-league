package domain

import (
	"fmt"
	"prediction-league/service/internal/models"
)

// ValidateTeam returns an error if validation rules are not satisfied for the provided Team
func ValidateTeam(t models.Team) error {
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
func FilterTeamsByIDs(ids []string, collection models.TeamCollection) (models.TeamCollection, error) {
	var teams = make(models.TeamCollection)

	for _, id := range ids {
		team, err := collection.GetByID(id)
		if err != nil {
			return nil, NotFoundError{fmt.Errorf("missing team id: %s", id)}
		}

		if _, ok := teams[id]; ok {
			return nil, ConflictError{fmt.Errorf("team id exists multiple times: %s", id)}
		}

		teams[id] = team
	}

	return teams, nil
}
