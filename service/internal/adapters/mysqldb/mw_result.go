package mysqldb

import (
	"context"
	"database/sql"
	"fmt"
	"prediction-league/service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// MatchWeekResultRepo implements operations against a sql db
type MatchWeekResultRepo struct {
	db     *sql.DB
	idFn   idFunc
	timeFn timeFunc
}

// GetBySubmissionID returns the MatchWeekSubmission that matches the provided id
func (m *MatchWeekResultRepo) GetBySubmissionID(ctx context.Context, submissionID uuid.UUID) (*domain.MatchWeekResult, error) {
	// TODO: feat - implement repo method
	return nil, nil
}

// Insert the provided MatchWeekResult into the database
func (m *MatchWeekResultRepo) Insert(ctx context.Context, mwResult *domain.MatchWeekResult) error {
	// TODO: feat - implement repo method
	return nil
}

// Update the provided MatchWeekResult by its submission id
func (m *MatchWeekResultRepo) Update(ctx context.Context, mwResult *domain.MatchWeekResult) error {
	// TODO: feat - implement repo method
	return nil
}

// NewMatchWeekResultRepo instantiates a new MatchWeekSubmissionRepo with the provided attributes
func NewMatchWeekResultRepo(db *sql.DB, idFn idFunc, timeFn timeFunc) (*MatchWeekResultRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("db: %w", domain.ErrIsNil)
	}
	if idFn == nil {
		idFn = uuid.NewUUID
	}
	if timeFn == nil {
		timeFn = time.Now
	}
	return &MatchWeekResultRepo{
		db:     db,
		idFn:   idFn,
		timeFn: timeFn,
	}, nil
}
