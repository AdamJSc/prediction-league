package domain

import "context"

// FootballDataSource defines the interface for our external football data source
type FootballDataSource interface {
	RetrieveLatestStandingsBySeason(ctx context.Context, s Season) (Standings, error)
}

// EmailClient defines the interface for our email client
type EmailClient interface {
	SendEmail(ctx context.Context, em Email) error
}
