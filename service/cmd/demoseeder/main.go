package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"prediction-league/service/internal/adapters/mysqldb"
	"prediction-league/service/internal/domain"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	log.Println("run started...")

	if err := run(); err != nil {
		log.Fatalf("run failed: %s", err.Error())
	}

	log.Println("run succeeded!")
}

func run() error {
	// parse command line arguments
	var seasonID, realmName string
	flag.StringVar(&seasonID, "season", "", "id of season to seed entities against")
	flag.StringVar(&realmName, "realm", "", "name of realm to seed entities against")
	flag.Parse()

	if seasonID == "" {
		return fmt.Errorf("-season flag must be provided")
	}
	if realmName == "" {
		return fmt.Errorf("-realm flag must be provided")
	}

	// retrieve season by provided id
	seasonCollection, err := domain.GetSeasonCollection()
	if err != nil {
		return fmt.Errorf("cannot retrieve season collection: %w", err)
	}
	season, err := seasonCollection.GetByID(seasonID)
	if err != nil {
		return fmt.Errorf("cannot retrieve season: %w", err)
	}

	teams := domain.GetTeamCollection()

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
		teams:               teams,
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

var entrantNicknames = []string{
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

var timestamp = time.Date(2018, 5, 26, 14, 0, 0, 0, time.FixedZone("Europe/London", 3600))

type job struct {
	realmName           string
	season              domain.Season
	teams               domain.TeamCollection
	entryRepo           *mysqldb.EntryRepo
	entryPredictionRepo *mysqldb.EntryPredictionRepo
}

func (j *job) process(ctx context.Context) error {
	// generate entries
	entries := make([]*domain.Entry, 0)
	for _, nickName := range entrantNicknames {
		entries = append(entries, j.generateEntry(entryParams{
			entrantNickname: nickName,
			timestamp:       timestamp,
		}))
	}

	// insert entries
	for _, entry := range entries {
		if err := j.entryRepo.Insert(ctx, entry); err != nil {
			return fmt.Errorf("cannot generate entry: %w", err)
		}
	}

	return nil
}

type entryParams struct {
	entrantNickname string
	timestamp       time.Time
}

func (j *job) generateEntry(p entryParams) *domain.Entry {
	lowercaseNickname := strings.ToLower(p.entrantNickname)

	name := fmt.Sprintf("Sir Mr %s Esq.", p.entrantNickname)
	email := fmt.Sprintf("%s@demo.com", lowercaseNickname)
	paymentMethod := domain.EntryPaymentMethodOther
	paymentRef := fmt.Sprintf("%s_payment_ref", lowercaseNickname)
	createdAt := p.timestamp
	approvedAt := p.timestamp.Add(time.Second) // one second after creation

	entryID := uuid.New()

	return &domain.Entry{
		ID:              entryID,
		SeasonID:        j.season.ID,
		RealmName:       j.realmName,
		EntrantName:     name,
		EntrantNickname: p.entrantNickname,
		EntrantEmail:    email,
		Status:          domain.EntryStatusPaid,
		PaymentMethod:   &paymentMethod,
		PaymentRef:      &paymentRef,
		ApprovedAt:      &approvedAt,
		CreatedAt:       createdAt,
		// don't set EntryPredictions yet
	}
}

// spec defines the config schema
type spec struct {
	MySQLURL string `envconfig:"MYSQL_URL" required:"true"`
}
