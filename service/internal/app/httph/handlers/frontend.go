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

func frontendLeaderBoardHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn := isLoggedIn(r)

		var writeResponse = func(data pages.LeaderBoardPageData, loggedIn bool) {
			var p = pages.Base{
				Title:      "Leaderboard",
				ActivePage: "leaderboard",
				IsLoggedIn: loggedIn,
				Data:       data,
			}

			if err := c.Template().ExecuteTemplate(w, "leaderboard", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.LeaderBoardPageData{Err: err}, loggedIn)
			return
		}
		defer cancel()

		data := getLeaderBoardPageData(
			ctx,
			domain.EntryAgent{EntryAgentInjector: c},
			domain.StandingsAgent{StandingsAgentInjector: c},
			domain.LeaderBoardAgent{LeaderBoardAgentInjector: c},
		)

		writeResponse(data, loggedIn)
	}
}

func frontendFAQHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn := isLoggedIn(r)

		var writeResponse = func(data pages.FAQPageData, loggedIn bool) {
			var p = pages.Base{
				Title:      "FAQ",
				ActivePage: "faq",
				IsLoggedIn: loggedIn,
				Data:       data,
			}

			if err := c.Template().ExecuteTemplate(w, "faq", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.FAQPageData{Err: err}, loggedIn)
			return
		}
		defer cancel()

		data := getFAQPageData(domain.RealmFromContext(ctx).Name)

		writeResponse(data, loggedIn)
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
