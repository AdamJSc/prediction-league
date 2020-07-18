package pages

import (
	"html/template"
	"prediction-league/service/internal/domain"
	"time"
)

type Base struct {
	Title                 string
	ActivePage            string
	IsLoggedIn            bool
	SupportEmailPlainText string
	RealmName             string
	RunningVersion        string
	VersionTimestamp      string
	Data                  interface{}
}

type PredictionPageData struct {
	Err         error
	Predictions struct {
		BeingAccepted    bool
		NextAcceptedFrom *time.Time
		AcceptedUntil    *time.Time
	}
	Teams struct {
		Raw         string
		LastUpdated time.Time
	}
	Entry struct {
		ID        string
		ShortCode string
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
	SupportEmailFormatted string
	PayPalClientID        string
	EntryFee              domain.RealmEntryFee
}

type FAQPageData struct {
	Err  error
	FAQs []FAQItem
}

type FAQItem struct {
	Question string
	Answer   template.HTML
}
