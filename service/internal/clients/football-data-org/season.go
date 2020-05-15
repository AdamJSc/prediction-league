package football_data_org

import "prediction-league/service/internal/models"

// SeasonIdentifier defines a season identifier for use with the football-data.org API
type SeasonIdentifier struct {
	models.ResourceIdentifier
	SeasonID string
}

func (f SeasonIdentifier) Value() string {
	return f.SeasonID
}
