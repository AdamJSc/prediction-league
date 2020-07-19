package models

type MessageIdentity struct {
	Name    string
	Address string
}

type EmailMessage struct {
	From      MessageIdentity
	To        MessageIdentity
	ReplyTo   MessageIdentity
	HTML      string
	PlainText string
}
