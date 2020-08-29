package mailgun

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/messages"

	"github.com/mailgun/mailgun-go/v3"
)

const mailgunSuccessMessage = "Queued. Thank you."

// Client defines our Mailgun API client
type Client struct {
	apiKey string
}

// SendEmail implements this method on the clients.EmailClient interface
func (c *Client) SendEmail(ctx context.Context, message messages.Email) error {
	mg := mailgun.NewMailgun(message.SenderDomain, c.apiKey)
	mg.SetAPIBase(mailgun.APIBaseEU)

	msg := mg.NewMessage(
		fmt.Sprintf("%s <%s>", message.From.Name, message.From.Address),
		message.Subject,
		message.PlainText,
		fmt.Sprintf("%s <%s>", message.To.Name, message.To.Address),
	)
	msg.SetTracking(true)
	msg.SetReplyTo(fmt.Sprintf("%s <%s>", message.ReplyTo.Name, message.ReplyTo.Address))

	result, id, err := mg.Send(ctx, msg)
	if err != nil {
		return err
	}
	if result != mailgunSuccessMessage {
		return fmt.Errorf("send message result: %s", result)
	}
	if id == "" {
		return errors.New("send id empty")
	}

	return nil
}

// NewClient generates a new Client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}
