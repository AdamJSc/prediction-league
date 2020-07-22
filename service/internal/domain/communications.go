package domain

import (
	"bytes"
	"fmt"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/emails"
	"prediction-league/service/internal/messages"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/views"
)

// CommunicationsAgentInjector defines the dependencies required by our CommunicationsAgent
type CommunicationsAgentInjector interface {
	Config() Config
	EmailQueue() chan messages.Email
	Template() *views.Templates
}

// CommunicationsAgent defines the behaviours for issuing communications
type CommunicationsAgent struct{ CommunicationsAgentInjector }

// IssueNewEntryEmail generates a "new entry" email for the provided Entry and pushes it to the send queue
func (c CommunicationsAgent) IssuesNewEntryEmail(entry *models.Entry) error {
	realm, ok := c.Config().Realms[entry.RealmName]
	if !ok {
		return fmt.Errorf("realm does not exist: %s", entry.RealmName)
	}

	season, err := datastore.Seasons.GetByID(entry.SeasonID)
	if err != nil {
		return err
	}

	d := emails.NewEntryEmailData{
		Name:           entry.EntrantName,
		SeasonName:     season.Name,
		PredictionsURL: fmt.Sprintf("%s/prediction", realm.Origin),
		ShortCode:      entry.ShortCode,
		SignOff:        realm.Contact.Name,
		URL:            realm.Origin,
		SupportEmail:   realm.Contact.EmailProper,
	}
	var emailContent bytes.Buffer
	if err := c.Template().ExecuteTemplate(&emailContent, "email_txt_new_entry", d); err != nil {
		return err
	}

	email := messages.Email{
		From: messages.Identity{
			Name:    "", // TODO - add to realm
			Address: "", // TODO - add to realm
		},
		To: messages.Identity{
			Name:    entry.EntrantName,
			Address: entry.EntrantEmail,
		},
		ReplyTo: messages.Identity{
			Name:    "", // TODO - add to realm
			Address: realm.Contact.EmailProper,
		},
		Subject:   "You're in!",
		PlainText: emailContent.String(),
	}

	c.EmailQueue() <- email

	return nil
}
