package logger_test

import (
	"bytes"
	"errors"
	"log"
	"os"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

var testDate time.Time

func TestMain(m *testing.M) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatalf("cannot load time location: %s", err.Error())
	}
	testDate = time.Date(2018, 5, 26, 14, 0, 0, 0, loc)
	os.Exit(m.Run())
}

type mockWriter struct {
	buf *bytes.Buffer
}

func (m *mockWriter) Write(b []byte) (int, error) {
	return m.buf.Write(b)
}

type mockClock struct {
	t time.Time
}

func (m *mockClock) Now() time.Time {
	return m.t
}

func TestNewLogger(t *testing.T) {
	t.Run("passing nil must return the expected error", func(t *testing.T) {
		if _, gotErr := logger.NewLogger(nil, &domain.RealClock{}); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}
		if _, gotErr := logger.NewLogger(os.Stdout, nil); !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestLogger_Info(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}
	l, _ := logger.NewLogger(wr, c)

	l.Info("hello world")

	want := "2018/05/26 14:00:00 INFO: hello world\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestLogger_Infof(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}
	l, _ := logger.NewLogger(wr, c)

	l.Infof("hello %d", 123)

	want := "2018/05/26 14:00:00 INFO: hello 123\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
