package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

// CommunicationsAgent defines the behaviours for issuing communications
type CommunicationsAgent struct {
	er   EntryRepository
	epr  EntryPredictionRepository
	sr   StandingsRepository
	emlQ EmailQueue
	tpl  *Templates
	sc   SeasonCollection
	tc   TeamCollection
	rc   RealmCollection
}

// IssueNewEntryEmail generates a "new entry" email for the provided Entry and pushes it to the send queue
func (c *CommunicationsAgent) IssueNewEntryEmail(ctx context.Context, entry *Entry, paymentDetails *PaymentDetails) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	if paymentDetails == nil {
		return InternalError{errors.New("no payment details provided")}
	}

	if paymentDetails.Amount == "" || paymentDetails.Reference == "" || paymentDetails.MerchantName == "" {
		return ValidationError{Reasons: []string{"invalid payment details"}}
	}

	realm, err := c.rc.GetByName(entry.RealmName)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get realm with id '%s': %w", entry.RealmName, err)}
	}

	season, err := c.sc.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get season with id '%s': %w", entry.SeasonID, err)}
	}

	d := NewEntryEmailData{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		PaymentDetails: *paymentDetails,
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
		ShortCode:      entry.ShortCode,
	}
	var emailContent bytes.Buffer
	if err := c.tpl.ExecuteTemplate(&emailContent, "email_txt_new_entry", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectNewEntry, emailContent.String())
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// IssueRoundCompleteEmail generates a "round complete" email for the provided ScoredEntryPrediction and pushes it to the send queue
func (c *CommunicationsAgent) IssueRoundCompleteEmail(ctx context.Context, sep ScoredEntryPrediction, isFinalRound bool) error {
	entry, err := c.getEntryFromScoredEntryPrediction(ctx, sep)
	if err != nil {
		return err
	}

	standings, err := c.getStandingsFromScoredEntryPrediction(ctx, sep)
	if err != nil {
		return err
	}

	realm, err := c.rc.GetByName(entry.RealmName)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get realm with id '%s': %w", entry.RealmName, err)}
	}

	season, err := c.sc.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get season with id '%s': %w", entry.SeasonID, err)}
	}

	rankingsAsStrings, err := TeamRankingsAsStrings(sep.Rankings, standings.Rankings, c.tc)
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
	if isFinalRound {
		templateName = "email_txt_final_round_complete"
	}

	var emailContent bytes.Buffer
	if err := c.tpl.ExecuteTemplate(&emailContent, templateName, d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, fmt.Sprintf(EmailSubjectRoundComplete, standings.RoundNumber), emailContent.String())
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// IssueShortCodeResetBeginEmail generates a "short code reset begin" email for the provided Entry and pushes it to the send queue
func (c *CommunicationsAgent) IssueShortCodeResetBeginEmail(ctx context.Context, entry *Entry, resetToken string) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, err := c.rc.GetByName(entry.RealmName)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get realm with id '%s': %w", entry.RealmName, err)}
	}

	season, err := c.sc.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get season with id '%s': %w", entry.SeasonID, err)}
	}

	d := ShortCodeResetBeginEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		ResetURL:       fmt.Sprintf("%s/reset/%s", realm.Origin, resetToken),
	}
	var emailContent bytes.Buffer
	if err := c.tpl.ExecuteTemplate(&emailContent, "email_txt_short_code_reset_begin", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectShortCodeResetBegin, emailContent.String())
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// IssueShortCodeResetCompleteEmail generates a "short code reset complete" email for the provided Entry and pushes it to the send queue
func (c *CommunicationsAgent) IssueShortCodeResetCompleteEmail(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, err := c.rc.GetByName(entry.RealmName)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get realm with id '%s': %w", entry.RealmName, err)}
	}

	season, err := c.sc.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get season with id '%s': %w", entry.SeasonID, err)}
	}

	d := ShortCodeResetCompleteEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
		ShortCode:      entry.ShortCode,
	}
	var emailContent bytes.Buffer
	if err := c.tpl.ExecuteTemplate(&emailContent, "email_txt_short_code_reset_complete", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectShortCodeResetComplete, emailContent.String())
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// IssuePredictionWindowOpenEmail generates a "prediction window open" email for the provided Entry and pushes it to the send queue
func (c *CommunicationsAgent) IssuePredictionWindowOpenEmail(ctx context.Context, entry *Entry, tf SequencedTimeFrame) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, err := c.rc.GetByName(entry.RealmName)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get realm with id '%s': %w", entry.RealmName, err)}
	}

	season, err := c.sc.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get season with id '%s': %w", entry.SeasonID, err)}
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
	if err := c.tpl.ExecuteTemplate(&emailContent, "email_txt_prediction_window_open", d); err != nil {
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
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// IssuePredictionWindowClosingEmail generates a "prediction window closing" email for the provided Entry and pushes it to the send queue
func (c *CommunicationsAgent) IssuePredictionWindowClosingEmail(ctx context.Context, entry *Entry, tf SequencedTimeFrame) error {
	if entry == nil {
		return InternalError{errors.New("no entry provided")}
	}

	realm, err := c.rc.GetByName(entry.RealmName)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get realm with id '%s': %w", entry.RealmName, err)}
	}

	season, err := c.sc.GetByID(entry.SeasonID)
	if err != nil {
		return NotFoundError{fmt.Errorf("cannot get season with id '%s': %w", entry.SeasonID, err)}
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
	if err := c.tpl.ExecuteTemplate(&emailContent, "email_txt_prediction_window_closing", d); err != nil {
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
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// getEntryFromScoredEntryPrediction retrieves the relationally-affiliated entry from the provided scored entry prediction
func (c *CommunicationsAgent) getEntryFromScoredEntryPrediction(ctx context.Context, sep ScoredEntryPrediction) (*Entry, error) {
	// retrieve entry prediction from scored entry prediction
	entryPredictions, err := c.epr.Select(ctx, map[string]interface{}{
		"id": sep.EntryPredictionID,
	}, false)
	if err != nil {
		return nil, domainErrorFromRepositoryError(err)
	}
	if len(entryPredictions) != 1 {
		return nil, ConflictError{fmt.Errorf("expected 1 entry prediction, found %d", len(entryPredictions))}
	}

	// retrieve entry from entry prediction
	entries, err := c.er.Select(ctx, map[string]interface{}{
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
func (c *CommunicationsAgent) getStandingsFromScoredEntryPrediction(ctx context.Context, sep ScoredEntryPrediction) (*Standings, error) {
	// retrieve standings from scored entry prediction
	standings, err := c.sr.Select(ctx, map[string]interface{}{
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

// NewCommunicationsAgent returns a new CommunicationsAgent using the provided repositories
func NewCommunicationsAgent(er EntryRepository, epr EntryPredictionRepository, sr StandingsRepository, emlQ EmailQueue, tpl *Templates, sc SeasonCollection, tc TeamCollection, rc RealmCollection) (*CommunicationsAgent, error) {
	switch {
	case er == nil:
		return nil, fmt.Errorf("entry repository: %w", ErrIsNil)
	case epr == nil:
		return nil, fmt.Errorf("entry prediction repository: %w", ErrIsNil)
	case sr == nil:
		return nil, fmt.Errorf("standings repository: %w", ErrIsNil)
	case emlQ == nil:
		return nil, fmt.Errorf("email queue: %w", ErrIsNil)
	case tpl == nil:
		return nil, fmt.Errorf("teplates: %w", ErrIsNil)
	case sc == nil:
		return nil, fmt.Errorf("season collection: %w", ErrIsNil)
	case tc == nil:
		return nil, fmt.Errorf("team collection: %w", ErrIsNil)
	case rc == nil:
		return nil, fmt.Errorf("realm collection: %w", ErrIsNil)
	}

	return &CommunicationsAgent{er, epr, sr, emlQ, tpl, sc, tc, rc}, nil
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

// EmailQueue defines behaviours for sending and reading Emails on a queue
type EmailQueue interface {
	Send(ctx context.Context, eml Email) error
	Read() chan Email
	Close() error
}

// InMemEmailQueue defines an Email queue that operates in memory
type InMemEmailQueue struct{ ch chan Email }

// Send implements domain.EmailQueue
func (i *InMemEmailQueue) Send(_ context.Context, eml Email) error {
	i.ch <- eml
	return nil
}

// Read implements domain.EmailQueue
func (i *InMemEmailQueue) Read() chan Email {
	return i.ch
}

// Close implements domain.EmailQueue
func (i *InMemEmailQueue) Close() error {
	close(i.ch)
	return nil
}

func NewInMemEmailQueue() *InMemEmailQueue {
	return &InMemEmailQueue{ch: make(chan Email, 1)}
}

// EmailClient defines the interface for our email client
type EmailClient interface {
	SendEmail(ctx context.Context, em Email) error
}

// NoopEmailClient provides an EmailClient implementation that logs details of a sent email
type NoopEmailClient struct {
	l Logger
}

// SendEmail implements EmailClient
func (l *NoopEmailClient) SendEmail(_ context.Context, em Email) error {
	l.l.Infof("sent email: %+v", em)
	return nil
}

func NewNoopEmailClient(l Logger) (*NoopEmailClient, error) {
	if l == nil {
		return nil, fmt.Errorf("logger: %w", ErrIsNil)
	}
	return &NoopEmailClient{l}, nil
}
