package domain_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestNewCommunicationsAgent(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		// TODO - tests: add wantErr
		emlQ := domain.NewInMemEmailQueue()

		tt := []struct {
			er   domain.EntryRepository
			epr  domain.EntryPredictionRepository
			sr   domain.StandingsRepository
			emlQ domain.EmailQueue
			tpl  *domain.Templates
			sc   domain.SeasonCollection
			tc   domain.TeamCollection
			rc   domain.RealmCollection
		}{
			{nil, epr, sr, emlQ, tpl, sc, tc, rc},
			{er, nil, sr, emlQ, tpl, sc, tc, rc},
			{er, epr, nil, emlQ, tpl, sc, tc, rc},
			{er, epr, sr, nil, tpl, sc, tc, rc},
			{er, epr, sr, emlQ, nil, sc, tc, rc},
			{er, epr, sr, emlQ, tpl, nil, tc, rc},
			{er, epr, sr, emlQ, tpl, sc, nil, rc},
			{er, epr, sr, emlQ, tpl, sc, tc, nil},
		}

		for _, tc := range tt {
			_, gotErr := domain.NewCommunicationsAgent(tc.er, tc.epr, tc.sr, tc.emlQ, tc.tpl, tc.sc, tc.tc, tc.rc)
			if !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
			}
		}
	})
}

