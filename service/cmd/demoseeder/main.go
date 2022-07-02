package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

	db, err := sql.Open("mysql", config.MySQLURL)
	if err != nil {
		return fmt.Errorf("cannot open mysql connection: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT * FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("cannot execute db command: %w", err)
	}

	rowCount := 0
	for rows.Next() {
		rowCount++
		var (
			version uint8
			dirty   bool
		)
		if err := rows.Scan(&version, &dirty); err != nil {
			return fmt.Errorf("cannot scan row #%d: %w", rowCount, err)
		}
		log.Printf("row #%d: version = %d, dirty = %t", rowCount, version, dirty)
	}

	return nil
}

// spec defines the config state for the script
type spec struct {
	MySQLURL string `envconfig:"MYSQL_URL" required:"true"`
}
