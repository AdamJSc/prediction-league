package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}

func getRouteParam(r *http.Request, name string, value *string) error {
	val, ok := mux.Vars(r)[name]
	if !ok {
		return fmt.Errorf("invalid param: %s", name)
	}
	*value = val
	return nil
}
