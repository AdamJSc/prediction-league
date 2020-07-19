package sendgrid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prediction-league/service/internal/models"

	"github.com/sendgrid/sendgrid-go"
)

const baseURL = "https://api.sendgrid.com"

// Client defines our SendGrid API client
type Client struct {
	apiKey string
}

// SendEmail implements this method on the clients.EmailClient interface
func (c *Client) SendEmail(message models.EmailMessage) error {
	requestBody := transformEmailMessageToSendMailRequest(message)
	marshaledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	request := sendgrid.GetRequest(c.apiKey, "/v3/mail/send", baseURL)
	request.Method = http.MethodPost
	request.Body = marshaledRequestBody

	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("send email failed: %d %s", response.StatusCode, response.Body)
	}

	return nil
}

// NewClient generates a new Client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

type sendMailContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendMailEmailName struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type sendMailRequestPersonalization struct {
	To      []sendMailEmailName `json:"to"`
	Subject string
}

type sendMailRequest struct {
	Personalizations []sendMailRequestPersonalization `json:"personalizations"`
	From             sendMailEmailName                `json:"from"`
	ReplyTo          sendMailEmailName                `json:"reply_to"`
	Content          []sendMailContent                `json:"content"`
}

func transformEmailMessageToSendMailRequest(message models.EmailMessage) *sendMailRequest {
	return &sendMailRequest{
		Personalizations: []sendMailRequestPersonalization{
			{
				To: []sendMailEmailName{
					{
						Email: message.To.Address,
						Name:  message.To.Name,
					},
				},
				Subject: message.Subject,
			},
		},
		From: sendMailEmailName{
			Email: message.From.Address,
			Name:  message.From.Name,
		},
		ReplyTo: sendMailEmailName{
			Email: message.ReplyTo.Address,
			Name:  message.ReplyTo.Name,
		},
		Content: []sendMailContent{
			{
				Type:  "text/plain",
				Value: message.PlainText,
			},
			{
				Type:  "text/html",
				Value: message.HTML,
			},
		},
	}
}
