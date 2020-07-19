package clients

import (
	"context"
	"prediction-league/service/internal/models"
)

// FootballDataSource defines the interface for our external football data source
type FootballDataSource interface {
	RetrieveLatestStandingsBySeason(ctx context.Context, s *models.Season) (*models.Standings, error)
}

// EmailClient defines the interface for our email client
type EmailClient interface {
	SendEmail(message models.EmailMessage) error
}
