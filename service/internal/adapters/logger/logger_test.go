package logger_test

import (
	"bytes"
	"errors"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
)

type mockWriter struct {
	buf *bytes.Buffer
}

func (m *mockWriter) Write(b []byte) (int, error) {
	return m.buf.Write(b)
}

func TestNewLogger(t *testing.T) {
	t.Run("passing nil must return the expected error", func(t *testing.T) {
		if _, gotErr := logger.NewLogger(nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestLogger_Info(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	l, _ := logger.NewLogger(wr)

	l.Info("hello world")

	want := "INFO: hello world\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestLogger_Infof(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	l, _ := logger.NewLogger(wr)

	l.Infof("hello %d", 123)

	want := "INFO: hello 123\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
