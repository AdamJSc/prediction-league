package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MatchWeekResult represents the scored result of the associated match week submission
type MatchWeekResult struct {
	MatchWeekSubmissionID uuid.UUID           // id of the associated match week submission (one result per submission)
	TeamRankings          []ResultTeamRanking // represents the team rankings of the associated match week submission, along with actual standings position and points "hit" for each team
	Score                 int64               // overall score attributed to the match week result, after all modifiers have been applied
	Modifiers             []ModifierSummary   // summary of all modifiers used to affect the score (replaying the value of these modifiers should provide us with the same overall score each time)
	CreatedAt             time.Time           // date that result was created
	UpdatedAt             *time.Time          // date that result was most recently updated, if applicable
}

// ResultTeamRanking associates a team ranking with the position and "hit" value calculated from an associated mw standings object
type ResultTeamRanking struct {
	TeamRanking         // team id + position from the original mw submission
	StandingsPos uint16 // team's "actual" position within the standings
	Hit          int64  // resulting points "hit" (absolute difference between the team's submitted ranking position and standings position)
}

// ModifierSummary represents the modifiers applied to a particular mw result
type ModifierSummary struct {
	Code  ModifierCode // arbitrary code representing the modifier
	Value int64        // value applied to the overall score by the modifier
}

// ModifierCode defines a predefined set of values to be used on a ModifierSummary
type ModifierCode string

const (
	BaseScoreModifierCode       ModifierCode = "BASE_SCORE"
	TeamRankingsHitModifierCode ModifierCode = "RANKINGS_HIT"
)

// MatchWeekResultModifier defines a function which modifies the provided mw result object in some way (i.e. affects the overall score)
//
// Each modifier function *should* also apply a ModifierSummary to the provided mw result object, so that the modifiers can be "replayed" if ever required.
type MatchWeekResultModifier func(result *MatchWeekResult) error

// NewMatchWeekResult returns a new MatchWeekResult object that has been enriched by the provided modifiers
func NewMatchWeekResult(mwSubmissionID uuid.UUID, modifiers ...MatchWeekResultModifier) (*MatchWeekResult, error) {
	result := &MatchWeekResult{
		MatchWeekSubmissionID: mwSubmissionID,
	}

	for _, modifier := range modifiers {
		if modifier == nil {
			continue
		}
		if err := modifier(result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// BaseScoreModifier overrides the match week result's score with the provided value
func BaseScoreModifier(score int64) MatchWeekResultModifier {
	return func(result *MatchWeekResult) error {
		result.Score = score
		result.Modifiers = append(result.Modifiers, ModifierSummary{
			Code:  BaseScoreModifierCode,
			Value: score,
		})

		return nil
	}
}

// TeamRankingsHitModifier populates the match week result's team rankings based on the scores produced
// by comparing the rankings of the provided submission and standings objects
func TeamRankingsHitModifier(submission *MatchWeekSubmission, standings *MatchWeekStandings) MatchWeekResultModifier {
	return func(result *MatchWeekResult) error {
		if submission == nil || standings == nil {
			return nil
		}

		// calculate "hit" values for each submission ranking
		resultRankings, totalHits, err := getResultTeamRankings(submission, standings)
		if err != nil {
			return err
		}

		result.TeamRankings = resultRankings
		result.Score = result.Score - totalHits // deduct total hits from current score
		result.Modifiers = append(result.Modifiers, ModifierSummary{
			Code:  TeamRankingsHitModifierCode,
			Value: -totalHits,
		})

		return nil
	}
}

// getResultTeamRankings returns each of the submission's team rankings enriched with a "hit" value
// (the number of points to deduct from the overall score) based on the provided standings
func getResultTeamRankings(submission *MatchWeekSubmission, standings *MatchWeekStandings) ([]ResultTeamRanking, int64, error) {
	// ensure that both sets of rankings have the same number of entries
	submissionCount := len(submission.TeamRankings)
	standingsCount := len(standings.TeamRankings)
	if submissionCount != standingsCount {
		return nil, 0, fmt.Errorf("rankings count mismatch: submission %d: standings %d", submissionCount, standingsCount)
	}

	// TODO: move de-duplicating of submission team rankings to entity creation
	// check both sets of rankings for duplicates
	submissionRankings := submission.TeamRankings
	if err := checkForDuplicateTeamRankings(submissionRankings); err != nil {
		return nil, 0, fmt.Errorf("submission team rankings: %w", err)
	}

	// TODO: move de-duplicating of standings team rankings to entity creation
	standingsRankings := getTeamRankingsfromStandingsTeamRankings(standings.TeamRankings)
	if err := checkForDuplicateTeamRankings(standingsRankings); err != nil {
		return nil, 0, fmt.Errorf("standings team rankings: %w", err)
	}

	// TODO: move sorting of submission team rankings to entity creation
	// ensure both sets of rankings are sorted by position ascending
	sort.SliceStable(submission.TeamRankings, func(i, j int) bool {
		return submission.TeamRankings[i].Position < submission.TeamRankings[j].Position
	})

	// TODO: move sorting of standings team rankings to entity creation
	sort.SliceStable(standings.TeamRankings, func(i, j int) bool {
		return standings.TeamRankings[i].Position < standings.TeamRankings[j].Position
	})

	// map each standings ranking by its team id (so we can access them directly while cycling through the submission rankings)
	standRankMap := make(map[string]StandingsTeamRanking)
	for _, standRank := range standings.TeamRankings {
		standRankMap[standRank.TeamID] = standRank
	}

	missingTeamIDs := make([]string, 0)
	resultRankings := make([]ResultTeamRanking, 0)
	var totalHits int64 = 0

	// populate hits for submission rankings based on standings rankings
	for _, subRank := range submission.TeamRankings {
		resultRank := ResultTeamRanking{
			TeamRanking: subRank,
		}

		// get the standings ranking for the current submission ranking team id
		standRank, ok := standRankMap[subRank.TeamID]
		if !ok {
			// log team id as missing from standings rankings, and move onto next submission ranking
			missingTeamIDs = append(missingTeamIDs, subRank.TeamID)
			continue
		}

		subRankHit := calculateHit(subRank, standRank)

		resultRank.StandingsPos = standRank.Position
		resultRank.Hit = subRankHit
		totalHits = totalHits + subRankHit

		resultRankings = append(resultRankings, resultRank)
	}

	if len(missingTeamIDs) > 0 {
		return nil, 0, fmt.Errorf("team ids missing from standings rankings: '%s'", strings.Join(missingTeamIDs, "', '"))
	}

	return resultRankings, totalHits, nil
}

// calculateHit returns the difference between the positions of the two provided team rankings as a positive integer
func calculateHit(submissionRanking TeamRanking, standingsRanking StandingsTeamRanking) int64 {
	diff := int16(submissionRanking.Position) - int16(standingsRanking.Position)
	abs := math.Abs(float64(diff))
	return int64(abs)
}
