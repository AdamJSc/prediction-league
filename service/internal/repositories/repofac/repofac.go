package repofac

import (
	"github.com/LUSHDigital/core-sql"
	"prediction-league/service/internal/repositories"
)

// NewEntryPredictionDatabaseRepository instantiates a new EntryPredictionDatabaseRepository with the provided DB agent
func NewEntryPredictionDatabaseRepository(db coresql.Agent) *repositories.EntryPredictionDatabaseRepository {
	return &repositories.EntryPredictionDatabaseRepository{Agent: db}
}

// NewEntryDatabaseRepository instantiates a new EntryDatabaseRepository with the provided DB agent
func NewEntryDatabaseRepository(db coresql.Agent) *repositories.EntryDatabaseRepository {
	return &repositories.EntryDatabaseRepository{Agent: db}
}

// NewScoredEntryPredictionDatabaseRepository instantiates a new ScoredEntryPredictionDatabaseRepository with the provided DB agent
func NewScoredEntryPredictionDatabaseRepository(db coresql.Agent) *repositories.ScoredEntryPredictionDatabaseRepository {
	return &repositories.ScoredEntryPredictionDatabaseRepository{Agent: db}
}
