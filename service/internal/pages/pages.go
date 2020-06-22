package pages

import (
	"prediction-league/service/internal/models"
	"time"
)

type Base struct {
	Title      string
	ActivePage string
	IsLoggedIn bool
	Data       interface{}
}

type SelectionPageData struct {
	Err                    error
	IsAcceptingSelections  bool
	SelectionsNextAccepted *models.TimeFrame
	Teams                  struct {
		Raw         string
		LastUpdated time.Time
	}
	Entry struct {
		ID        string
		ShortCode string
	}
}
