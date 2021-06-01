package domain

import (
	"context"
	"errors"
	"fmt"
)

// FootballDataSource defines the interface for our external football data source
type FootballDataSource interface {
	RetrieveLatestStandingsBySeason(ctx context.Context, s Season) (Standings, error)
}

type NoopFootballDataSource struct {
	l Logger
}

// RetrieveLatestStandingsBySeason implements FootballDataSource
func (l *NoopFootballDataSource) RetrieveLatestStandingsBySeason(_ context.Context, s Season) (Standings, error) {
	// TODO - logger: replace with debugf logger method
	l.l.Infof("noop retrieved latest standings for season: %s", s.ID)
	return Standings{}, errors.New("process aborted")
}

func NewNoopFootballDataSource(l Logger) (*NoopFootballDataSource, error) {
	if l == nil {
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	return &NoopFootballDataSource{l}, nil
}
