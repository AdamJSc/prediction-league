package seeder

import (
	"context"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core-sql/sqltypes"
	"github.com/LUSHDigital/uuid"
	"log"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/models"
	"prediction-league/service/internal/repositories"
	"time"
)

var entries = []models.Entry{
	{
		EntrantName:     "Adam S",
		EntrantNickname: "AdamS",
		EntrantEmail:    "seeded1@example.net",
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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
		EntryPredictions: []models.EntryPrediction{
			models.NewEntryPrediction([]string{
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

// MustSeed inserts the existing entries for the 2019/20 Season in the localhost Realm
func MustSeed(db coresql.Agent) {
	ctx := context.Background()
	entryRepo := repositories.NewEntryDatabaseRepository(db)
	entryPredictionRepo := repositories.NewEntryPredictionDatabaseRepository(db)

	seasonID := "FakeSeason"
	realmName := "localhost"
	paymentMethod := models.EntryPaymentMethodOther
	paymentRef := "payment_ref"
	approvedAt := time.Now()

	for _, entry := range entries {
		shortCode, err := domain.GenerateUniqueShortCode(ctx, db)
		if err != nil {
			log.Fatal(err)
		}

		entry.ID = uuid.Must(uuid.NewV4())
		entry.ShortCode = shortCode
		entry.SeasonID = seasonID
		entry.RealmName = realmName
		entry.Status = models.EntryStatusReady
		entry.PaymentMethod = sqltypes.ToNullString(&paymentMethod)
		entry.PaymentRef = sqltypes.ToNullString(&paymentRef)
		entry.ApprovedAt = sqltypes.ToNullTime(approvedAt)

		if err := entryRepo.Insert(context.Background(), &entry); err != nil {
			switch err.(type) {
			case repositories.DuplicateDBRecordError:
				// already seeded, so we can fail silently
				continue
			}
			log.Fatal(err)
		}

		for _, entryPrediction := range entry.EntryPredictions {
			entryPrediction.ID = uuid.Must(uuid.NewV4())
			entryPrediction.EntryID = entry.ID

			if err := entryPredictionRepo.Insert(context.Background(), &entryPrediction); err != nil {
				switch err.(type) {
				case repositories.DuplicateDBRecordError:
					// already seeded, so we can fail silently
					continue
				}
				log.Fatal(err)
			}
		}
	}
}
