package domain_test

import (
	"bytes"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/uuid"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/views"
	"testing"
	"time"
)

type testCommsAgentInjector struct {
	config    domain.Config
	db        coresql.Agent
	queue     chan domain.Email
	templates *views.Templates
}

func (t testCommsAgentInjector) Config() domain.Config         { return t.config }
func (t testCommsAgentInjector) MySQL() coresql.Agent          { return t.db }
func (t testCommsAgentInjector) EmailQueue() chan domain.Email { return t.queue }
func (t testCommsAgentInjector) Template() *views.Templates    { return t.templates }

func TestCommunicationsAgent_IssueNewEntryEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	testPaymentDetails := domain.PaymentDetails{
		Amount:       "£12.34",
		Reference:    "PAYMENT_REFERENCE",
		MerchantName: "MERCHANT_NAME",
	}

	injector := testCommsAgentInjector{
		config:    testConfig,
		queue:     make(chan domain.Email, 1),
		templates: templates,
	}

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
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

		if err := agent.IssueNewEntryEmail(ctx, &entry, &testPaymentDetails); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectNewEntry

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_new_entry", domain.NewEntryEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			PaymentDetails: testPaymentDetails,
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
			ShortCode:      entry.ShortCode,
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		err := agent.IssueNewEntryEmail(ctx, nil, &testPaymentDetails)
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

		err := agent.IssueNewEntryEmail(ctx, &entry, nil)
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

		missingAmount := testPaymentDetails
		missingAmount.Amount = ""
		err := agent.IssueNewEntryEmail(ctx, &entry, &missingAmount)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingReference := testPaymentDetails
		missingReference.Reference = ""
		err = agent.IssueNewEntryEmail(ctx, &entry, &missingReference)
		if !cmp.ErrorType(err, domain.ValidationError{})().Success() {
			expectedTypeOfGot(t, domain.ValidationError{}, err)
		}

		missingMerchantName := testPaymentDetails
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

		err := agent.IssueNewEntryEmail(ctx, &entry, &testPaymentDetails)
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

		err := agent.IssueNewEntryEmail(ctx, &entry, &testPaymentDetails)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssueRoundCompleteEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	entry := insertEntry(t, generateTestEntry(t,
		"Harry Redknapp",
		"MrHarryR",
		"harry.redknapp@football.net",
	))
	entryPrediction := insertEntryPrediction(t, generateTestEntryPrediction(t, entry.ID))
	standings := insertStandings(t, generateTestStandings(t))
	scoredEntryPrediction := insertScoredEntryPrediction(t, generateTestScoredEntryPrediction(t, entryPrediction.ID, standings.ID))

	injector := testCommsAgentInjector{
		config:    testConfig,
		db:        db,
		queue:     make(chan domain.Email, 1),
		templates: templates,
	}

	defer close(injector.queue)

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	t.Run("issue round complete email with a valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedRankingStrings, err := domain.TeamRankingsAsStrings(scoredEntryPrediction.Rankings, standings.Rankings)
		if err != nil {
			t.Fatal(err)
		}

		expectedSubject := fmt.Sprintf(domain.EmailSubjectRoundComplete, standings.RoundNumber)

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_round_complete", domain.RoundCompleteEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			RoundNumber:       standings.RoundNumber,
			RankingsAsStrings: expectedRankingStrings,
			LeaderBoardURL:    fmt.Sprintf("%s/leaderboard", testRealm.Origin),
		})

		if err := agent.IssueRoundCompleteEmail(ctx, &scoredEntryPrediction, false); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		expectedRankingStrings, err := domain.TeamRankingsAsStrings(scoredEntryPrediction.Rankings, standings.Rankings)
		if err != nil {
			t.Fatal(err)
		}

		expectedSubject := fmt.Sprintf(domain.EmailSubjectRoundComplete, standings.RoundNumber)

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_final_round_complete", domain.RoundCompleteEmailData{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			RoundNumber:       standings.RoundNumber,
			RankingsAsStrings: expectedRankingStrings,
			LeaderBoardURL:    fmt.Sprintf("%s/leaderboard", testRealm.Origin),
		})

		if err := agent.IssueRoundCompleteEmail(ctx, &scoredEntryPrediction, true); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		invalidUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		sep := scoredEntryPrediction
		sep.EntryPredictionID = invalidUUID

		err = agent.IssueRoundCompleteEmail(ctx, &sep, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose standings ID does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		invalidUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		sep := scoredEntryPrediction
		sep.StandingsID = invalidUUID

		err = agent.IssueRoundCompleteEmail(ctx, &sep, false)
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

		err := agent.IssueRoundCompleteEmail(ctx, &invalidScoredEntryPrediction, false)
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

		err := agent.IssueRoundCompleteEmail(ctx, &invalidScoredEntryPrediction, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose rankings are empty must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		sep := scoredEntryPrediction
		sep.Rankings = nil

		err := agent.IssueRoundCompleteEmail(ctx, &sep, false)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssueShortCodeResetBeginEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	injector := testCommsAgentInjector{
		config:    testConfig,
		queue:     make(chan domain.Email, 1),
		templates: templates,
	}

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

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

		if err := agent.IssueShortCodeResetBeginEmail(ctx, &entry, resetToken); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectShortCodeResetBegin

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_short_code_reset_begin", domain.ShortCodeResetBeginEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			ResetURL: fmt.Sprintf("%s/reset/%s", testRealm.Origin, resetToken),
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		err := agent.IssueShortCodeResetBeginEmail(ctx, nil, "dat_string")
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

		err := agent.IssueShortCodeResetBeginEmail(ctx, &entry, "dat_string")
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

		err := agent.IssueShortCodeResetBeginEmail(ctx, &entry, "dat_string")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssueShortCodeResetCompleteEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	injector := testCommsAgentInjector{
		config:    testConfig,
		queue:     make(chan domain.Email, 1),
		templates: templates,
	}

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	t.Run("issue short code reset complete email with a valid entry must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		if err := agent.IssueShortCodeResetCompleteEmail(ctx, &entry); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectShortCodeResetComplete

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_short_code_reset_complete", domain.ShortCodeResetCompleteEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
			ShortCode:      entry.ShortCode,
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		err := agent.IssueShortCodeResetCompleteEmail(ctx, nil)
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

		err := agent.IssueShortCodeResetCompleteEmail(ctx, &entry)
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

		err := agent.IssueShortCodeResetCompleteEmail(ctx, &entry)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestCommunicationsAgent_IssuePredictionWindowOpenEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	injector := testCommsAgentInjector{
		config:    testConfig,
		queue:     make(chan domain.Email, 1),
		templates: templates,
	}

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
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

		if err := agent.IssuePredictionWindowOpenEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectPredictionWindowOpen

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_prediction_window_open", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            1,
				Total:              2,
				IsLast:             false,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		if err := agent.IssuePredictionWindowOpenEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectPredictionWindowOpenFinal

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_prediction_window_open", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            2,
				Total:              2,
				IsLast:             true,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		err := agent.IssuePredictionWindowOpenEmail(ctx, nil, domain.SequencedTimeFrame{})
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

		err := agent.IssuePredictionWindowOpenEmail(ctx, &entry, domain.SequencedTimeFrame{})
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

		err := agent.IssuePredictionWindowOpenEmail(ctx, &entry, domain.SequencedTimeFrame{})
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

		err := agent.IssuePredictionWindowOpenEmail(ctx, &entry, window)
		if err == nil || err.Error() != expectedErrMsg {
			expectedGot(t, expectedErrMsg, err)
		}
	})
}

