package pages

import (
	"time"
)

type Base struct {
	Title      string
	ActivePage string
	IsLoggedIn bool
	Data       interface{}
}

type PredictionPageData struct {
	Err        error
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
