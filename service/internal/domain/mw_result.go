package domain

import (
	"time"

	"github.com/google/uuid"
)

// MatchWeekResult represents the scored result of the associated match week submission
type MatchWeekResult struct {
	MatchWeekSubmissionID uuid.UUID            // id of the associated match week submission (one result per submission)
	TeamRankings          []TeamRankingWithHit // represents the team rankings of the associated match week submission, along with actual position of each associated team and each points "hit"
	Score                 int64                // overall score attributed to the match week result, after all modifiers have been applied
	Modifiers             []ModifierSummary    // summary of all modifiers used to affect the score (replaying the value of these modifiers should provide us with the same overall score each time)
	CreatedAt             time.Time            // date that result was created
	UpdatedAt             *time.Time           // date that result was most recently updated, if applicable
}

// TeamRankingWithHit associates a team ranking with the values calculated from an associated mw standings object
type TeamRankingWithHit struct {
	SubmittedRanking TeamRanking // ranking associated with the original mw submission
	StandingsPos     uint16      // team's position within an associated mw standings object
	Hit              int64       // resulting points "hit" (absolute difference between the ranking position and standings position)
}

// ModifierSummary represents the modifiers applied to a particular mw result
type ModifierSummary struct {
	Code  string // arbitrary code representing the modifier
	Value int64  // value applied to the overall score by the modifier
}

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
