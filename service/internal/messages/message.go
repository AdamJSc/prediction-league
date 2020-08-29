package messages

type Identity struct {
	Name    string
	Address string
}

type Email struct {
	From         Identity
	To           Identity
	ReplyTo      Identity
	SenderDomain string
	Subject      string
	PlainText    string
}
