package handlers

import (
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/pages"
)

func frontendIndexHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	var p = pages.Base{
		Title: "Home",
		Menu:  pages.InflatedMenu("home"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Template().ExecuteTemplate(w, "index", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendResultsHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	var p = pages.Base{
		Title: "Results",
		Menu:  pages.InflatedMenu("results"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Template().ExecuteTemplate(w, "results", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendFAQHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	var p = pages.Base{
		Title: "FAQ",
		Menu:  pages.InflatedMenu("faq"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Template().ExecuteTemplate(w, "faq", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendEnterHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	var p = pages.Base{
		Title: "Enter",
		Menu:  pages.InflatedMenu("enter"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Template().ExecuteTemplate(w, "enter", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}
