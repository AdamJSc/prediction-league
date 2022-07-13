package domain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
)

const (
	EmailSubjectNewEntry            = "You're In!"
	EmailSubjectRoundCompleteFormat = "Match Week %d begins!"
	EmailSubjectFinalRoundComplete  = "Thanks for playing!"
	EmailSubjectMagicLogin          = "Your login link"
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
		PredictionsURL: GetPredictionURL(&realm),
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

	d := RoundCompleteEmailData{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		RoundNumber:    standings.RoundNumber,
		LeaderBoardURL: GetLeaderBoardURL(&realm),
		PredictionsURL: GetPredictionURL(&realm),
	}

	templateName := "email_txt_round_complete"
	nextRound := standings.RoundNumber + 1
	subject := fmt.Sprintf(EmailSubjectRoundCompleteFormat, nextRound)
	if isFinalRound {
		templateName = "email_txt_final_round_complete"
		subject = EmailSubjectFinalRoundComplete
	}

	var emailContent bytes.Buffer
	if err := c.tpl.ExecuteTemplate(&emailContent, templateName, d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, subject, emailContent.String())
	if err := c.emlQ.Send(ctx, email); err != nil {
		return fmt.Errorf("cannot send email to queue: %w", err)
	}

	return nil
}

// IssueMagicLoginEmail generates a magic login email for the provided Entry and pushes it to the send queue
func (c *CommunicationsAgent) IssueMagicLoginEmail(ctx context.Context, entry *Entry, tokenId string) error {
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

	d := MagicLoginEmail{
		MessagePayload: newMessagePayload(realm, entry.EntrantName, season.Name),
		LoginURL:       GetMagicLoginURL(&realm, &Token{ID: tokenId}),
	}
	var emailContent bytes.Buffer
	if err := c.tpl.ExecuteTemplate(&emailContent, "email_txt_magic_login", d); err != nil {
		return err
	}

	recipient := Identity{
		Name:    entry.EntrantName,
		Address: entry.EntrantEmail,
	}
	email := newEmail(realm, recipient, EmailSubjectMagicLogin, emailContent.String())
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
			Name:    realm.Contact.SenderName,
			Address: realm.Contact.EmailDoNotReply,
		},
		To: to,
		ReplyTo: Identity{
			Name:    realm.Contact.SenderName,
			Address: realm.Contact.EmailProper,
		},
		SenderDomain: realm.Contact.SenderDomain,
		Subject:      subject,
		PlainText:    plainText,
	}
}

// newMessagePayload returns an email data object inflated with the provided data items
func newMessagePayload(realm Realm, recipientName string, seasonName string) MessagePayload {
	return MessagePayload{
		RecipientName: recipientName,
		GameName:      realm.Config.GameName,
		SignOff:       realm.Contact.SignOffName,
		SeasonName:    seasonName,
		URL:           GetHomeURL(&realm),
		SupportEmail:  realm.Contact.EmailProper,
	}
}

// MessagePayload defines the fields required by an email message
type MessagePayload struct {
	RecipientName string
	GameName      string
	SignOff       string
	SeasonName    string
	URL           string
	SupportEmail  string
}

// PaymentDetails defines the fields relating to a payment
type PaymentDetails struct {
	Amount       string
	Reference    string
	MerchantName string
}

// NewEntryEmailData defines the fields relating to the content of a new entry email
type NewEntryEmailData struct {
	MessagePayload
	PaymentDetails PaymentDetails
	PredictionsURL string
}

// RoundCompleteEmailData defines the fields relating to the content of a round complete email
type RoundCompleteEmailData struct {
	MessagePayload
	RoundNumber    int
	LeaderBoardURL string
	PredictionsURL string
}

// MagicLoginEmail defines the fields relating to the content of a magic login email
type MagicLoginEmail struct {
	MessagePayload
	LoginURL string
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
