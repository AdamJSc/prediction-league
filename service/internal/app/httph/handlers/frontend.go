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
		p := newPage(r, c, "Home", "home", nil)

		if err := c.Template().ExecuteTemplate(w, "index", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendLeaderBoardHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data pages.LeaderBoardPageData) {
			p := newPage(r, c, "Leaderboard", "leaderboard", data)

			if err := c.Template().ExecuteTemplate(w, "leaderboard", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.LeaderBoardPageData{Err: err})
			return
		}
		defer cancel()

		data := getLeaderBoardPageData(
			ctx,
			domain.EntryAgent{EntryAgentInjector: c},
			domain.StandingsAgent{StandingsAgentInjector: c},
			domain.LeaderBoardAgent{LeaderBoardAgentInjector: c},
		)

		writeResponse(data)
	}
}

func frontendFAQHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data pages.FAQPageData) {
			p := newPage(r, c, "FAQ", "faq", data)

			if err := c.Template().ExecuteTemplate(w, "faq", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.FAQPageData{Err: err})
			return
		}
		defer cancel()

		data := getFAQPageData(domain.RealmFromContext(ctx).Name)

		writeResponse(data)
	}
}

func frontendJoinHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data := pages.JoinPageData{
			PayPalClientID: c.Config().PayPalClientID,
		}

		p := newPage(r, c, "Join", "join", data)

		if err := c.Template().ExecuteTemplate(w, "join", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendPredictionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data pages.PredictionPageData) {
			p := newPage(r, c, "Update My Prediction", "prediction", data)

			if err := c.Template().ExecuteTemplate(w, "prediction", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(pages.PredictionPageData{Err: err})
			return
		}
		defer cancel()

		data := getPredictionPageData(
			ctx,
			getAuthCookieValue(r),
			domain.EntryAgent{EntryAgentInjector: c},
			domain.TokenAgent{TokenAgentInjector: c},
		)

		writeResponse(data)
	}
}
