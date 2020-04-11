package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/core/rest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const (
	envServicePort = "SERVICE_PORT"
	envMySQLURL    = "MYSQL_URL"
)

func main() {
	// setup
	mustLoadEnv([]string{envServicePort})
	db := coresql.MustOpen("mysql", os.Getenv(envMySQLURL))

	r := mux.NewRouter()
	r.HandleFunc("/{anything}", func(w http.ResponseWriter, r *http.Request) {
		pathVal, ok := mux.Vars(r)["anything"]
		if !ok {
			rest.JSONError("vars went wrong").WriteTo(w)
			return
		}

		rows, err := db.Query("SHOW DATABASES")
		if err != nil {
			rest.JSONError(err).WriteTo(w)
			return
		}

		var results []string
		var result string

		for rows.Next() {
			if err := rows.Scan(&result); err != nil {
				rest.JSONError(err).WriteTo(w)
				return
			}
			results = append(results, result)
		}

		rest.OKResponse(&rest.Data{
			Type:    pathVal,
			Content: results,
		}, nil).WriteTo(w)
	})

	port := os.Getenv(envServicePort)
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func mustLoadEnv(requiredKeys []string) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	for _, key := range requiredKeys {
		if os.Getenv(key) == "" {
			log.Fatal(fmt.Errorf("missing env var '%s'", key))
		}
	}
}
