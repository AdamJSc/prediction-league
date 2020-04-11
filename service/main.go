package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
)

const (
	envServicePort = "SERVICE_PORT"
)

func main() {
	mustLoadEnv([]string{envServicePort})

	r := mux.NewRouter()
	r.HandleFunc("/{anything}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	port := os.Getenv(envServicePort)
	fmt.Printf("Listening on port %s...", port)
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
