package app

import (
	"errors"
	"fmt"
	"net/http"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
	"time"
)

var genericErr = errors.New("oh no! something went wrong :'( we've been told about it, so please try again")

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

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			c.logger.Errorf("cannot get context from request: %s", err.Error())
			writeResponse(view.LeaderBoardPageData{Err: genericErr})
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

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			c.logger.Errorf("cannot get context from request: %s", err.Error())
			writeResponse(view.PredictionPageData{Err: genericErr})
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
			p := newPage(r, c, "Login", "", "Login", data)

			// TODO - feat: update page template
			if err := c.templates.ExecuteTemplate(w, "magic-login-generate", p); err != nil {
				internalError(fmt.Errorf("cannot execute template: %w", err)).writeTo(w)
			}
		}

		// parse request body (standard form)
		if err := r.ParseForm(); err != nil {
			c.logger.Errorf("cannot parse form: %s", err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
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
			c.logger.Errorf("cannot get context from request: %s", err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}
		defer cancel()

		// get realm from context
		realm := domain.RealmFromContext(ctx)

		// retrieve entry
		entry, err := retrieveEntryByEmailOrNickname(ctx, input.EmailNickname, realm.SeasonID, realm.Name, c.entryAgent)
		if err != nil {
			switch {
			case errors.As(err, &domain.NotFoundError{}):
				// we can't find an existing entry, but we don't want to let the user know
				// just pretend everything is ok...
				writeResponse(view.GenerateMagicLoginPageData{EmailNickname: input.EmailNickname})
				return
			}
			c.logger.Errorf("cannot retrieve entry '%s': %s", input.EmailNickname, err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// does realm name match our entry?
		if ctxRealmName := domain.RealmFromContext(ctx).Name; ctxRealmName != entry.RealmName {
			c.logger.Errorf("context realm name '%s' does not match entry realm name '%s'", ctxRealmName, entry.RealmName)
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// TODO - feat: delete any other magic login tokens that are in-flight for this user

		// generate magic login token
		token, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeMagicLogin, entry.ID.String())
		if err != nil {
			c.logger.Errorf("cannot generate magic login token for entry id '%s': %s", entry.ID.String(), err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// issue email with magic login link
		if err := c.commsAgent.IssueMagicLoginEmail(nil, entry, token.ID); err != nil {
			c.logger.Errorf("cannot issue magic login email for entry id '%s': %s", entry.ID.String(), err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// all good!
		writeResponse(view.GenerateMagicLoginPageData{EmailNickname: input.EmailNickname})
	}
}

func frontendRedeemMagicLoginHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO - feat: redirect to login failed page with location header
		var writeResponse = func(data view.RedeemMagicLoginPageData) {
			p := newPage(r, c, "Login failed", "", "Login failed", data)

			if err := c.templates.ExecuteTemplate(w, "magic-login-failed", p); err != nil {
				internalError(fmt.Errorf("cannot execute template: %w", err)).writeTo(w)
			}
		}

		// parse magic token from route
		var mTknID string
		if err := getRouteParam(r, "magic_token_id", &mTknID); err != nil {
			c.logger.Errorf("cannot parse route param 'magic_token_id': %s", err.Error())
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			c.logger.Errorf("cannot get context from request: %s", err.Error())
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}
		defer cancel()

		// retrieve token
		mTkn, err := c.tokenAgent.RetrieveTokenByID(ctx, mTknID)
		if err != nil {
			switch {
			case errors.As(err, &domain.NotFoundError{}):
				c.logger.Errorf("magic token '%s' not found", mTknID)
				writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
				return
			}
			c.logger.Errorf("cannot retrieve magic token '%s': %s", mTknID, err.Error())
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}

		// is token a magic login token?
		if mTkn.Type != domain.TokenTypeMagicLogin {
			c.logger.Errorf("token id '%s' is not a magic token", mTkn.ID)
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}

		// has token expired?
		if mTkn.ExpiresAt.Before(c.clock.Now()) {
			c.logger.Errorf("magic token id '%s' has expired", mTkn.ID)
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}

		// retrieve entry
		entry, err := c.entryAgent.RetrieveEntryByID(ctx, mTkn.Value)
		if err != nil {
			c.logger.Errorf("cannot retrieve entry with magic token id '%s': value '%s': %s", mTkn.ID, mTkn.Value, err.Error())
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}

		// does realm name match our entry?
		if ctxRealmName := domain.RealmFromContext(ctx).Name; ctxRealmName != entry.RealmName {
			c.logger.Errorf("context realm name '%s' does not match entry realm name '%s'", ctxRealmName, entry.RealmName)
			writeResponse(view.RedeemMagicLoginPageData{Err: genericErr})
			return
		}

		// we've made it this far!
		// generate a new auth token for our entry, and set it as a cookie
		aTkn, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeAuth, entry.ID.String())
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		setAuthCookieValue(aTkn.ID, w, r)

		// delete magic token
		// TODO - feat: replace with token redeem
		if err := c.tokenAgent.DeleteToken(ctx, *mTkn); err != nil {
			// log error and continue
			c.logger.Errorf("cannot delete magic token id '%s': %s", mTkn.ID, err.Error())
		}

		// add redirect header
		w.Header().Set("Location", "/prediction")
		w.WriteHeader(http.StatusFound)
	}
}
