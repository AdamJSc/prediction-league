package clients

import (
	"context"
	"prediction-league/service/internal/domain"
)

// FootballDataSource defines the interface for our external football data source
type FootballDataSource interface {
	RetrieveLatestStandingsBySeason(ctx context.Context, s *domain.Season) (*domain.Standings, error)
}

// EmailClient defines the interface for our email client
type EmailClient interface {
	SendEmail(ctx context.Context, em domain.Email) error
}
