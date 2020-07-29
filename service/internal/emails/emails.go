package emails

type EmailData struct {
	Name         string
	SignOff      string
	SeasonName   string
	URL          string
	SupportEmail string
}

type PaymentDetails struct {
	Amount       string
	Reference    string
	MerchantName string
}

type NewEntryEmailData struct {
	EmailData
	PaymentDetails PaymentDetails
	PredictionsURL string
	ShortCode      string
}

type RoundCompleteEmailData struct {
	EmailData
	RoundNumber       int
	RankingsAsStrings []string
	LeaderBoardURL    string
}
