package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	// numberOfMatchWeeks to generate seed data for
	numberOfMatchWeeks uint8 = 10

	// realmName to use when generating seeded entries
	realmName = "localhost"

	// seasonID denotes which season to use to randomise team rankings from
	seasonID = domain.FakeSeasonID

	// weekDuration represents one week
	weekDuration = time.Hour * 24 * 7
)

var (
	// baseTimestamp to use as a basis for running the job
	baseTimestamp = time.Date(2018, 5, 26, 14, 0, 0, 0, time.FixedZone("Europe/London", 3600))

	// entrantNicknames to generate entries and predictions on behalf of
	entrantNicknames = []string{
		"Fletch",
		"Puuuugh",
		"Brett",
		"Cookie",
		"Tommy",
		"HarryA",
		"Franno",
		"CharlieD",
		"Ritchie",
		"Shwan",
	}

	// randomiser instantiates a random number generator
	randomiser = rand.New(rand.NewSource(time.Now().UnixNano()))

	// topSixTeamIDs should be randomised separately to the remaining set of IDs
	topSixTeamIDs = []string{
		"AFC",  // arsenal
		"CFC",  // chelsea
		"LFC",  // liverpool
		"MCFC", // man city
		"MUFC", // man utd
		"THFC", // spurs
	}
)

func main() {
	log.Println("run started...")

	if err := run(); err != nil {
		log.Fatalf("run failed: %s", err.Error())
	}

	log.Println("run succeeded!")
}

// spec defines the config schema
type spec struct {
	MySQLURL string `envconfig:"MYSQL_URL" required:"true"`
}

func run() error {
	// retrieve season by provided id
	seasonCollection, err := domain.GetSeasonCollection()
	if err != nil {
		return fmt.Errorf("cannot retrieve season collection: %w", err)
	}
	season, err := seasonCollection.GetByID(seasonID)
	if err != nil {
		return fmt.Errorf("cannot retrieve season: %w", err)
	}

	// parse env
	_, currentFilename, _, _ := runtime.Caller(1)
	envPath := filepath.Dir(currentFilename) + "/.env"
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("error loading .env: %s", err.Error())
		log.Println("continuing anyway...")
	}

	// parse config from env
	config := &spec{}
	if err := envconfig.Process("", config); err != nil {
		return fmt.Errorf("cannot parse config: %w", err)
	}

	// connect to db and instantiate repos
	db, err := sql.Open("mysql", config.MySQLURL)
	if err != nil {
		return fmt.Errorf("cannot open mysql connection: %w", err)
	}
	defer db.Close()

	entryRepo, err := mysqldb.NewEntryRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate new entry repo: %w", err)
	}

	entryPredictionRepo, err := mysqldb.NewEntryPredictionRepo(db)
	if err != nil {
		return fmt.Errorf("cannot instantiate new entry prediction repo: %w", err)
	}

	// run job
	j := &job{
		realmName:           realmName,
		season:              season,
		entryRepo:           entryRepo,
		entryPredictionRepo: entryPredictionRepo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := j.process(ctx); err != nil {
		return fmt.Errorf("cannot process job: %w", err)
	}

	return nil
}

type job struct {
	realmName           string
	season              domain.Season
	entryRepo           *mysqldb.EntryRepo
	entryPredictionRepo *mysqldb.EntryPredictionRepo
}

func (j *job) process(ctx context.Context) error {
	// generate entries
	entries := make([]*domain.Entry, 0)
	for _, nickName := range entrantNicknames {
		entries = append(entries, j.generateEntry(entryParams{
			entrantNickname:     nickName,
			timestamp:           baseTimestamp,
			numberOfPredictions: numberOfMatchWeeks,
		}))
	}

	// insert entries and entry predictions
	for _, entry := range entries {
		// insert entry
		if err := j.entryRepo.Insert(ctx, entry); err != nil {
			return fmt.Errorf("cannot generate entry: %w", err)
		}

		for _, entryPrediction := range entry.EntryPredictions {
			// insert entry prediction
			if err := j.entryPredictionRepo.Insert(ctx, &entryPrediction); err != nil {
				return fmt.Errorf("cannot generate entry prediction: %w", err)
			}
		}
	}

	log.Printf("%#v", entries)

	return nil
}

