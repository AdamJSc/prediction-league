package app

import (
	"errors"
	"fmt"
	"net/http"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
	"time"
)

func frontendIndexHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		defer cancel()

		seasonID := domain.RealmFromContext(ctx).SeasonID
		season, err := c.seasons.GetByID(seasonID)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

		bannerTitle := fmt.Sprintf("%s<br />Prediction League", season.Name)

		p := newPage(r, c, "Home", "home", bannerTitle, nil)

		if err := c.templates.ExecuteTemplate(w, "index", p); err != nil {
			internalError(err).writeTo(w)
			return
		}
	}
}

func frontendLeaderBoardHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.LeaderBoardPageData) {
			p := newPage(r, c, "Leaderboard", "leaderboard", "Leaderboard", data)

			if err := c.templates.ExecuteTemplate(w, "leaderboard", p); err != nil {
				internalError(err).writeTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(view.LeaderBoardPageData{Err: err})
			return
		}
		defer cancel()

		data := getLeaderBoardPageData(
			ctx,
			c.entryAgent,
			c.standingsAgent,
			c.lbAgent,
			c.seasons,
			c.teams,
			c.clock,
		)

		writeResponse(data)
	}
}

func frontendFAQHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.FAQPageData) {
			p := newPage(r, c, "FAQ", "faq", "FAQ", data)

			if err := c.templates.ExecuteTemplate(w, "faq", p); err != nil {
				internalError(err).writeTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		defer cancel()

		data := getFAQPageData(domain.RealmFromContext(ctx).Name)

		writeResponse(data)
	}
}

func frontendJoinHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		defer cancel()

		season, err := c.seasons.GetByID(domain.RealmFromContext(ctx).SeasonID)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}

		now := time.Now()
		entriesOpen := season.EntriesAccepted.HasBegunBy(now) && !season.EntriesAccepted.HasElapsedBy(now)
		entriesClosed := season.EntriesAccepted.HasElapsedBy(now)

		data := view.JoinPageData{
			EntriesOpen:     entriesOpen,
			EntriesOpenTS:   season.EntriesAccepted.From,
			EntriesClosed:   entriesClosed,
			EntriesClosedTS: season.EntriesAccepted.Until,
			SeasonName:      season.Name,
			PayPalClientID:  c.config.PayPalClientID,
			EntryFee:        domain.RealmFromContext(ctx).EntryFee,
		}

		p := newPage(r, c, "Join", "join", "Join", data)

		if err := c.templates.ExecuteTemplate(w, "join", p); err != nil {
			internalError(err).writeTo(w)
			return
		}
	}
}

func frontendPredictionHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.PredictionPageData) {
			p := newPage(r, c, "Update My Prediction", "prediction", "Update My Prediction", data)

			if err := c.templates.ExecuteTemplate(w, "prediction", p); err != nil {
				internalError(err).writeTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(view.PredictionPageData{Err: err})
			return
		}
		defer cancel()

		data := getPredictionPageData(
			ctx,
			getAuthCookieValue(r),
			c.entryAgent,
			c.tokenAgent,
			c.seasons,
			c.teams,
			c.clock,
		)

		writeResponse(data)
	}
}

func frontendGenerateMagicLoginHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.GenerateMagicLoginPageData) {
			p := newPage(r, c, "Send me a magic login", "", "Login", data)

			// TODO - feat: update page template
			if err := c.templates.ExecuteTemplate(w, "magic-login-generate", p); err != nil {
				internalError(err).writeTo(w)
			}
		}

		// parse request body (standard form)
		if err := r.ParseForm(); err != nil {
			writeResponse(view.GenerateMagicLoginPageData{Err: err})
			return
		}
		var input generateMagicLoginRequest
		for k, v := range r.Form {
			if k == "email_nickname" && len(v) > 0 {
				input.EmailNickname = v[0]
			}
		}

		// check that input is valid
		if input.EmailNickname == "" {
			writeResponse(view.GenerateMagicLoginPageData{Err: errors.New("invalid request")})
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(view.GenerateMagicLoginPageData{Err: err})
			return
		}
		defer cancel()

		// get realm from context
		realm := domain.RealmFromContext(ctx)

		// retrieve entry
		entry, err := retrieveEntryByEmailOrNickname(ctx, input.EmailNickname, realm.SeasonID, realm.Name, c.entryAgent)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				// we can't find an existing entry, but we don't want to let the user know
				// just pretend everything is ok...
				writeResponse(view.GenerateMagicLoginPageData{EmailNickname: input.EmailNickname})
				return
			}
			writeResponse(view.GenerateMagicLoginPageData{Err: err})
			return
		}

		// does realm name match our entry?
		if domain.RealmFromContext(ctx).Name != entry.RealmName {
			writeResponse(view.GenerateMagicLoginPageData{Err: errors.New("invalid realm")})
			return
		}

		// generate magic login token
		token, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeMagicLogin, entry.ID.String())
		if err != nil {
			writeResponse(view.GenerateMagicLoginPageData{Err: err})
			return
		}

		// issue email with magic login link
		if err := c.commsAgent.IssueMagicLoginEmail(nil, entry, token.ID); err != nil {
			writeResponse(view.GenerateMagicLoginPageData{Err: err})
			return
		}

		// all good!
		writeResponse(view.GenerateMagicLoginPageData{EmailNickname: input.EmailNickname})
	}
}

func frontendShortCodeResetCompleteHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	invalidTokenErr := errors.New("oh no! looks like your token is invalid :'( please try resetting your short code again")

	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.ShortCodeResetCompletePageData) {
			p := newPage(r, c, "Reset my Short Code", "", "Reset my Short Code", data)

			if err := c.templates.ExecuteTemplate(w, "short-code-reset-complete", p); err != nil {
				internalError(err).writeTo(w)
			}
		}

		// parse reset token from route
		var resetToken string
		if err := getRouteParam(r, "reset_token", &resetToken); err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}
		defer cancel()

		// retrieve token
		token, err := c.tokenAgent.RetrieveTokenByID(ctx, resetToken)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				writeResponse(view.ShortCodeResetCompletePageData{Err: invalidTokenErr})
				return
			}
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// has token expired?
		if token.ExpiresAt.Before(c.clock.Now()) {
			writeResponse(view.ShortCodeResetCompletePageData{Err: invalidTokenErr})
			return
		}

		// is token a magic login token?
		if token.Type != domain.TokenTypeMagicLogin {
			writeResponse(view.ShortCodeResetCompletePageData{Err: invalidTokenErr})
			return
		}

		// retrieve entry
		entry, err := c.entryAgent.RetrieveEntryByID(ctx, token.Value)
		if err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// does realm name match our entry?
		if domain.RealmFromContext(ctx).Name != entry.RealmName {
			writeResponse(view.ShortCodeResetCompletePageData{Err: errors.New("invalid realm")})
			return
		}

		// we've made it this far!
		// now generate a new short code
		newShortCode, err := c.entryAgent.GenerateUniqueShortCode(ctx)
		if err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// update our entry's short code
		entry.ShortCode = newShortCode
		entry, err = c.entryAgent.UpdateEntry(ctx, entry)
		if err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// delete short code reset token
		if err := c.tokenAgent.DeleteToken(ctx, *token); err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// issue email confirming short code reset
		if err := c.commsAgent.IssueShortCodeResetCompleteEmail(nil, &entry); err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// all good!
		writeResponse(view.ShortCodeResetCompletePageData{ShortCode: newShortCode})
	}
}
