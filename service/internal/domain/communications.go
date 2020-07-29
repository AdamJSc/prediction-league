package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/emails"
	"prediction-league/service/internal/messages"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"prediction-league/service/internal/views"
)

const (
	EmailSubjectNewEntry      = "You're In!"
	EmailSubjectRoundComplete = "End of Round %d"
)

// CommunicationsAgentInjector defines the dependencies required by our CommunicationsAgent
type CommunicationsAgentInjector interface {
	Config() Config
	MySQL() coresql.Agent
	EmailQueue() chan messages.Email
	Template() *views.Templates
}

// CommunicationsAgent defines the behaviours for issuing communications
type CommunicationsAgent struct{ CommunicationsAgentInjector }

// IssueNewEntryEmail generates a "new entry" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssueNewEntryEmail(_ context.Context, entry *models.Entry, paymentDetails *emails.PaymentDetails) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	if paymentDetails == nil {
		return InternalError{errors.New("no payment details provided")}
	}

	if paymentDetails.Amount == "" || paymentDetails.Reference == "" || paymentDetails.MerchantName == "" {
		return ValidationError{Reasons: []string{"invalid payment details"}}
	}

	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return NotFoundError{fmt.Errorf("realm does not exist: %s", entry.RealmName)}
	}

	season, err := datastore.Seasons.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	d := emails.NewEntryEmailData{
		EmailData:      newEmailData(realm, entry.EntrantName, season.Name),
		PaymentDetails: *paymentDetails,
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
		ShortCode:      entry.ShortCode,
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_new_entry", d); err != nil {
		return err
	}

	recipient := messages.Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectNewEntry, emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// IssueRoundCompleteEmail generates a "round complete" email for the provided Scored Entry Prediction and pushes it to the send queue
func (c CommunicationsAgent) IssueRoundCompleteEmail(ctx context.Context, sep *models.ScoredEntryPrediction, finalRound bool) error {
	entry, err := getEntryFromScoredEntryPrediction(ctx, sep, c.MySQL())
	if err != nil {
		return err
	}

	standings, err := getStandingsFromScoredEntryPrediction(ctx, sep, c.MySQL())
	if err != nil {
		return err
	}

	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return NotFoundError{fmt.Errorf("realm does not exist: %s", entry.RealmName)}
	}

	season, err := datastore.Seasons.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	rankingsAsStrings, err := TeamRankingsAsStrings(sep.Rankings, standings.Rankings)
	if err != nil {
		return err
	}

	d := emails.RoundCompleteEmailData{
		EmailData:         newEmailData(realm, entry.EntrantName, season.Name),
		RoundNumber:       standings.RoundNumber,
		RankingsAsStrings: rankingsAsStrings,
		LeaderBoardURL:    fmt.Sprintf("%s/leaderboard", realm.Origin),
	}

	templateName := "email_txt_round_complete"
	if finalRound {
		templateName = "email_txt_final_round_complete"
	}

	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, templateName, d); err != nil {
		return err
	}

	recipient := messages.Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, fmt.Sprintf(EmailSubjectRoundComplete, standings.RoundNumber), emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// getEntryFromScoredEntryPrediction retrieves the relationally-affiliated entry from the provided scored entry prediction
func getEntryFromScoredEntryPrediction(ctx context.Context, sep *models.ScoredEntryPrediction, db coresql.Agent) (*models.Entry, error) {
	entryPredictionRepo := repositories.NewEntryPredictionDatabaseRepository(db)
	entryRepo := repositories.NewEntryDatabaseRepository(db)

	// retrieve entry prediction from scored entry prediction
	entryPredictions, err := entryPredictionRepo.Select(ctx, map[string]interface{}{
		"id": sep.EntryPredictionID,
	}, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}
	if len(entryPredictions) != 1 {
		return nil, ConflictError{fmt.Errorf("expected 1 entry prediction, found %d", len(entryPredictions))}
	}

	// retrieve entry from entry prediction
	entries, err := entryRepo.Select(ctx, map[string]interface{}{
		"id": entryPredictions[0].EntryID,
	}, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}
	if len(entries) != 1 {
		return nil, ConflictError{fmt.Errorf("expected 1 entry, found %d", len(entries))}
	}

	return &entries[0], nil
}

// getStandingsFromScoredEntryPrediction retrieves the relationally-affiliated standings from the provided scored entry prediction
func getStandingsFromScoredEntryPrediction(ctx context.Context, sep *models.ScoredEntryPrediction, db coresql.Agent) (*models.Standings, error) {
	standingsRepo := repositories.NewStandingsDatabaseRepository(db)

	// retrieve standings from scored entry prediction
	standings, err := standingsRepo.Select(ctx, map[string]interface{}{
		"id": sep.StandingsID,
	}, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}
	if len(standings) != 1 {
		return nil, ConflictError{fmt.Errorf("expected 1 standings, found %d", len(standings))}
	}

	return &standings[0], nil
}

// newEmail returns an email message object inflated with the provided data items
func newEmail(realm Realm, to messages.Identity, subject, plainText string) messages.Email {
	return messages.Email{
		From: messages.Identity{
			Name:    realm.Contact.Name,
			Address: realm.Contact.EmailDoNotReply,
		},
		To: to,
		ReplyTo: messages.Identity{
			Name:    realm.Contact.Name,
			Address: realm.Contact.EmailProper,
		},
		Subject:   subject,
		PlainText: plainText,
	}
}

// newEmailData returns an email data object inflated with the provided data items
func newEmailData(realm Realm, name string, seasonName string) emails.EmailData {
	return emails.EmailData{
		Name:         name,
		SignOff:      realm.Contact.Name,
		SeasonName:   seasonName,
		URL:          realm.Origin,
		SupportEmail: realm.Contact.EmailProper,
	}
}
