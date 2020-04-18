package handlers

import (
	"log"
	"net/http"
)

func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}
