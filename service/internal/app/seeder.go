package app

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"prediction-league/service/internal/domain"
	"time"
)

// TODO - implement as Service with Worker interface (Run/Halt)

// Seed runs the seeder for the app startup
func Seed(cnt *container) error {
	seeds, err := generateSeedEntries()
	if err != nil {
		return fmt.Errorf("cannot generate entries to seed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chErr := make(chan error, 1)
	chDone := make(chan struct{}, 1)
	go func() {
		if err := cnt.entryAgent.SeedEntries(ctx, seeds); err != nil {
			chErr <- err
		}
		chDone <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			chErr <- ctx.Err()
		case err := <-chErr:
			return fmt.Errorf("cannot seed entries: %w", err)
		case <-chDone:
			return nil
		}
	}
}

func generateSeedEntries() ([]domain.Entry, error) {
	entries := []domain.Entry{
		{
			EntrantName:     "Adam S",
			EntrantNickname: "AdamS",
			EntrantEmail:    "seeded1@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"THFC",
					"MUFC",
					"AFC",
					"CFC",
					"LCFC",
					"WWFC",
					"AFCB",
					"WHUFC",
					"EFC",
					"NCFC",
					"CPFC",
					"WFC",
					"BFC",
					"SUFC",
					"SFC",
					"BHAFC",
					"AVFC",
					"NUFC",
				}),
			},
		},
		{
			EntrantName:     "Ben M",
			EntrantNickname: "BenM",
			EntrantEmail:    "seeded2@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"CFC",
					"MUFC",
					"AFC",
					"THFC",
					"EFC",
					"LCFC",
					"WWFC",
					"SFC",
					"AFCB",
					"WHUFC",
					"WFC",
					"CPFC",
					"AVFC",
					"NCFC",
					"SUFC",
					"NUFC",
					"BFC",
					"BHAFC",
				})},
		},
		{
			EntrantName:     "Chris F",
			EntrantNickname: "ChrisF",
			EntrantEmail:    "seeded3@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"THFC",
					"LFC",
					"AFC",
					"MUFC",
					"CFC",
					"LCFC",
					"WHUFC",
					"EFC",
					"SFC",
					"WWFC",
					"BFC",
					"WFC",
					"AFCB",
					"AVFC",
					"NCFC",
					"NUFC",
					"CPFC",
					"SUFC",
					"BHAFC",
				})},
		},
		{
			EntrantName:     "Dan N",
			EntrantNickname: "DanN",
			EntrantEmail:    "seeded4@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"THFC",
					"AFC",
					"MUFC",
					"CFC",
					"LCFC",
					"WWFC",
					"AFCB",
					"EFC",
					"SFC",
					"WFC",
					"CPFC",
					"BFC",
					"AVFC",
					"WHUFC",
					"NUFC",
					"SUFC",
					"BHAFC",
					"NCFC",
				})},
		},
		{
			EntrantName:     "Ed T",
			EntrantNickname: "EdT",
			EntrantEmail:    "seeded5@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"AFC",
					"CFC",
					"THFC",
					"EFC",
					"MUFC",
					"LCFC",
					"WWFC",
					"WHUFC",
					"WFC",
					"CPFC",
					"AFCB",
					"SFC",
					"BFC",
					"AVFC",
					"NUFC",
					"BHAFC",
					"NCFC",
					"SUFC",
				})},
		},
		{
			EntrantName:     "Gary B",
			EntrantNickname: "GaryB",
			EntrantEmail:    "seeded6@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"LFC",
					"MCFC",
					"THFC",
					"MUFC",
					"CFC",
					"AFC",
					"LCFC",
					"EFC",
					"WHUFC",
					"AFCB",
					"AVFC",
					"WWFC",
					"WFC",
					"CPFC",
					"SFC",
					"NUFC",
					"BFC",
					"SUFC",
					"BHAFC",
					"NCFC",
				})},
		},
		{
			EntrantName:     "Nigel B",
			EntrantNickname: "NigelB",
			EntrantEmail:    "seeded7@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"LFC",
					"MCFC",
					"AFC",
					"THFC",
					"MUFC",
					"CFC",
					"EFC",
					"WHUFC",
					"WWFC",
					"LCFC",
					"SFC",
					"AFCB",
					"WFC",
					"CPFC",
					"BHAFC",
					"NUFC",
					"NCFC",
					"AVFC",
					"BFC",
					"SUFC",
				})},
		},
		{
			EntrantName:     "Ray G",
			EntrantNickname: "RayG",
			EntrantEmail:    "seeded8@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"LFC",
					"MCFC",
					"THFC",
					"CFC",
					"MUFC",
					"AFC",
					"LCFC",
					"EFC",
					"WWFC",
					"SFC",
					"WHUFC",
					"AFCB",
					"CPFC",
					"WFC",
					"BFC",
					"AVFC",
					"NUFC",
					"NCFC",
					"BHAFC",
					"SUFC",
				})},
		},
		{
			EntrantName:     "Rich L",
			EntrantNickname: "RichL",
			EntrantEmail:    "seeded9@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"THFC",
					"CFC",
					"AFC",
					"MUFC",
					"LCFC",
					"EFC",
					"WFC",
					"SFC",
					"AFCB",
					"WHUFC",
					"BFC",
					"CPFC",
					"WWFC",
					"AVFC",
					"BHAFC",
					"NUFC",
					"NCFC",
					"SUFC",
				})},
		},
		{
			EntrantName:     "Tom M",
			EntrantNickname: "TomM",
			EntrantEmail:    "seeded10@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"THFC",
					"AFC",
					"EFC",
					"CFC",
					"MUFC",
					"LCFC",
					"WWFC",
					"WHUFC",
					"WFC",
					"SFC",
					"AVFC",
					"AFCB",
					"BFC",
					"SUFC",
					"CPFC",
					"NUFC",
					"NCFC",
					"BHAFC",
				})},
		},
		{
			EntrantName:     "Trev H",
			EntrantNickname: "TrevH",
			EntrantEmail:    "seede11@example.net",
			EntryPredictions: []domain.EntryPrediction{
				domain.NewEntryPrediction([]string{
					"MCFC",
					"LFC",
					"THFC",
					"MUFC",
					"AFC",
					"CFC",
					"EFC",
					"LCFC",
					"WWFC",
					"AFCB",
					"AVFC",
					"WFC",
					"WHUFC",
					"SFC",
					"SUFC",
					"CPFC",
					"NUFC",
					"BHAFC",
					"BFC",
					"NCFC",
				})},
		},
	}

	seasonID := "FakeSeason"
	realmName := "localhost"
	paymentMethod := domain.EntryPaymentMethodOther
	paymentRef := "payment_ref"
	approvedAt := time.Now()

	for i := range entries {
		entryID, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("cannot generate uuid: %w", err)
		}
		entryPredictionID, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("cannot generate uuid: %w", err)
		}

		e := &entries[i]
		e.ID = entryID
		e.SeasonID = seasonID
		e.RealmName = realmName
		e.Status = domain.EntryStatusReady
		e.PaymentMethod = &paymentMethod
		e.PaymentRef = &paymentRef
		e.ApprovedAt = &approvedAt

		for j := range e.EntryPredictions {
			ep := &e.EntryPredictions[j]
			ep.ID = entryPredictionID
			ep.EntryID = e.ID
		}
	}

	return entries, nil
}
