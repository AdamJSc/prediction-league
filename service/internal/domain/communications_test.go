package domain_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"prediction-league/service/internal/domain"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert/cmp"
)

func TestNewCommunicationsAgent(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		emlQ := &domain.InMemEmailQueue{}

		tt := []struct {
			er      domain.EntryRepository
			epr     domain.EntryPredictionRepository
			sr      domain.StandingsRepository
			emlQ    domain.EmailQueue
			tpl     *domain.Templates
			sc      domain.SeasonCollection
			tc      domain.TeamCollection
			rc      domain.RealmCollection
			wantErr error
		}{
			{nil, epr, sr, emlQ, tpl, sc, tc, rc, domain.ErrIsNil},
			{er, nil, sr, emlQ, tpl, sc, tc, rc, domain.ErrIsNil},
			{er, epr, nil, emlQ, tpl, sc, tc, rc, domain.ErrIsNil},
			{er, epr, sr, nil, tpl, sc, tc, rc, domain.ErrIsNil},
			{er, epr, sr, emlQ, nil, sc, tc, rc, domain.ErrIsNil},
			{er, epr, sr, emlQ, tpl, nil, tc, rc, domain.ErrIsNil},
			{er, epr, sr, emlQ, tpl, sc, nil, rc, domain.ErrIsNil},
			{er, epr, sr, emlQ, tpl, sc, tc, nil, domain.ErrIsNil},
			{er, epr, sr, emlQ, tpl, sc, tc, rc, nil},
		}

		for idx, tc := range tt {
			agent, gotErr := domain.NewCommunicationsAgent(tc.er, tc.epr, tc.sr, tc.emlQ, tc.tpl, tc.sc, tc.tc, tc.rc)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && agent == nil {
				t.Fatalf("tc #%d: want non-empty agent, got nil", idx)
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

		expectedSubject := fmt.Sprintf(domain.EmailSubjectRoundComplete, standings.RoundNumber+1)

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_round_complete", domain.RoundCompleteEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			RoundNumber:    standings.RoundNumber,
			LeaderBoardURL: fmt.Sprintf("%s/leaderboard", rlm.Origin),
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
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

		expectedSubject := domain.EmailSubjectFinalRoundComplete

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_final_round_complete", domain.RoundCompleteEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			RoundNumber:    standings.RoundNumber,
			LeaderBoardURL: fmt.Sprintf("%s/leaderboard", rlm.Origin),
			PredictionsURL: fmt.Sprintf("%s/prediction", rlm.Origin),
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
}

func TestCommunicationsAgent_IssueMagicLoginEmail(t *testing.T) {
	defer truncate(t)

	t.Run("issue magic login email with a valid entry must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		tokenId := "MAGIC12345"

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		if err := agent.IssueMagicLoginEmail(ctx, &entry, tokenId); err != nil {
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

		expectedSubject := domain.EmailSubjectMagicLogin

		expectedPlainText := mustExecuteTemplate(t, tpl, "email_txt_magic_login", domain.MagicLoginEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      rlm.Contact.Name,
				URL:          rlm.Origin,
				SupportEmail: rlm.Contact.EmailProper,
			},
			LoginURL: fmt.Sprintf("%s/login/%s", rlm.Origin, tokenId),
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

	t.Run("issue magic login email with no entry must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		emlQ := domain.NewInMemEmailQueue()

		agent, err := domain.NewCommunicationsAgent(er, epr, sr, emlQ, tpl, sc, tc, rc)
		if err != nil {
			t.Fatal(err)
		}

		err = agent.IssueMagicLoginEmail(ctx, nil, "dat_string")
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

		err = agent.IssueMagicLoginEmail(ctx, &entry, "dat_string")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue magic login email with an entry whose season does not exist must fail", func(t *testing.T) {
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

		err = agent.IssueMagicLoginEmail(ctx, &entry, "dat_string")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
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

func TestNewNoopEmailClient(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		l := &mockLogger{}

		tt := []struct {
			l       domain.Logger
			wantErr error
		}{
			{nil, domain.ErrIsNil},
			{l, nil},
		}
		for idx, tc := range tt {
			emlCl, gotErr := domain.NewNoopEmailClient(tc.l)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && emlCl == nil {
				t.Fatalf("tc #%d: want non-empty email client, got nil", idx)
			}
		}
	})
}

func TestNoopEmailClient_SendEmail(t *testing.T) {
	t.Run("happy path must log the expected output", func(t *testing.T) {
		l := newMockLogger()

		emCl, err := domain.NewNoopEmailClient(l)
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

		wantOutput := "sent email: {" +
			"From:{Name:Paul Mc Address:Bass Town} " +
			"To:{Name:John L Address:Sunglassesville} " +
			"ReplyTo:{Name: Address:} " +
			"SenderDomain:bands.liverpool.net " +
			"Subject:Cavern Bar " +
			"PlainText:We're out of lime cordial, can you pick some up?" +
			"}"
		gotOutput := l.buf.String()

		diff := gocmp.Diff(wantOutput, gotOutput)
		if diff != "" {
			t.Fatalf("want logged output %s, got %s, diff %s", wantOutput, gotOutput, diff)
		}
	})
}
