package mailgun

import (
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"

	mggo "github.com/mailgun/mailgun-go/v3"
)

const mailgunSuccessMessage = "Queued. Thank you."

// Client defines our Mailgun API client
type Client struct {
	apiKey string
}

// SendEmail implements this method on the clients.EmailClient interface
func (c *Client) SendEmail(ctx context.Context, msg domain.Email) error {
	mg := mggo.NewMailgun(msg.SenderDomain, c.apiKey)
	mg.SetAPIBase(mggo.APIBaseEU)

	mgMsg := mg.NewMessage(
		fmt.Sprintf("%s <%s>", msg.From.Name, msg.From.Address),
		msg.Subject,
		msg.PlainText,
		fmt.Sprintf("%s <%s>", msg.To.Name, msg.To.Address),
	)
	mgMsg.SetTracking(false)
	mgMsg.SetReplyTo(fmt.Sprintf("%s <%s>", msg.ReplyTo.Name, msg.ReplyTo.Address))

	result, id, err := mg.Send(ctx, mgMsg)
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
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key: %w", domain.ErrIsEmpty)
	}

	return &Client{
		apiKey: apiKey,
	}, nil
}