type entryParams struct {
	entrantNickname     string
	timestamp           time.Time
	numberOfPredictions uint8
}

func (j *job) generateEntry(p entryParams) *domain.Entry {
	lowercaseNickname := strings.ToLower(p.entrantNickname)

	name := fmt.Sprintf("Mr Sir %s Esq.", p.entrantNickname)
	email := fmt.Sprintf("%s@demo.com", lowercaseNickname)
	paymentMethod := domain.EntryPaymentMethodOther
	paymentRef := fmt.Sprintf("%s_payment_ref", lowercaseNickname)
	createdAt := p.timestamp
	approvedAt := p.timestamp.Add(time.Second) // one second after creation

	entryID := uuid.New()

	entryPredictions := make([]domain.EntryPrediction, 0)
	for i := uint8(1); i <= p.numberOfPredictions; i++ {
		predictionTimestamp := p.timestamp.Add(time.Duration(i) * weekDuration) // provided timestamp + one week per iteration
		entryPredictions = append(entryPredictions, j.generateEntryPrediction(entryPredictionParams{
			entryID:   entryID,
			timestamp: predictionTimestamp,
		}))
	}

	return &domain.Entry{
		ID:               entryID,
		SeasonID:         j.season.ID,
		RealmName:        j.realmName,
		EntrantName:      name,
		EntrantNickname:  p.entrantNickname,
		EntrantEmail:     email,
		Status:           domain.EntryStatusPaid,
		PaymentMethod:    &paymentMethod,
		PaymentRef:       &paymentRef,
		EntryPredictions: entryPredictions,
		ApprovedAt:       &approvedAt,
		CreatedAt:        createdAt,
	}
}

type entryPredictionParams struct {
	entryID   uuid.UUID
	timestamp time.Time
}

func (j *job) generateEntryPrediction(p entryPredictionParams) domain.EntryPrediction {
	return domain.EntryPrediction{
		ID:        uuid.New(),
		EntryID:   p.entryID,
		Rankings:  j.generateFullRandomRankingCollection(),
		CreatedAt: p.timestamp,
	}
}

func (j *job) generateFullRandomRankingCollection() domain.RankingCollection {
	collection := make(domain.RankingCollection, 0)

	topSix := randomiseStringsOrder(topSixTeamIDs)
	remainder := stringsDiff(j.season.TeamIDs, topSix)
	remainder = randomiseStringsOrder(remainder)
	concatTeamIDs := append(topSix, remainder...)

	for idx, teamID := range concatTeamIDs {
		collection = append(collection, domain.Ranking{
			ID:       teamID,
			Position: idx + 1,
		})
	}

	return collection
}

// stringsDiff returns a slice of strings that appear within the full slice, but not the subset
func stringsDiff(full, subset []string) []string {
	diff := make([]string, 0)

	isInSubset := func(input string) bool {
		for _, s := range subset {
			if input == s {
				return true
			}
		}
		return false
	}

	for _, f := range full {
		if !isInSubset(f) {
			diff = append(diff, f)
		}
	}

	return diff
}

// randomiseStringsOrder returns a copy of the provided strings sorted into a random order
func randomiseStringsOrder(input []string) []string {
	copied := make([]string, 0)
	for _, s := range input {
		copied = append(copied, s)
	}

	sort.SliceStable(copied, func(i, j int) bool {
		return randomBool()
	})

	return copied
}

// randomBool returns a randomised true/false value
func randomBool() bool {
	randNum := randomiser.Intn(2) // either 0 or 1
	return randNum == 1
}
