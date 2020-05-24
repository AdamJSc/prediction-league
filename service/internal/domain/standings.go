package domain

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"time"
)

// StandingsAgentInjector defines the dependencies required by our StandingsAgent
type StandingsAgentInjector interface {
	MySQL() coresql.Agent
}

// StandingsAgent defines the behaviours for handling Standings
type StandingsAgent struct{ StandingsAgentInjector }

// CreateStandings handles the creation of a new Standings in the database
func (s StandingsAgent) CreateStandings(ctx Context, standings models.Standings) (models.Standings, error) {
	db := s.MySQL()

	// generate a new entry ID
	id, err := uuid.NewV4()
	if err != nil {
		return models.Standings{}, InternalError{err}
	}

	// override these values
	standings.ID = id
	standings.CreatedAt = time.Time{}
	standings.UpdatedAt = sqltypes.NullTime{}

	standingsRepo := repositories.NewStandingsDatabaseRepository(db)

	// write entry to database
	if err := standingsRepo.Insert(ctx, &standings); err != nil {
		return models.Standings{}, domainErrorFromDBError(err)
	}

	return standings, nil
}

// RetrieveStandingsByID handles the retrieval of an existing Standings in the database by its ID
func (s StandingsAgent) RetrieveStandingsByID(ctx Context, id string) (models.Standings, error) {
	standingsRepo := repositories.NewStandingsDatabaseRepository(s.MySQL())

	retrievedStandings, err := standingsRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return models.Standings{}, domainErrorFromDBError(err)
	}

	return retrievedStandings[0], nil
}

// RetrieveStandingsBySeasonAndRoundNumber handles the retrieval of an existing Standings in the database by its Season ID and Round Number
func (s StandingsAgent) RetrieveStandingsBySeasonAndRoundNumber(ctx Context, seasonID string, roundNumber int) (models.Standings, error) {
	standingsRepo := repositories.NewStandingsDatabaseRepository(s.MySQL())

	retrievedStandings, err := standingsRepo.Select(ctx, map[string]interface{}{
		"season_id":    seasonID,
		"round_number": roundNumber,
	}, false)
	if err != nil {
		return models.Standings{}, domainErrorFromDBError(err)
	}

	return retrievedStandings[0], nil
}

// UpdateStandings handles the updating of an existing Standings in the database
func (s StandingsAgent) UpdateStandings(ctx Context, standings models.Standings) (models.Standings, error) {
	standingsRepo := repositories.NewStandingsDatabaseRepository(s.MySQL())

	// ensure the entry exists
	if err := standingsRepo.ExistsByID(ctx, standings.ID.String()); err != nil {
		return models.Standings{}, domainErrorFromDBError(err)
	}

	// write to database
	if err := standingsRepo.Update(ctx, &standings); err != nil {
		return models.Standings{}, domainErrorFromDBError(err)
	}

	return standings, nil
}