func TestCommunicationsAgent_IssueNewEntryEmail(t *testing.T) {
	defer truncate(t)

	payment := domain.PaymentDetails{
		Amount:       "£12.34",
		Reference:    "PAYMENT_REFERENCE",
		MerchantName: "MERCHANT_NAME",
	}

	t.Run("issue new entry email with a valid entry must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssueNewEntryEmail(ctx, &entry, &payment); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectNewEntry

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_new_entry", domain.NewEntryEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			PaymentDetails: payment,
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
			ShortCode:      entry.ShortCode,
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue new entry email with no entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueNewEntryEmail(ctx, nil, &payment)
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("issue new entry email with no payment details must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueNewEntryEmail(ctx, &entry, nil)
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("issue new entry email with empty payment details must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		missingAmount := payment
		missingAmount.Amount = ""

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueNewEntryEmail(ctx, &entry, &missingAmount)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingReference := payment
		missingReference.Reference = ""
		err = agent.IssueNewEntryEmail(ctx, &entry, &missingReference)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingMerchantName := payment
		missingMerchantName.MerchantName = ""
		err = agent.IssueNewEntryEmail(ctx, &entry, &missingMerchantName)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}
	})

	t.Run("issue new entry email with an entry whose realm does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.RealmName = "not_a_valid_realm"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueNewEntryEmail(ctx, &entry, &payment)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue new entry email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.SeasonID = "not_a_valid_season"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueNewEntryEmail(ctx, &entry, &payment)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssueRoundCompleteEmail(t *testing.T) {
	defer truncate(t)

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))

	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))
	scoredEntryPrediction := insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID))

	t.Run("issue round complete email with a valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedRankingStrings, err := domain.TeamRankingsAsStrings(scoredEntryPrediction.Rankings, standings.Rankings, tc)
		if err != nil {
			t.Fatal(err)
		}

		expectedSubject := fmt.Sprintf(domain.EmailSubjectRoundComplete, standings.RoundNumber)

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_round_complete", domain.RoundCompleteEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			RoundNumber:       standings.RoundNumber,
			RankingsAsStrings: expectedRankingStrings,
			LeaderBoardURL:    fmt.Sprintf("%s/leaderboard", rlm.Origin),
		})

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssueRoundCompleteEmail(ctx, scoredEntryPrediction, false); err != nil {
			t.Fatal(err)
		}
		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue final round complete email with a valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedRankingStrings, err := domain.TeamRankingsAsStrings(scoredEntryPrediction.Rankings, standings.Rankings, tc)
		if err != nil {
			t.Fatal(err)
		}

		expectedSubject := fmt.Sprintf(domain.EmailSubjectRoundComplete, standings.RoundNumber)

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_final_round_complete", domain.RoundCompleteEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			RoundNumber:       standings.RoundNumber,
			RankingsAsStrings: expectedRankingStrings,
			LeaderBoardURL:    fmt.Sprintf("%s/leaderboard", rlm.Origin),
		})

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssueRoundCompleteEmail(ctx, scoredEntryPrediction, true); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose entry prediction ID does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		invalidUUID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		sep := scoredEntryPrediction
		sep.EntryPredictionID = invalidUUID

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueRoundCompleteEmail(ctx, sep, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose standings ID does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		invalidUUID, err := uuid.NewRandom()
		if err != nil {
			t.Fatal(err)
		}

		sep := scoredEntryPrediction
		sep.StandingsID = invalidUUID

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueRoundCompleteEmail(ctx, sep, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose realm does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entryWithInvalidRealm := generateTestEntry(t,
			"Jamie Redknapp",
			"MrJamieR",
			"jamie.redknapp@football.net",
		)
		entryWithInvalidRealm.RealmName = "not_a_valid_realm"
		entryWithInvalidRealm = insertEntry(t, entryWithInvalidRealm)

		invalidEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entryWithInvalidRealm.ID))
		invalidScoredEntryPrediction := insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, invalidEntryPrediction.ID, standings.ID))

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueRoundCompleteEmail(ctx, invalidScoredEntryPrediction, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entryWithInvalidSeason := generateTestEntry(t,
			"Frank Lampard",
			"FrankieLamps",
			"frank.lampard@football.net",
		)
		entryWithInvalidSeason.SeasonID = "__LOL__"
		entryWithInvalidSeason = insertEntry(t, entryWithInvalidSeason)

		invalidEntryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entryWithInvalidSeason.ID))
		invalidScoredEntryPrediction := insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, invalidEntryPrediction.ID, standings.ID))

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueRoundCompleteEmail(ctx, invalidScoredEntryPrediction, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose rankings are empty must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		sep := scoredEntryPrediction
		sep.Rankings = nil

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueRoundCompleteEmail(ctx, sep, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssueShortCodeResetBeginEmail(t *testing.T) {
	defer truncate(t)

	t.Run("issue short code reset begin email with a valid entry must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		resetToken := "RESET12345"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssueShortCodeResetBeginEmail(ctx, &entry, resetToken); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectShortCodeResetBegin

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_short_code_reset_begin", domain.ShortCodeResetBeginEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			ResetURL: fmt.Sprintf("%s/reset/%s", rlm.Origin, resetToken),
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue short code reset begin email with no entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueShortCodeResetBeginEmail(ctx, nil, "dat_string")
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("issue short code reset begin email with an entry whose realm does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.RealmName = "not_a_valid_realm"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueShortCodeResetBeginEmail(ctx, &entry, "dat_string")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue short code reset begin email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.SeasonID = "not_a_valid_season"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueShortCodeResetBeginEmail(ctx, &entry, "dat_string")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssueShortCodeResetCompleteEmail(t *testing.T) {
	defer truncate(t)

	t.Run("issue short code reset complete email with a valid entry must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssueShortCodeResetCompleteEmail(ctx, &entry); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectShortCodeResetComplete

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_short_code_reset_complete", domain.ShortCodeResetCompleteEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
			ShortCode:      entry.ShortCode,
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue short code reset complete email with no entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueShortCodeResetCompleteEmail(ctx, nil)
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("issue short code reset complete email with an entry whose realm does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.RealmName = "not_a_valid_realm"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueShortCodeResetCompleteEmail(ctx, &entry)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue short code reset complete email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.SeasonID = "not_a_valid_season"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueShortCodeResetCompleteEmail(ctx, &entry)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssuePredictionWindowOpenEmail(t *testing.T) {
	defer truncate(t)

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("issue prediction window open email with a valid entry and sequenced timeframe that is not last must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		window := domain.SequencedTimeFrame{
			Count: 1,
			Total: 2,
			Current: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
		}

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err = agent.IssuePredictionWindowOpenEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectPredictionWindowOpen

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_prediction_window_open", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            1,
				Total:              2,
				IsLast:             false,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue prediction window open email with a valid entry and sequenced timeframe that is last must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		window := domain.SequencedTimeFrame{
			Count: 2,
			Total: 2,
			Current: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
		}

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssuePredictionWindowOpenEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectPredictionWindowOpenFinal

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_prediction_window_open", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            2,
				Total:              2,
				IsLast:             true,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue prediction window open email with no entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowOpenEmail(ctx, nil, domain.SequencedTimeFrame{})
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("issue prediction window open email with an entry whose realm does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.RealmName = "not_a_valid_realm"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowOpenEmail(ctx, &entry, domain.SequencedTimeFrame{})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue prediction window open email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.SeasonID = "not_a_valid_season"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowOpenEmail(ctx, &entry, domain.SequencedTimeFrame{})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue prediction window open email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		window := domain.SequencedTimeFrame{
			Count:   1,
			Total:   2,
			Current: nil, // nil should fail test
			Next: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
		}

		expectedErrMsg := domain.ErrCurrentTimeFrameIsMissing.Error()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowOpenEmail(ctx, &entry, window)
		if err == nil || err.Error() != expectedErrMsg {
			expectedGot(t, expectedErrMsg, err)
		}
	})
}

func TestCommunicationsAgent_IssuePredictionWindowClosingEmail(t *testing.T) {
	defer truncate(t)

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("issue prediction window closing email with a valid entry and sequenced timeframe that is not last must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		window := domain.SequencedTimeFrame{
			Count: 1,
			Total: 2,
			Current: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
			Next: &domain.TimeFrame{
				From:  time.Date(2018, 5, 29, 16, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 29, 17, 0, 0, 0, loc),
			},
		}

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err = agent.IssuePredictionWindowClosingEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectPredictionWindowClosing

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_prediction_window_closing", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            1,
				Total:              2,
				IsLast:             false,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
				NextOpeningDate:    "Tue 29 May",
				NextOpeningTime:    "4:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue prediction window closing email with a valid entry and sequenced timeframe that is last must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		window := domain.SequencedTimeFrame{
			Count: 2,
			Total: 2,
			Current: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
		}

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssuePredictionWindowClosingEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		if err := emlQ.Close(); err != nil {
			t.Fatal(err)
		}

		emls := make([]domain.Email, 0)
		for eml := range emlQ.Read() {
			emls = append(emls, eml)
		}

		if len(emls) != 1 {
			t.Fatalf("want 1 email, got %d", len(emls))
		}

		email := emls[0]

		expectedSubject := domain.EmailSubjectPredictionWindowClosingFinal

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_prediction_window_closing", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            2,
				Total:              2,
				IsLast:             true,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
		})

		if email.From.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.From.Name)
		}
		if email.From.Address != rlm.Contact.EmailDoNotReply {
			expectedGot(t, rlm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != rlm.Contact.Name {
			expectedGot(t, rlm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != rlm.Contact.EmailProper {
			expectedGot(t, rlm.Contact.EmailProper, email.ReplyTo.Address)
		}
		if email.Subject != expectedSubject {
			expectedGot(t, expectedSubject, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue prediction window closing email with no entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowClosingEmail(ctx, nil, domain.SequencedTimeFrame{})
		if !cmp.ErrorType(err, domain.InternalError{})().Success() {
			expectedTypeOfGot(t, domain.InternalError{}, err)
		}
	})

	t.Run("issue prediction window closing email with an entry whose realm does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.RealmName = "not_a_valid_realm"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowClosingEmail(ctx, &entry, domain.SequencedTimeFrame{})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue prediction window closing email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.SeasonID = "not_a_valid_season"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowClosingEmail(ctx, &entry, domain.SequencedTimeFrame{})
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue prediction window closing email with an entry whose season does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		window := domain.SequencedTimeFrame{
			Count:   1,
			Total:   2,
			Current: nil, // nil should fail test
			Next: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
		}

		expectedErrMsg := domain.ErrCurrentTimeFrameIsMissing.Error()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssuePredictionWindowClosingEmail(ctx, &entry, window)
		if err == nil || err.Error() != expectedErrMsg {
			expectedGot(t, expectedErrMsg, err)
		}
	})
}

func TestGenerateWindowDataFromSequencedTimeFrame(t *testing.T) {
	t.Run("generating window data from sequenced timeframe without a current timeframe must return error", func(t *testing.T) {
		sequenced := domain.SequencedTimeFrame{} // Current is nil

		if _, err := domain.GenerateWindowDataFromSequencedTimeFrame(sequenced); err != domain.ErrCurrentTimeFrameIsMissing {
			expectedGot(t, domain.ErrCurrentTimeFrameIsMissing, err)
		}
	})

	t.Run("generating window data from sequenced timeframe without next timeframe must succeed", func(t *testing.T) {
		loc, err := time.LoadLocation("Europe/London")
		if err != nil {
			t.Fatal(err)
		}

		sequenced := domain.SequencedTimeFrame{
			Current: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
			Count: 123,
			Total: 456,
		}

		expected := domain.WindowData{
			Current:            123,
			Total:              456,
			IsLast:             false,
			CurrentClosingDate: "Sat 26 May",
			CurrentClosingTime: "3:00pm",
		}

		actual, err := domain.GenerateWindowDataFromSequencedTimeFrame(sequenced)
		if err != nil {
			t.Fatal(err)
		}

		diff := gocmp.Diff(expected, *actual)
		if diff != "" {
			expectedGot(t, "empty diff", diff)
		}
	})

	t.Run("generating window data from sequenced timeframe with next timeframe must succeed", func(t *testing.T) {
		loc, err := time.LoadLocation("Europe/London")
		if err != nil {
			t.Fatal(err)
		}

		sequenced := domain.SequencedTimeFrame{
			Current: &domain.TimeFrame{
				From:  time.Date(2018, 5, 26, 14, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 26, 15, 0, 0, 0, loc),
			},
			Next: &domain.TimeFrame{
				From:  time.Date(2018, 5, 29, 16, 0, 0, 0, loc),
				Until: time.Date(2018, 5, 29, 17, 0, 0, 0, loc),
			},
			Count: 456,
			Total: 456,
		}

		expected := domain.WindowData{
			Current:            456,
			Total:              456,
			IsLast:             true,
			CurrentClosingDate: "Sat 26 May",
			CurrentClosingTime: "3:00pm",
			NextOpeningDate:    "Tue 29 May",
			NextOpeningTime:    "4:00pm",
		}

		actual, err := domain.GenerateWindowDataFromSequencedTimeFrame(sequenced)
		if err != nil {
			t.Fatal(err)
		}

		diff := gocmp.Diff(expected, *actual)
		if diff != "" {
			expectedGot(t, "empty diff", diff)
		}
	})
}

// mustExecuteTemplate executes the provided template name with the provided template data or produces a test failure on error
func mustExecuteTemplate(t *testing.T, templates *domain.Templates, templateName string, templateData interface{}) string {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, templateName, templateData); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

func TestNewLoggerEmailClient(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		// TODO - tests: replace with tt and wantErr
		l := &logger.Logger{}

		if _, gotErr := domain.NewLoggerEmailClient(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}

		emlCl, err := domain.NewLoggerEmailClient(l)
		if err != nil {
			t.Fatal(err)
		}
		if emlCl == nil {
			t.Fatal("want non-empty logger email client, got nil")
		}
	})
}

