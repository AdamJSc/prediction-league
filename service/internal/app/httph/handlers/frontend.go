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

func frontendLoginHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// temporary proxy for setting a cookie and redirecting to edit page
		// in reality this will be a post route that verifies login credentials
		setAuthCookieValue(w, "7bee64af-7768-40f2-bed3-690786962304", stripPort(r.Host)) // arbitrary seeded user
		w.Header().Set("Location", "/selection")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func frontendSelectionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	var p = pages.Base{
		Title: "Make Selection",
		Menu:  pages.InflatedMenu("home"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		p.Data = getSelectionPageData(r, c)

		if err := c.Template().ExecuteTemplate(w, "edit", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}
