package repofac

import (
	"github.com/LUSHDigital/core-sql"
	"prediction-league/service/internal/repositories"
)

// NewEntryDatabaseRepository instantiates a new EntryDatabaseRepository with the provided DB agent
func NewEntryDatabaseRepository(db coresql.Agent) *repositories.EntryDatabaseRepository {
	return &repositories.EntryDatabaseRepository{Agent: db}
}

