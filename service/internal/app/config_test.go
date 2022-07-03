package app_test

import (
	"errors"
	"fmt"
	"prediction-league/service/internal/app"
	"prediction-league/service/internal/domain"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewConfigFromOptions(t *testing.T) {
	t.Run("must set the expected values successfully", func(t *testing.T) {
		option1 := app.ConfigOption(func(config *app.Config) error {
			config.ServicePort = "foo"
			return nil
		})

		option2 := app.ConfigOption(func(config *app.Config) error {
			config.LogLevel = "bar"
			return nil
		})

		wantConfig := &app.Config{
			ServicePort: "foo",
			LogLevel:    "bar",
		}

		gotConfig, err := app.NewConfigFromOptions(option1, option2)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(wantConfig, gotConfig); diff != "" {
			t.Fatalf("mismatch: %s", diff)
		}
	})

	t.Run("must return the expected error", func(t *testing.T) {
		err := errors.New("sad times :'(")

		option1 := app.ConfigOption(func(config *app.Config) error {
			config.ServicePort = "foo"
			return nil
		})

		option2 := app.ConfigOption(func(config *app.Config) error {
			config.LogLevel = "bar"
			return nil
		})

		optionErr := app.ConfigOption(func(config *app.Config) error {
			return err
		})

		if _, gotErr := app.NewConfigFromOptions(option1, optionErr, option2); !errors.Is(gotErr, err) {
			t.Fatalf("want error %+v, got %+v", err, gotErr)
		}
	})
}

func TestNewLoadEnvConfigOption(t *testing.T) {
	t.Run("must load config from env successfully and skip non-existent path", func(t *testing.T) {
		l := &mockLogger{}

		opt := app.NewLoadEnvConfigOption(l, "testdata/config_test.env", "non_existent_path")

		wantConfig := &app.Config{
			ServicePort:          "1234",
			MySQLURL:             "test-db-user:test-db-pwd@tcp(localhost:3306)/test-db-name?parseTime=true",
			MigrationsPath:       "test_migrations_url",
			AdminBasicAuth:       "test_admin_basic_auth",
			LogLevel:             "test_loglevel",
			FootballDataAPIToken: "test_football_data_api_token",
			PayPalClientID:       "test_paypal_client_id",
			MailgunAPIKey:        "test_mailgun_api_key",
		}

		gotConfig := &app.Config{}
		if err := opt(gotConfig); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(wantConfig, gotConfig); diff != "" {
			t.Fatalf("config mismatch: %s", diff)
		}

		wantLogMsgs := map[string][]string{
			"infof": []string{
				"skipping env file 'non_existent_path': not found",
			},
		}

		gotLogMsgs := l.msgs

		if diff := cmp.Diff(wantLogMsgs, gotLogMsgs); diff != "" {
			t.Fatalf("logs mismatch: %s", diff)
		}
	})
}

func TestNewBuildVersionConfigOption(t *testing.T) {
	t.Run("must set config value successfully", func(t *testing.T) {
		opt := app.NewBuildVersionConfigOption("v1.2.3")

		wantConfig := &app.Config{
			BuildVersion: "v1.2.3",
		}

		gotConfig := &app.Config{}
		if err := opt(gotConfig); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(wantConfig, gotConfig); diff != "" {
			t.Fatalf("config mismatch: %s", diff)
		}
	})
}

func TestNewBuildTimestampConfigOption(t *testing.T) {
	t.Run("must set config value successfully", func(t *testing.T) {
		opt := app.NewBuildTimestampConfigOption("1970-01-01T00:00:00Z")

		wantConfig := &app.Config{
			BuildTimestamp: "1970-01-01T00:00:00Z",
		}

		gotConfig := &app.Config{}
		if err := opt(gotConfig); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(wantConfig, gotConfig); diff != "" {
			t.Fatalf("config mismatch: %s", diff)
		}
	})
}

type mockLogger struct {
	domain.Logger
	msgs map[string][]string
}

func (m *mockLogger) init(level string) {
	if m.msgs == nil {
		m.msgs = make(map[string][]string)
	}

	if _, ok := m.msgs[level]; !ok {
		m.msgs[level] = make([]string, 0)
	}
}

func (m *mockLogger) addMsg(level, msg string) {
	m.init(level)
	m.msgs[level] = append(m.msgs[level], msg)
}

func (m *mockLogger) Infof(msg string, a ...interface{}) {
	m.addMsg("infof", fmt.Sprintf(msg, a...))
}
