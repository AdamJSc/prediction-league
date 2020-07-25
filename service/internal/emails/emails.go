package emails

type EmailData struct {
	Name         string
	SignOff      string
	SeasonName   string
	URL          string
	SupportEmail string
}

type NewEntryEmailData struct {
	EmailData
	PredictionsURL string
	ShortCode      string
}

type RoundCompleteEmailData struct {
	EmailData
	RoundNumber       int
	RankingsAsStrings []string
	LeaderBoardURL    string
}
