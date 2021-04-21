package handlers

import (
	"errors"
	"fmt"
	"github.com/LUSHDigital/core/rest"
	"net/http"
	"prediction-league/service/internal/app/httph"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/view"
	"time"
)

func frontendIndexHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
		defer cancel()

		seasonID := domain.RealmFromContext(ctx).SeasonID
		season, err := domain.SeasonsDataStore.GetByID(seasonID)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}

		bannerTitle := fmt.Sprintf("%s<br />Prediction League", season.Name)

		p := newPage(r, c, "Home", "home", bannerTitle, nil)

		if err := c.Template().ExecuteTemplate(w, "index", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendLeaderBoardHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.LeaderBoardPageData) {
			p := newPage(r, c, "Leaderboard", "leaderboard", "Leaderboard", data)

			if err := c.Template().ExecuteTemplate(w, "leaderboard", p); err != nil {
				rest.InternalError(err).WriteTo(w)
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
			&domain.EntryAgent{EntryAgentInjector: c},
			&domain.StandingsAgent{StandingsAgentInjector: c},
			&domain.LeaderBoardAgent{LeaderBoardAgentInjector: c},
		)

		writeResponse(data)
	}
}

func frontendFAQHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.FAQPageData) {
			p := newPage(r, c, "FAQ", "faq", "FAQ", data)

			if err := c.Template().ExecuteTemplate(w, "faq", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
		defer cancel()

		data := getFAQPageData(domain.RealmFromContext(ctx).Name)

		writeResponse(data)
	}
}

func frontendJoinHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
		defer cancel()

		season, err := domain.SeasonsDataStore.GetByID(domain.RealmFromContext(ctx).SeasonID)
		if err != nil {
			rest.InternalError(err).WriteTo(w)
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
			PayPalClientID:  c.Config().PayPalClientID,
			EntryFee:        domain.RealmFromContext(ctx).EntryFee,
		}

		p := newPage(r, c, "Join", "join", "Join", data)

		if err := c.Template().ExecuteTemplate(w, "join", p); err != nil {
			rest.InternalError(err).WriteTo(w)
			return
		}
	}
}

func frontendPredictionHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.PredictionPageData) {
			p := newPage(r, c, "Update My Prediction", "prediction", "Update My Prediction", data)

			if err := c.Template().ExecuteTemplate(w, "prediction", p); err != nil {
				rest.InternalError(err).WriteTo(w)
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
			&domain.EntryAgent{EntryAgentInjector: c},
			&domain.TokenAgent{TokenAgentInjector: c},
		)

		writeResponse(data)
	}
}

func frontendShortCodeResetBeginHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	entryAgent := &domain.EntryAgent{EntryAgentInjector: c}
	tokenAgent := &domain.TokenAgent{TokenAgentInjector: c}
	commsAgent := &domain.CommunicationsAgent{CommunicationsAgentInjector: c}

	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.ShortCodeResetBeginPageData) {
			p := newPage(r, c, "Reset my Short Code", "", "Reset my Short Code", data)

			if err := c.Template().ExecuteTemplate(w, "short-code-reset-begin", p); err != nil {
				rest.InternalError(err).WriteTo(w)
			}
		}

		// parse request body (standard form)
		if err := r.ParseForm(); err != nil {
			writeResponse(view.ShortCodeResetBeginPageData{Err: err})
			return
		}
		var input shortCodeResetRequest
		for k, v := range r.Form {
			if k == "email_nickname" && len(v) > 0 {
				input.EmailNickname = v[0]
			}
		}

		// check that input is valid
		if input.EmailNickname == "" {
			writeResponse(view.ShortCodeResetBeginPageData{Err: errors.New("invalid request")})
			return
		}

		// get context from request
		ctx, cancel, err := contextFromRequest(r, c)
		if err != nil {
			writeResponse(view.ShortCodeResetBeginPageData{Err: err})
			return
		}
		defer cancel()

		// get realm from context
		realm := domain.RealmFromContext(ctx)

		// retrieve entry
		entry, err := retrieveEntryByEmailOrNickname(ctx, input.EmailNickname, realm.SeasonID, realm.Name, entryAgent)
		if err != nil {
			switch err.(type) {
			case domain.NotFoundError:
				// we can't find an existing entry, but we don't want to let the user know
				// just pretend everything is ok...
				writeResponse(view.ShortCodeResetBeginPageData{EmailNickname: input.EmailNickname})
				return
			}
			writeResponse(view.ShortCodeResetBeginPageData{Err: err})
			return
		}

		// does realm name match our entry?
		if domain.RealmFromContext(ctx).Name != entry.RealmName {
			writeResponse(view.ShortCodeResetBeginPageData{Err: errors.New("invalid realm")})
			return
		}

		// generate short code reset token
		token, err := tokenAgent.GenerateToken(ctx, domain.TokenTypeShortCodeResetToken, entry.ID.String())
		if err != nil {
			writeResponse(view.ShortCodeResetBeginPageData{Err: err})
			return
		}

		// issue email with short code reset link
		if err := commsAgent.IssueShortCodeResetBeginEmail(nil, entry, token.ID); err != nil {
			writeResponse(view.ShortCodeResetBeginPageData{Err: err})
			return
		}

		// all good!
		writeResponse(view.ShortCodeResetBeginPageData{EmailNickname: input.EmailNickname})
	}
}

func frontendShortCodeResetCompleteHandler(c *httph.HTTPAppContainer) func(w http.ResponseWriter, r *http.Request) {
	entryAgent := &domain.EntryAgent{EntryAgentInjector: c}
	tokenAgent := &domain.TokenAgent{TokenAgentInjector: c}
	commsAgent := &domain.CommunicationsAgent{CommunicationsAgentInjector: c}

	invalidTokenErr := errors.New("oh no! looks like your token is invalid :'( please try resetting your short code again")

	return func(w http.ResponseWriter, r *http.Request) {
		var writeResponse = func(data view.ShortCodeResetCompletePageData) {
			p := newPage(r, c, "Reset my Short Code", "", "Reset my Short Code", data)

			if err := c.Template().ExecuteTemplate(w, "short-code-reset-complete", p); err != nil {
				rest.InternalError(err).WriteTo(w)
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
		token, err := tokenAgent.RetrieveTokenByID(ctx, resetToken)
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
		if token.ExpiresAt.Before(domain.TimestampFromContext(ctx)) {
			writeResponse(view.ShortCodeResetCompletePageData{Err: invalidTokenErr})
			return
		}

		// is token a short code refresh token?
		if token.Type != domain.TokenTypeShortCodeResetToken {
			writeResponse(view.ShortCodeResetCompletePageData{Err: invalidTokenErr})
			return
		}

		// retrieve entry
		entry, err := entryAgent.RetrieveEntryByID(ctx, token.Value)
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
		newShortCode, err := entryAgent.GenerateUniqueShortCode(ctx)
		if err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// update our entry's short code
		entry.ShortCode = newShortCode
		entry, err = entryAgent.UpdateEntry(ctx, entry)
		if err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// delete short code reset token
		if err := tokenAgent.DeleteToken(ctx, *token); err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// issue email confirming short code reset
		if err := commsAgent.IssueShortCodeResetCompleteEmail(nil, &entry); err != nil {
			writeResponse(view.ShortCodeResetCompletePageData{Err: err})
			return
		}

		// all good!
		writeResponse(view.ShortCodeResetCompletePageData{ShortCode: newShortCode})
	}
}
