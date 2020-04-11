package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/LUSHDigital/core/rest"

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
		pathVal, ok := mux.Vars(r)["anything"]
		if !ok {
			rest.JSONError("vars went wrong").WriteTo(w)
			return
		}
		rest.OKResponse(&rest.Data{
			Type:    pathVal,
			Content: "Hello World!",
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
