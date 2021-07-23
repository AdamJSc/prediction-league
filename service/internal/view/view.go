package view

import (
	"html/template"
	"prediction-league/service/internal/domain"
	"time"
)

type Base struct {
	Title                 string
	BannerTitle           template.HTML
	ActivePage            string
	IsLoggedIn            bool
	SupportEmailPlainText string
	SupportEmailFormatted string
	RealmName             string
	RunningVersion        string
	VersionTimestamp      string
	AnalyticsCode         string
	Data                  interface{}
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
		ID        string
		ShortCode string // TODO - ShortCode: migrate auth to use session token
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
	FAQs []FAQItem
}

type FAQItem struct {
	Question string
	Answer   template.HTML
}

type GenerateMagicLoginPageData struct {
	Err       error
	EmailAddr string
}
