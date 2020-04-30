package handlers

import (
	"errors"
	"github.com/LUSHDigital/core/rest"
	"github.com/gorilla/mux"
	"net/http"
	"prediction-league/service/internal/app/httph"
)

func frontendGreetingHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		greeting, ok := mux.Vars(r)["greeting"]
		if !ok {
			rest.InternalError(errors.New("no greeting var /shrug")).WriteTo(w)
		}

		var page = struct {
			Greeting string
		}{
			Greeting: greeting,
		}

		if err := c.Template().ExecuteTemplate(w, "index", page); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}
