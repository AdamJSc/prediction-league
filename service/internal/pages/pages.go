package pages

import (
	"html/template"
	"time"
)

type Base struct {
	Title            string
	ActivePage       string
	IsLoggedIn       bool
	RunningVersion   string
	VersionTimestamp string
	Data             interface{}
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
	PayPalClientID string
}

type FAQPageData struct {
	Err  error
	FAQs []FAQItem
}

type FAQItem struct {
	Question string
	Answer   template.HTML
}
