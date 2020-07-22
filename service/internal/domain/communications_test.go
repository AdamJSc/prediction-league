package domain_test

import (
	"bytes"
	"fmt"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/emails"
	"prediction-league/service/internal/messages"
	"prediction-league/service/internal/views"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
)

type testCommsAgentInjector struct {
	config    domain.Config
	queue     chan messages.Email
	templates *views.Templates
}

func (t testCommsAgentInjector) Config() domain.Config           { return t.config }
func (t testCommsAgentInjector) EmailQueue() chan messages.Email { return t.queue }
func (t testCommsAgentInjector) Template() *views.Templates      { return t.templates }

func TestCommunicationsAgent_IssuesNewEntryEmail(t *testing.T) {
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
		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		if err := agent.IssuesNewEntryEmail(&entry); err != nil {
			t.Fatal(err)
		}

		queue := agent.EmailQueue()
		close(queue)

		if len(queue) != 1 {
			expectedGot(t, 1, queue)
		}

		email := <-queue

		expectedPlainText := mustExecuteTemplate(t, templates, "email_txt_new_entry", emails.NewEntryEmailData{
			Name:           entry.EntrantName,
			SeasonName:     testSeason.Name,
			PredictionsURL: fmt.Sprintf("%s/prediction", testRealm.Origin),
			ShortCode:      entry.ShortCode,
			SignOff:        testRealm.Contact.Name,
			URL:            testRealm.Origin,
			SupportEmail:   testRealm.Contact.EmailProper,
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
		if email.Subject != domain.EmailSubjectNewEntry {
			expectedGot(t, domain.EmailSubjectNewEntry, email.Subject)
		}
		if email.PlainText != expectedPlainText {
			t.Fatal(gocmp.Diff(expectedPlainText, email.PlainText))
		}
	})

	t.Run("issue new entry email with an entry whose realm does not exist must fail", func(t *testing.T) {
		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.RealmName = "not_a_valid_realm"

		err := agent.IssuesNewEntryEmail(&entry)
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})

	t.Run("issue new entry email with an entry whose season does not exist must fail", func(t *testing.T) {
		entry := generateTestEntry(
			t,
			"Harry Redknapp",
			"Mr Harry R",
			"harry.redknapp@football.net",
		)

		entry.SeasonID = "not_a_valid_season"

		err := agent.IssuesNewEntryEmail(&entry)
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
