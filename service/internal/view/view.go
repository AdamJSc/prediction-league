package view

import (
	"html/template"
	"prediction-league/service/internal/domain"
	"time"
)

type Base struct {
	PageTitle          string
	BannerTitle        template.HTML
	ActivePage         string
	IsLoggedIn         bool
	Realm              *domain.Realm
	SeasonName         string
	HomePageURL        string
	LeaderBoardPageURL string
	JoinPageURL        string
	FAQPageURL         string
	PredictionPageURL  string
	LoginPageURL       string
	BuildVersion       string
	BuildTimestamp     string
	Data               interface{}
}

type PredictionPageData struct {
	Err         error
	Predictions struct {
		Status        string
		IsClosing     bool
		Limit         int
		AcceptedFrom  time.Time
		AcceptedUntil time.Time
	}
	Teams struct {
		Raw         string
		LastUpdated time.Time
	}
	Entry struct {
		ID              string
		PredictionToken string
	}
}

type LeaderBoardPageData struct {
	Err         error
	RoundNumber int
	Season      struct {
		ID       string
		RawTeams string
	}
	Entries struct {
		RawEntries  string
		RawRankings string
	}
	LastUpdated time.Time
}

type JoinPageData struct {
	EntriesOpen     bool
	EntriesOpenTS   time.Time
	EntriesClosed   bool
	EntriesClosedTS time.Time
	SeasonName      string
	PayPalClientID  string
	EntryFee        domain.RealmEntryFee
}

type FAQPageData struct {
	Err  error
	FAQs []domain.RealmFAQ
}

type GenerateMagicLoginPageData struct {
	Err       error
	EmailAddr string
}
