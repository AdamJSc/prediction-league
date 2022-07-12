package app

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
	"time"
)

var genericErr = errors.New("oh no! something went wrong :'( we've been told about it... please try again")

func frontendIndexHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			internalError(err).writeTo(w)
			return
		}
		defer cancel()

		realm := domain.RealmFromContext(ctx)

		p := newPage(r, c, "Home", "home", realm.Config.HomePageHeading, nil)

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

		realm := domain.RealmFromContext(ctx)
		data := view.FAQPageData{
			FAQs: realm.FAQs,
		}

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

		realm := domain.RealmFromContext(ctx)
		season, err := c.seasons.GetByID(realm.Config.SeasonID)
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
			EntryFee:        realm.EntryFee,
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
			c.logger,
		)

		writeResponse(data)
	}
}

func frontendGenerateMagicLoginHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.GenerateMagicLoginPageData) {
			p := newPage(r, c, "Login", "", "Login", data)

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
			if k == "email_addr" && len(v) > 0 {
				input.EmailAddr = v[0]
			}
		}

		// check that input is valid
		if input.EmailAddr == "" {
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
		entry, err := retrieveEntryByEmailAddr(ctx, input.EmailAddr, realm.Config.SeasonID, realm.Config.Name, c.entryAgent)
		if err != nil {
			switch {
			case errors.As(err, &domain.NotFoundError{}):
				// we can't find an existing entry, but we don't want to let the user know
				// just pretend everything is ok...
				writeResponse(view.GenerateMagicLoginPageData{EmailAddr: input.EmailAddr})
				return
			}
			c.logger.Errorf("cannot retrieve entry for '%s': %s", input.EmailAddr, err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// does realm name match our entry?
		if realm.Config.Name != entry.RealmName {
			c.logger.Errorf("context realm name '%s' does not match entry realm name '%s'", realm.Config.Name, entry.RealmName)
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// purge unused prediction tokens for this user
		if _, err := c.tokenAgent.DeleteInFlightTokens(ctx, domain.TokenTypeMagicLogin, entry.ID.String()); err != nil {
			c.logger.Errorf("cannot purge in-flight magic login tokens: %s", err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// generate new magic login token
		mTkn, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeMagicLogin, entry.ID.String())
		if err != nil {
			c.logger.Errorf("cannot generate magic login token for entry id '%s': %s", entry.ID.String(), err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// issue email with magic login link
		if err := c.commsAgent.IssueMagicLoginEmail(nil, entry, mTkn.ID); err != nil {
			c.logger.Errorf("cannot issue magic login email for entry id '%s': %s", entry.ID.String(), err.Error())
			writeResponse(view.GenerateMagicLoginPageData{Err: genericErr})
			return
		}

		// all good!
		writeResponse(view.GenerateMagicLoginPageData{EmailAddr: input.EmailAddr})
	}
}

func frontendRedeemMagicLoginHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeRedirect = func(loc string) {
			w.Header().Set("Location", loc)
			w.WriteHeader(http.StatusFound)
		}

		redirFail := "/failed"

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			c.logger.Errorf("cannot get context from request: %s", err.Error())
			writeRedirect(redirFail)
			return
		}
		defer cancel()

		realm := domain.RealmFromContext(ctx)
		redirOk := domain.GetPredictionURL(realm)

		// parse magic token from route
		var mTknID string
		if err := getRouteParam(r, "magic_token_id", &mTknID); err != nil {
			c.logger.Errorf("cannot parse route param 'magic_token_id': %s", err.Error())
			writeRedirect(redirFail)
			return
		}

		// retrieve token
		mTkn, err := c.tokenAgent.RetrieveTokenByID(ctx, mTknID)
		if err != nil {
			switch {
			case errors.As(err, &domain.NotFoundError{}):
				c.logger.Errorf("magic token '%s' not found", mTknID)
				writeRedirect(redirFail)
				return
			}
			c.logger.Errorf("cannot retrieve magic token '%s': %s", mTknID, err.Error())
			writeRedirect(redirFail)
			return
		}

		// is token a magic login token?
		if mTkn.Type != domain.TokenTypeMagicLogin {
			c.logger.Errorf("token id '%s' is not a magic token", mTkn.ID)
			writeRedirect(redirFail)
			return
		}

		// has token expired?
		if mTkn.ExpiresAt.Before(c.clock.Now()) {
			c.logger.Errorf("magic token id '%s' has expired", mTkn.ID)
			writeRedirect(redirFail)
			return
		}

		// retrieve entry
		entry, err := c.entryAgent.RetrieveEntryByID(ctx, mTkn.Value)
		if err != nil {
			c.logger.Errorf("cannot retrieve entry with magic token id '%s': value '%s': %s", mTkn.ID, mTkn.Value, err.Error())
			writeRedirect(redirFail)
			return
		}

		// does realm name match our entry?
		if realm.Config.Name != entry.RealmName {
			c.logger.Errorf("context realm name '%s' does not match entry realm name '%s'", realm.Config.Name, entry.RealmName)
			writeRedirect(redirFail)
			return
		}

		// we've made it this far!
		// generate a new auth token for our entry, and set it as a cookie
		authTkn, err := c.tokenAgent.GenerateToken(ctx, domain.TokenTypeAuth, entry.ID.String())
		if err != nil {
			responseFromError(err).writeTo(w)
			return
		}
		setAuthCookieValue(authTkn.ID, w, r)

		// redeem magic token
		if err := c.tokenAgent.RedeemToken(ctx, *mTkn); err != nil {
			// log error and continue
			c.logger.Errorf("cannot redeem magic token id '%s': %s", mTkn.ID, err.Error())
		}

		// all ok!
		writeRedirect(redirOk)
	}
}

func frontendMagicLoginFailedHandler(c *container) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}
		p := newPage(r, c, "Login", "", "Login", nil)
		if err := c.templates.ExecuteTemplate(buf, "magic-login-failed", p); err != nil {
			internalError(fmt.Errorf("cannot execute template: %w", err)).writeTo(w)
			return
		}
		w.Write(buf.Bytes())
	}
}
