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

func frontendLeaderboardHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p = pages.Base{
			Title:      "Leaderboard",
			ActivePage: "leaderboard",
			IsLoggedIn: isLoggedIn(r),
		}

		if err := c.Template().ExecuteTemplate(w, "leaderboard", p); err != nil {
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

func frontendJoinHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var p = pages.Base{
			Title:      "Join",
			ActivePage: "join",
			IsLoggedIn: isLoggedIn(r),
		}

		if err := c.Template().ExecuteTemplate(w, "join", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendPredictionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn := isLoggedIn(r)

		var writeResponse = func(data pages.PredictionPageData, loggedIn bool) {
			var p = pages.Base{
				Title:      "Update My Prediction",
				ActivePage: "prediction",
				IsLoggedIn: loggedIn,
				Data:       data,
			}

			if err := c.Template().ExecuteTemplate(w, "prediction", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.PredictionPageData{Err: err}, loggedIn)
			return
		}
		defer cancel()

		data := getPredictionPageData(
			ctx,
			getAuthCookieValue(r),
			domain.EntryAgent{EntryAgentInjector: c},
			domain.TokenAgent{TokenAgentInjector: c},
		)

		writeResponse(data, loggedIn)
	}
}
