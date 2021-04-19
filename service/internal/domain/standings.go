package domain

import (
	"context"
	"errors"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"prediction-league/service/internal/repositories/repofac"
	"sort"
	"time"
)

// Standings provides a data type for league standings that have been retrieved from an external data source
type Standings struct {
	ID          uuid.UUID         `db:"id"`
	SeasonID    string            `db:"season_id"`
	RoundNumber int               `db:"round_number"`
	Rankings    []RankingWithMeta `db:"rankings"`
	Finalised   bool              `db:"finalised"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   sqltypes.NullTime `db:"updated_at"`
}

// These methods make the Standings struct sortable
func (s Standings) Len() int      { return len(s.Rankings) }
func (s Standings) Swap(i, j int) { s.Rankings[i], s.Rankings[j] = s.Rankings[j], s.Rankings[i] }
func (s Standings) Less(i, j int) bool {
	return s.Rankings[i].Position < s.Rankings[j].Position
}

// StandingsAgentInjector defines the dependencies required by our StandingsAgent
type StandingsAgentInjector interface {
	MySQL() coresql.Agent
}

// StandingsAgent defines the behaviours for handling Standings
type StandingsAgent struct { StandingsAgentInjector }

// CreateStandings handles the creation of a new Standings in the database
func (s StandingsAgent) CreateStandings(ctx context.Context, standings Standings) (Standings, error) {
	db := s.MySQL()

	// generate a new entry ID
	id, err := uuid.NewV4()
	if err != nil {
		return Standings{}, InternalError{err}
	}

	// override these values
	standings.ID = id
	standings.CreatedAt = time.Now().Truncate(time.Second)
	standings.UpdatedAt = sqltypes.NullTime{}

	standingsRepo := repofac.NewStandingsDatabaseRepository(db)

	// write entry to database
	if err := standingsRepo.Insert(ctx, &standings); err != nil {
		return Standings{}, domainErrorFromRepositoryError(err)
	}

	return standings, nil
}

// RetrieveStandingsByID handles the retrieval of an existing Standings in the database by its ID
func (s StandingsAgent) RetrieveStandingsByID(ctx context.Context, id string) (Standings, error) {
	standingsRepo := repofac.NewStandingsDatabaseRepository(s.MySQL())

	retrievedStandings, err := standingsRepo.Select(ctx, map[string]interface{}{
		"id": id,
	}, false)
	if err != nil {
		return Standings{}, domainErrorFromRepositoryError(err)
	}

	return retrievedStandings[0], nil
}

// RetrieveStandingsBySeasonAndRoundNumber handles the retrieval of an existing Standings in the database by its Season ID and Round Number
func (s StandingsAgent) RetrieveStandingsBySeasonAndRoundNumber(ctx context.Context, seasonID string, roundNumber int) (Standings, error) {
	standingsRepo := repofac.NewStandingsDatabaseRepository(s.MySQL())

	retrievedStandings, err := standingsRepo.Select(ctx, map[string]interface{}{
		"season_id":    seasonID,
		"round_number": roundNumber,
	}, false)
	if err != nil {
		return Standings{}, domainErrorFromRepositoryError(err)
	}

	return retrievedStandings[0], nil
}

// RetrieveLatestStandingsBySeasonIDAndTimestamp handles the retrieval of the latest Standings in the database by its ID
func (s StandingsAgent) RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx context.Context, seasonID string, ts time.Time) (Standings, error) {
	standingsRepo := repofac.NewStandingsDatabaseRepository(s.MySQL())

	retrievedStandings, err := standingsRepo.SelectLatestBySeasonIDAndTimestamp(ctx, seasonID, ts)
	if err != nil {
		return Standings{}, domainErrorFromRepositoryError(err)
	}

	return retrievedStandings, nil
}

// UpdateStandings handles the updating of an existing Standings in the database
func (s StandingsAgent) UpdateStandings(ctx context.Context, standings Standings) (Standings, error) {
	standingsRepo := repofac.NewStandingsDatabaseRepository(s.MySQL())

	// ensure the entry exists
	if err := standingsRepo.ExistsByID(ctx, standings.ID.String()); err != nil {
		return Standings{}, domainErrorFromRepositoryError(err)
	}

	// override these values
	standings.UpdatedAt = sqltypes.ToNullTime(time.Now().Truncate(time.Second))

	// write to database
	if err := standingsRepo.Update(ctx, &standings); err != nil {
		return Standings{}, domainErrorFromRepositoryError(err)
	}

	return standings, nil
}

// RetrieveStandingsIfNotFinalised provides a helpful wrapper method to return the standings represented by the provided
// season ID and round number if not finalised, otherwise returns the provided default standings
func (s StandingsAgent) RetrieveStandingsIfNotFinalised(
	ctx context.Context,
	seasonID string,
	roundNumber int,
	defaultStandings Standings,
) (Standings, error) {
	standings, err := s.RetrieveStandingsBySeasonAndRoundNumber(
		ctx,
		seasonID,
		roundNumber,
	)
	if err != nil {
		switch err.(type) {
		case NotFoundError:
			// unable to find standings for provided season ID and round number
			// fallback to default standings
			return defaultStandings, nil
		default:
			// something went wrong whilst retrieving our previous standings...
			return Standings{}, err
		}
	}

	if !standings.Finalised {
		// return standings that have not yet been finalised
		return standings, nil
	}

	// fallback to default standings
	return defaultStandings, nil
}

// ValidateAndSortStandings sorts and validates the provided standings
func ValidateAndSortStandings(standings *Standings) error {
	if standings == nil {
		return InternalError{errors.New("standings not provided")}
	}

	// ensure that all team IDs are valid
	for _, ranking := range standings.Rankings {
		if _, err := TeamsDataStore.GetByID(ranking.ID); err != nil {
			return NotFoundError{err}
		}
	}

	// default standings sort (ascending by Rankings[].Position)
	sort.Sort(standings)

	return nil
}
