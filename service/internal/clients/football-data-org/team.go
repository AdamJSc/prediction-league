package football_data_org

import (
	"prediction-league/service/internal/models"
	"strconv"
)

// TeamIdentifier defines a team identifier for use with the football-data.org API
type TeamIdentifier struct {
	models.ResourceIdentifier
	TeamID int
}

func (t TeamIdentifier) Value() string {
	return strconv.Itoa(t.TeamID)
}
