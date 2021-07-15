package app

import (
	"prediction-league/service/internal/domain"
)

type createEntryRequest struct {
	EntrantName     string `json:"entrant_name"`
	EntrantNickname string `json:"entrant_nickname"`
	EntrantEmail    string `json:"entrant_email"`
	RealmPIN        string `json:"pin"`
}

func (r createEntryRequest) ToEntryModel() domain.Entry {
	return domain.Entry{
		EntrantName:     r.EntrantName,
		EntrantNickname: r.EntrantNickname,
		EntrantEmail:    r.EntrantEmail,
	}
}

type updateEntryPaymentDetailsRequest struct {
	PaymentMethod string `json:"payment_method"`
	PaymentRef    string `json:"payment_ref"`
	PaymentAmount string `json:"payment_amount"`
	MerchantName  string `json:"merchant_name"`
	ShortCode     string `json:"short_code"`
}

type createEntryPredictionRequest struct {
	EntryShortCode string   `json:"entry_short_code"`
	RankingIDs     []string `json:"ranking_ids"`
}

func (r createEntryPredictionRequest) ToEntryPredictionModel() domain.EntryPrediction {
	return domain.NewEntryPrediction(r.RankingIDs)
}

// TODO - feat: reduce to just email
type generateMagicLoginRequest struct {
	EmailNickname string
}
