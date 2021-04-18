package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	repofac2 "prediction-league/service/internal/repositories/repofac"
	"prediction-league/service/internal/views"
)

const (
	EmailSubjectNewEntry                     = "You're In!"
	EmailSubjectRoundComplete                = "End of Round %d"
	EmailSubjectShortCodeResetBegin          = "Resetting your Short Code"
	EmailSubjectShortCodeResetComplete       = "Your new Short Code"
	EmailSubjectPredictionWindowOpen         = "Prediction Window Open!"
	EmailSubjectPredictionWindowOpenFinal    = "Final Prediction Window Open!"
	EmailSubjectPredictionWindowClosing      = "Last chance to revise your Prediction!"
	EmailSubjectPredictionWindowClosingFinal = "Final chance to revise your Prediction this season!"
)

// CommunicationsAgentInjector defines the dependencies required by our CommunicationsAgent
type CommunicationsAgentInjector interface {
	Config() Config
	MySQL() coresql.Agent
	EmailQueue() chan Email
	Template() *views.Templates
}

// CommunicationsAgent defines the behaviours for issuing communications
type CommunicationsAgent struct{ CommunicationsAgentInjector }

// IssueNewEntryEmail generates a "new entry" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssueNewEntryEmail(_ context.Context, entry *Entry, paymentDetails *PaymentDetails) error {
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

	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	d := NewEntryEmailData{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		PaymentDetails: *paymentDetails,
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
		ShortCode:      entry.ShortCode,
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_new_entry", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectNewEntry, emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// IssueRoundCompleteEmail generates a "round complete" email for the provided Scored Entry Prediction and pushes it to the send queue
func (c CommunicationsAgent) IssueRoundCompleteEmail(ctx context.Context, sep *ScoredEntryPrediction, finalRound bool) error {
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

	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	rankingsAsStrings, err := TeamRankingsAsStrings(sep.Rankings, standings.Rankings)
	if err != nil {
		return err
	}

	d := RoundCompleteEmailData{
		MessagePayload:    newMessagePayload(realm, entry.EntrantName, season.Name),
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

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, fmt.Sprintf(EmailSubjectRoundComplete, standings.RoundNumber), emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// IssueShortCodeResetBeginEmail generates a "short code reset begin" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssueShortCodeResetBeginEmail(_ context.Context, entry *Entry, resetToken string) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return NotFoundError{fmt.Errorf("realm does not exist: %s", entry.RealmName)}
	}

	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	d := ShortCodeResetBeginEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		ResetURL:       fmt.Sprintf("%s/reset/%s", realm.Origin, resetToken),
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_short_code_reset_begin", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectShortCodeResetBegin, emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// IssueShortCodeResetCompleteEmail generates a "short code reset complete" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssueShortCodeResetCompleteEmail(_ context.Context, entry *Entry) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return NotFoundError{fmt.Errorf("realm does not exist: %s", entry.RealmName)}
	}

	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	d := ShortCodeResetCompleteEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
		ShortCode:      entry.ShortCode,
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_short_code_reset_complete", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectShortCodeResetComplete, emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// IssuePredictionWindowOpenEmail generates a "prediction window open" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssuePredictionWindowOpenEmail(_ context.Context, entry *Entry, tf SequencedTimeFrame) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return NotFoundError{fmt.Errorf("realm does not exist: %s", entry.RealmName)}
	}

	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	window, err := GenerateWindowDataFromSequencedTimeFrame(tf)
	if err != nil {
		return InternalError{err}
	}

	d := PredictionWindowEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		Window:         *window,
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_prediction_window_open", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}

	subject := EmailSubjectPredictionWindowOpen
	if window.IsLast {
		subject = EmailSubjectPredictionWindowOpenFinal
	}

	email := newEmail(realm, recipient, subject, emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// IssuePredictionWindowClosingEmail generates a "prediction window closing" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssuePredictionWindowClosingEmail(_ context.Context, entry *Entry, tf SequencedTimeFrame) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return NotFoundError{fmt.Errorf("realm does not exist: %s", entry.RealmName)}
	}

	season, err := SeasonsDataStore.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{err}
	}

	window, err := GenerateWindowDataFromSequencedTimeFrame(tf)
	if err != nil {
		return InternalError{err}
	}

	d := PredictionWindowEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		Window:         *window,
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_prediction_window_closing", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}

	subject := EmailSubjectPredictionWindowClosing
	if window.IsLast {
		subject = EmailSubjectPredictionWindowClosingFinal
	}

	email := newEmail(realm, recipient, subject, emailContent.String())
	c.EmailQueue() <- email

	return nil
}

