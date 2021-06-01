package logger_test

import (
	"bytes"
	"errors"
	"io"
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
	t.Run("passing nil must return expected error", func(t *testing.T) {
		tt := []struct {
			w       io.Writer
			cl      domain.Clock
			wantErr bool
		}{
			{nil, &domain.RealClock{}, true},
			{os.Stdout, nil, true},
			{os.Stdout, &domain.RealClock{}, false},
		}
		for idx, tc := range tt {
			l, gotErr := logger.NewLogger(tc.w, tc.cl)
			if tc.wantErr && !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("tc #%d: want ErrIsNil, got %s (%T)", idx, gotErr, gotErr)
			}
			if !tc.wantErr && l == nil {
				t.Fatalf("tc #%d: want non-empty logger, got nil", idx)
			}
		}
	})
}

func TestLogger_Debugf(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}
	l, _ := logger.NewLogger(wr, c)

	l.Debugf("hello %d", 123)

	want := "2018-05-26T14:00:00+01:00 DEBUG: [logger/logger_test.go:70] hello 123\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestLogger_Info(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}
	l, _ := logger.NewLogger(wr, c)

	l.Info("hello world")

	want := "2018-05-26T14:00:00+01:00 INFO: [logger/logger_test.go:85] hello world\n"
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

	want := "2018-05-26T14:00:00+01:00 INFO: [logger/logger_test.go:100] hello 123\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestLogger_Error(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}
	l, _ := logger.NewLogger(wr, c)

	l.Error("hello world")

	want := "2018-05-26T14:00:00+01:00 ERROR: [logger/logger_test.go:115] hello world\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestLogger_Errorf(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}
	l, _ := logger.NewLogger(wr, c)

	l.Errorf("hello %d", 123)

	want := "2018-05-26T14:00:00+01:00 ERROR: [logger/logger_test.go:130] hello 123\n"
	got := wr.buf.String()

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
