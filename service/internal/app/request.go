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
	PaymentMethod     string `json:"payment_method"`
	PaymentRef        string `json:"payment_ref"`
	PaymentAmount     string `json:"payment_amount"`
	MerchantName      string `json:"merchant_name"`
	RegistrationToken string `json:"reg_token"`
}

type createEntryPredictionRequest struct {
	PredictionToken string   `json:"entry_pred_token"`
	RankingIDs     []string `json:"ranking_ids"`
}

func (r createEntryPredictionRequest) ToEntryPredictionModel() domain.EntryPrediction {
	return domain.NewEntryPrediction(r.RankingIDs)
}

type generateMagicLoginRequest struct {
	EmailAddr string
}
