package handlers

import "prediction-league/service/internal/domain"

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
	EntryID     string `json:"entry_id"`
}
