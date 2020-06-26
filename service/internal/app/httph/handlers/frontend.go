package handlers

import (
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/pages"
)

func frontendIndexHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p = pages.Base{
			Title:      "Home",
			ActivePage: "home",
			IsLoggedIn: isLoggedIn(r),
		}

		if err := c.Template().ExecuteTemplate(w, "index", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendResultsHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p = pages.Base{
			Title:      "Results",
			ActivePage: "results",
			IsLoggedIn: isLoggedIn(r),
		}

		if err := c.Template().ExecuteTemplate(w, "results", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendFAQHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p = pages.Base{
			Title:      "FAQ",
			ActivePage: "faq",
			IsLoggedIn: isLoggedIn(r),
		}

		if err := c.Template().ExecuteTemplate(w, "faq", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendEnterHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p = pages.Base{
			Title:      "Enter",
			ActivePage: "enter",
			IsLoggedIn: isLoggedIn(r),
		}

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
		setAuthCookieValue("7bee64af-7768-40f2-bed3-690786962304", w, r) // arbitrary seeded user
		w.Header().Set("Location", "/selection")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func frontendSelectionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn := isLoggedIn(r)

		var writeResponse = func(data pages.SelectionPageData, loggedIn bool) {
			var p = pages.Base{
				Title:      "Update My Selection",
				ActivePage: "selection",
				IsLoggedIn: loggedIn,
				Data:       data,
			}

			if err := c.Template().ExecuteTemplate(w, "selection", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.SelectionPageData{Err: err}, loggedIn)
			return
		}
		defer cancel()

		data := getSelectionPageData(
			ctx,
			getAuthCookieValue(r),
			domain.EntryAgent{EntryAgentInjector: c},
			domain.TokenAgent{TokenAgentInjector: c})

		writeResponse(data, loggedIn)
	}
}
