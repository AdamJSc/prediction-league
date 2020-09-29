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

type Window struct {
	Current            int
	Total              int
	IsLast             bool
	CurrentClosingDate string
	CurrentClosingTime string
	NextOpeningDate    string
	NextOpeningTime    string
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

type ShortCodeResetBeginEmail struct {
	EmailData
	ResetURL string
}

type ShortCodeResetCompleteEmail struct {
	EmailData
	PredictionsURL string
	ShortCode      string
}

type PredictionWindowEmail struct {
	EmailData
	Window         Window
	PredictionsURL string
}