func TestLoggerEmailClient_SendEmail(t *testing.T) {
	t.Run("happy path must log the expected output", func(t *testing.T) {
		loc, err := time.LoadLocation("Europe/London")
		if err != nil {
			t.Fatal(err)
		}

		ts := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)
		cl := &mockClock{t: ts}
		buf := &bytes.Buffer{}

		l, err := logger.NewLogger(buf, cl)
		if err != nil {
			t.Fatal(err)
		}

		emCl, err := domain.NewLoggerEmailClient(l)
		if err != nil {
			t.Fatal(err)
		}

		em := domain.Email{
			From: domain.Identity{
				Name:    "Paul Mc",
				Address: "Bass Town",
			},
			To: domain.Identity{
				Name:    "John L",
				Address: "Sunglassesville",
			},
			SenderDomain: "bands.liverpool.net",
			Subject:      "Cavern Bar",
			PlainText:    "We're out of lime cordial, can you pick some up?",
		}
		if err := emCl.SendEmail(context.Background(), em); err != nil {
			t.Fatal(err)
		}

		wantOutput := "2018-05-26T14:00:00+01:00 INFO: [domain/communications.go:553] sent email: {" +
			"From:{Name:Paul Mc Address:Bass Town} " +
			"To:{Name:John L Address:Sunglassesville} " +
			"ReplyTo:{Name: Address:} " +
			"SenderDomain:bands.liverpool.net " +
			"Subject:Cavern Bar " +
			"PlainText:We're out of lime cordial, can you pick some up?" +
			"}\n"
		gotOutput := buf.String()

		diff := gocmp.Diff(wantOutput, gotOutput)
		if diff != "" {
			t.Fatalf("want logged output %s, got %s, diff %s", wantOutput, gotOutput, diff)
		}
	})
}