// getEntryFromScoredEntryPrediction retrieves the relationally-affiliated entry from the provided scored entry prediction
func getEntryFromScoredEntryPrediction(ctx context.Context, sep *ScoredEntryPrediction, db coresql.Agent) (*Entry, error) {
	entryPredictionRepo := repofac2.NewEntryPredictionDatabaseRepository(db)
	entryRepo := repofac2.NewEntryDatabaseRepository(db)

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
func getStandingsFromScoredEntryPrediction(ctx context.Context, sep *ScoredEntryPrediction, db coresql.Agent) (*Standings, error) {
	standingsRepo := repofac2.NewStandingsDatabaseRepository(db)

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

// GenerateWindowDataFromSequencedTimeFrame generates an email WindowData object from the provided SequencedTimeFrame
func GenerateWindowDataFromSequencedTimeFrame(tf SequencedTimeFrame) (*WindowData, error) {
	var window WindowData

	if tf.Current == nil {
		return nil, ErrCurrentTimeFrameIsMissing
	}

	window.Current = tf.Count
	window.Total = tf.Total

	window.CurrentClosingDate = tf.Current.Until.Format("Mon 2 January")
	window.CurrentClosingTime = tf.Current.Until.Format("3:04pm")

	if tf.Next != nil {
		window.NextOpeningDate = tf.Next.From.Format("Mon 2 January")
		window.NextOpeningTime = tf.Next.From.Format("3:04pm")
	}

	if tf.Count == tf.Total {
		window.IsLast = true
	}

	return &window, nil
}

// Identity defines a combination of name and address
type Identity struct {
	Name    string
	Address string
}

// Email defines the properties of an email message
type Email struct {
	From         Identity
	To           Identity
	ReplyTo      Identity
	SenderDomain string
	Subject      string
	PlainText    string
}

// newEmail returns an email message object inflated with the provided data items
func newEmail(realm Realm, to Identity, subject, plainText string) Email {
	return Email{
		From: Identity{
			Name:    realm.Contact.Name,
			Address: realm.Contact.EmailDoNotReply,
		},
		To: to,
		ReplyTo: Identity{
			Name:    realm.Contact.Name,
			Address: realm.Contact.EmailProper,
		},
		SenderDomain: realm.SenderDomain,
		Subject:      subject,
		PlainText:    plainText,
	}
}

// newMessagePayload returns an email data object inflated with the provided data items
func newMessagePayload(realm Realm, name string, seasonName string) MessagePayload {
	return MessagePayload{
		Name:         name,
		SignOff:      realm.Contact.Name,
		SeasonName:   seasonName,
		URL:          realm.Origin,
		SupportEmail: realm.Contact.EmailProper,
	}
}

// MessagePayload defines the fields required by an email message
type MessagePayload struct {
	Name         string
	SignOff      string
	SeasonName   string
	URL          string
	SupportEmail string
}

// PaymentDetails defines the fields relating to a payment
type PaymentDetails struct {
	Amount       string
	Reference    string
	MerchantName string
}

// WindowData defines the fields relating to a prediction window
type WindowData struct {
	Current            int
	Total              int
	IsLast             bool
	CurrentClosingDate string
	CurrentClosingTime string
	NextOpeningDate    string
	NextOpeningTime    string
}

// NewEntryEmailData defines the fields relating to the content of a new entry email
type NewEntryEmailData struct {
	MessagePayload
	PaymentDetails PaymentDetails
	PredictionsURL string
	ShortCode      string
}

// RoundCompleteEmailData defines the fields relating to the content of a round complete email
type RoundCompleteEmailData struct {
	MessagePayload
	RoundNumber       int
	RankingsAsStrings []string
	LeaderBoardURL    string
}

// ShortCodeResetBeginEmail defines the fields relating to the content of a short code reset begin email
type ShortCodeResetBeginEmail struct {
	MessagePayload
	ResetURL string
}

// ShortCodeResetBeginEmail defines the fields relating to the content of a short code reset begin email
type ShortCodeResetCompleteEmail struct {
	MessagePayload
	PredictionsURL string
	ShortCode      string
}

// PredictionWindowEmail defines the fields relating to the content of a prediction window email
type PredictionWindowEmail struct {
	MessagePayload
	Window         WindowData
	PredictionsURL string
}
