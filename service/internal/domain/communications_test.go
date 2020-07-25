package domain_test

import (
	"bytes"
	"fmt"
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/LUSHDigital/uuid"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/emails"
	"prediction-league/service/internal/messages"
	"prediction-league/service/internal/views"
	"testing"
)

type testCommsAgentInjector struct {
	config    domain.Config
	db        coresql.Agent
	queue     chan messages.Email
	templates *views.Templates
}

func (t testCommsAgentInjector) Config() domain.Config           { return t.config }
func (t testCommsAgentInjector) MySQL() coresql.Agent            { return t.db }
func (t testCommsAgentInjector) EmailQueue() chan messages.Email { return t.queue }
func (t testCommsAgentInjector) Template() *views.Templates      { return t.templates }

func TestCommunicationsAgent_IssuesNewEntryEmail(t *testing.T) {
	defer truncate(t)

	testConfig := domain.Config{
		Realms: make(map[string]domain.Realm),
	}
	testRealm := testRealm(t)
	testConfig.Realms[testRealm.Name] = testRealm

	injector := testCommsAgentInjector{
		config:    testConfig,
		queue:     make(chan messages.Email, 10), // only add 1 email at a time to channel in actual tests
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

		if err := agent.IssueNewEntryEmail(ctx, &entry); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedSubject := domain.EmailSubjectNewEntry

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_new_entry", emails.NewEntryEmailData{
			EmailData: emails.EmailData{
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

		err := agent.IssueNewEntryEmail(ctx, &entry)
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

		err := agent.IssueNewEntryEmail(ctx, &entry)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

// TODO - tests for IssueRoundCompleteEmail

func TestCommunicationsAgent_IssuesRoundCompleteEmail(t *testing.T) {
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
		queue:     make(chan messages.Email, 10), // only add 1 email at a time to channel in actual tests
		templates: templates,
	}

	agent := domain.CommunicationsAgent{
		CommunicationsAgentInjector: injector,
	}

	t.Run("issue round complete email with a valid scored entry prediction must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		if err := agent.IssueRoundCompleteEmail(ctx, &scoredEntryPrediction); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		rankingStrings, err := domain.TeamRankingsAsStrings(scoredEntryPrediction.Rankings)
		if err != nil {
			t.Fatal(err)
		}

		expectedSubject := fmt.Sprintf(domain.EmailSubjectRoundComplete, standings.RoundNumber)

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_round_complete", emails.RoundCompleteEmailData{
			EmailData: emails.EmailData{
				Name:         entry.EntrantName,
				SeasonName:   testSeason.Name,
				SignOff:      testRealm.Contact.Name,
				URL:          testRealm.Origin,
				SupportEmail: testRealm.Contact.EmailProper,
			},
			RoundNumber:       standings.RoundNumber,
			RankingsAsStrings: rankingStrings,
			LeaderBoardURL:    fmt.Sprintf("%s/leaderboard", testRealm.Origin),
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

	t.Run("issue round complete email with a scored entry prediction whose entry prediction ID does not exist must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		invalidUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}

		sep := scoredEntryPrediction
		sep.EntryPredictionID = invalidUUID

		err = agent.IssueRoundCompleteEmail(ctx, &sep)
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

		err = agent.IssueRoundCompleteEmail(ctx, &sep)
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

		err := agent.IssueRoundCompleteEmail(ctx, &invalidScoredEntryPrediction)
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

		err := agent.IssueRoundCompleteEmail(ctx, &invalidScoredEntryPrediction)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue round complete email with a scored entry prediction whose rankings are empty must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		sep := scoredEntryPrediction
		sep.Rankings = nil

		err := agent.IssueRoundCompleteEmail(ctx, &sep)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
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
