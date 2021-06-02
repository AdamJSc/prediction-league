package logger

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
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
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		tt := []struct {
			lvl     level
			w       io.Writer
			cl      domain.Clock
			wantErr error
		}{
			{level(123), os.Stdout, &domain.RealClock{}, domain.ErrIsInvalid},
			{LevelDebug, nil, &domain.RealClock{}, domain.ErrIsNil},
			{LevelDebug, os.Stdout, nil, domain.ErrIsNil},
			{LevelDebug, os.Stdout, &domain.RealClock{}, nil},
		}
		for idx, tc := range tt {
			l, gotErr := NewLogger(tc.lvl, tc.w, tc.cl)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && l == nil {
				t.Fatalf("tc #%d: want non-empty logger, got nil", idx)
			}
		}
	})
}

func TestLogger_Debugf(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}

	tt := []struct {
		lvl       level
		shouldLog bool
	}{
		{LevelDebug, true},
		{LevelInfo, true},
		{LevelError, true},
	}

	for idx, tc := range tt {
		l, _ := NewLogger(tc.lvl, wr, c)
		l.Debugf("hello %d", 123)
		var wantOut string
		if tc.shouldLog {
			wantOut = "2018-05-26T14:00:00+01:00 DEBUG: [logger/logger_test.go:81] hello 123\n"
		}
		got := wr.buf.String()
		if got != wantOut {
			t.Fatalf("tc #%d: want %s, got %s", idx, wantOut, got)
		}
		wr.buf.Reset()
	}
}

func TestLogger_Info(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}

	tt := []struct {
		lvl       level
		shouldLog bool
	}{
		{LevelDebug, false},
		{LevelInfo, true},
		{LevelError, true},
	}

	for idx, tc := range tt {
		l, _ := NewLogger(tc.lvl, wr, c)
		l.Info("hello world")
		var wantOut string
		if tc.shouldLog {
			wantOut = "2018-05-26T14:00:00+01:00 INFO: [logger/logger_test.go:109] hello world\n"
		}
		got := wr.buf.String()
		if got != wantOut {
			t.Fatalf("tc #%d: want %s, got %s", idx, wantOut, got)
		}
		wr.buf.Reset()
	}
}

func TestLogger_Infof(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}

	tt := []struct {
		lvl       level
		shouldLog bool
	}{
		{LevelDebug, false},
		{LevelInfo, true},
		{LevelError, true},
	}

	for idx, tc := range tt {
		l, _ := NewLogger(tc.lvl, wr, c)
		l.Infof("hello %d", 123)
		var wantOut string
		if tc.shouldLog {
			wantOut = "2018-05-26T14:00:00+01:00 INFO: [logger/logger_test.go:137] hello 123\n"
		}
		got := wr.buf.String()
		if got != wantOut {
			t.Fatalf("tc #%d: want %s, got %s", idx, wantOut, got)
		}
		wr.buf.Reset()
	}
}

func TestLogger_Error(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}

	tt := []struct {
		lvl       level
		shouldLog bool
	}{
		{LevelDebug, false},
		{LevelInfo, false},
		{LevelError, true},
	}

	for idx, tc := range tt {
		l, _ := NewLogger(tc.lvl, wr, c)
		l.Error("hello world")
		var wantOut string
		if tc.shouldLog {
			wantOut = "2018-05-26T14:00:00+01:00 ERROR: [logger/logger_test.go:165] hello world\n"
		}
		got := wr.buf.String()
		if got != wantOut {
			t.Fatalf("tc #%d: want %s, got %s", idx, wantOut, got)
		}
		wr.buf.Reset()
	}
}

func TestLogger_Errorf(t *testing.T) {
	wr := &mockWriter{buf: &bytes.Buffer{}}
	c := &mockClock{t: testDate}

	tt := []struct {
		lvl       level
		shouldLog bool
	}{
		{LevelDebug, false},
		{LevelInfo, false},
		{LevelError, true},
	}

	for idx, tc := range tt {
		l, _ := NewLogger(tc.lvl, wr, c)
		l.Errorf("hello %d", 123)
		var wantOut string
		if tc.shouldLog {
			wantOut = "2018-05-26T14:00:00+01:00 ERROR: [logger/logger_test.go:193] hello 123\n"
		}
		got := wr.buf.String()
		if got != wantOut {
			t.Fatalf("tc #%d: want %s, got %s", idx, wantOut, got)
		}
		wr.buf.Reset()
	}
}
