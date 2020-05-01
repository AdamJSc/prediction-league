package handlers

import (
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/app/httph"
)

func frontendIndexHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Template().ExecuteTemplate(w, "index", struct{}{}); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}