func TestCommunicationsAgent_IssuePredictionWindowClosingEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	injector := testCommsAgentInjector{
		config:    testConfig,
		queue:     make(chan domain.Email, 1),
		templates: templates,
	}

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
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

		if err := agent.IssuePredictionWindowClosingEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectPredictionWindowClosing

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_prediction_window_closing", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
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
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		if err := agent.IssuePredictionWindowClosingEmail(ctx, &entry, window); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectPredictionWindowClosingFinal

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_prediction_window_closing", domain.PredictionWindowEmail{
			MessagePayload: domain.MessagePayload{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			Window: domain.WindowData{
				Current:            2,
				Total:              2,
				IsLast:             true,
				CurrentClosingDate: "Sat 26 May",
				CurrentClosingTime: "3:00pm",
			},
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
		})

		if email.From.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.From.Name)
		}
		if email.From.Address != testRealm.Contact.EmailDoNotReply {
			expectedGot(t, testRealm.Contact.EmailDoNotReply, email.From.Address)
		}
		if email.To.Name != entry.EntrantName {
			expectedGot(t, entry.EntrantName, email.To.Name)
		}
		if email.To.Address != entry.EntrantEmail {
			expectedGot(t, entry.EntrantEmail, email.To.Address)
		}
		if email.ReplyTo.Name != testRealm.Contact.Name {
			expectedGot(t, testRealm.Contact.Name, email.ReplyTo.Name)
		}
		if email.ReplyTo.Address != testRealm.Contact.EmailProper {
			expectedGot(t, testRealm.Contact.EmailProper, email.ReplyTo.Address)
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

		err := agent.IssuePredictionWindowClosingEmail(ctx, nil, domain.SequencedTimeFrame{})
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

		err := agent.IssuePredictionWindowClosingEmail(ctx, &entry, domain.SequencedTimeFrame{})
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

		err := agent.IssuePredictionWindowClosingEmail(ctx, &entry, domain.SequencedTimeFrame{})
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

		err := agent.IssuePredictionWindowClosingEmail(ctx, &entry, window)
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
func mustExecuteTemplate(t *testing.T, templates *views.Templates, templateName string, templateData interface{}) string {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	if err := templates.ExecuteTemplate(buf, templateName, templateData); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
