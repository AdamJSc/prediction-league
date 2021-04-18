package handlers

import (
	"context"
	"encoding/json"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/pages"
)

func getLeaderBoardPageData(ctx context.Context, entryAgent domain.EntryAgent, standingsAgent domain.StandingsAgent, leaderBoardAgent domain.LeaderBoardAgent) pages.LeaderBoardPageData {
	var data pages.LeaderBoardPageData

	ctxRealm := domain.RealmFromContext(ctx)

	// retrieve realm's current season ID
	seasonID := ctxRealm.SeasonID
	data.Season.ID = seasonID

	// retrieve round number
	roundNumber := 1
	latestStandings, err := standingsAgent.RetrieveLatestStandingsBySeasonIDAndTimestamp(ctx, seasonID, domain.TimestampFromContext(ctx))
	switch err.(type) {
	case domain.NotFoundError:
	// do nothing, defaults to round number 1
	case nil:
		// no error, so let's get the round number
		roundNumber = latestStandings.RoundNumber
	default:
		// something else went wrong...
		data.Err = err
		return data
	}

	// retrieve latest leaderboard
	leaderBoard, err := leaderBoardAgent.RetrieveLeaderBoardBySeasonAndRoundNumber(ctx, seasonID, roundNumber)
	switch err.(type) {
	case domain.NotFoundError:
		// leaderboard can't be generated, an empty one will be returned
	case nil:
		// we've got a valid leaderboard
		rawRankings, err := json.Marshal(leaderBoard.Rankings)
		if err != nil {
			data.Err = err
			return data
		}
		data.RoundNumber = leaderBoard.RoundNumber
		data.Entries.RawRankings = string(rawRankings)
		if leaderBoard.LastUpdated != nil {
			data.LastUpdated = *leaderBoard.LastUpdated
		}
	default:
		// something else went wrong so display the error message
		data.Err = err
	}

	// retrieve entries
	entries, err := entryAgent.RetrieveEntriesBySeasonID(ctx, seasonID, true)
	if err != nil {
		switch err.(type) {
		case domain.NotFoundError:
			// we have a blank entry set, no more work needed so return early
			return data
		case nil:
		// no issues, drop through and carry on...
		default:
			// something else went wrong...
			data.Err = err
			return data
		}
	}

	// only retain entries if they belong to the current realm
	mappedEntries := make(map[string]string)
	for _, entry := range entries {
		if entry.RealmName == ctxRealm.Name {
			mappedEntries[entry.ID.String()] = entry.EntrantNickname
		}
	}
	rawEntries, err := json.Marshal(mappedEntries)
	if err != nil {
		data.Err = err
		return data
	}
	data.Entries.RawEntries = string(rawEntries)

	// finally, populate teams for season
	season, err := domain.SeasonsDataStore.GetByID(seasonID)
	if err != nil {
		data.Err = err
		return data
	}
	var teams []domain.Team
	for _, id := range season.TeamIDs {
		team, err := domain.TeamsDataStore.GetByID(id)
		if err != nil {
			data.Err = err
			return data
		}

		teams = append(teams, team)
	}
	rawTeams, err := json.Marshal(teams)
	if err != nil {
		data.Err = err
		return data
	}
	data.Season.RawTeams = string(rawTeams)

	return data
}
