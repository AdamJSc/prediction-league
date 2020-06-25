package handlers

import (
	"prediction-league/service/internal/models"
)

type createEntryRequest struct {
	EntrantName     string `json:"entrant_name"`
	EntrantNickname string `json:"entrant_nickname"`
	EntrantEmail    string `json:"entrant_email"`
	RealmPIN        string `json:"pin"`
}

func (r createEntryRequest) ToEntryModel() models.Entry {
	return models.Entry{
		EntrantName:     r.EntrantName,
		EntrantNickname: r.EntrantNickname,
		EntrantEmail:    r.EntrantEmail,
	}
}

type updateEntryPaymentDetailsRequest struct {
	PaymentMethod string `json:"payment_method"`
	PaymentRef    string `json:"payment_ref"`
	EntryID       string `json:"entry_id"`
}

type createEntrySelectionRequest struct {
	EntryShortCode string   `json:"entry_short_code"`
	RankingIDs     []string `json:"ranking_ids"`
}

func (r createEntrySelectionRequest) ToEntrySelectionModel() models.EntrySelection {
	return models.EntrySelection{
		Rankings: models.NewRankingCollectionFromIDs(r.RankingIDs),
	}
}

type selectionLoginRequest struct {
	EmailNickname string `json:"email_nickname"`
	ShortCode     string `json:"short_code"`
}
