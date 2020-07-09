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
		// leaderboard can't be generated, a blank one will be returned
	case nil:
		// we've got a valid leaderboard
		rawRankings, err := json.Marshal(leaderBoard.Rankings)
		if err != nil {
			data.Err = err
			return data
		}
		data.RoundNumber = leaderBoard.RoundNumber
		data.RawRankings = string(rawRankings)
		data.LastUpdated = *leaderBoard.LastUpdated
	default:
		// something else went wrong so display the error message
		data.Err = err
	}

	// retrieve entries
	entries, err := entryAgent.RetrieveEntriesBySeasonID(ctx, seasonID)
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
	data.RawEntries = string(rawEntries)

	return data
}
