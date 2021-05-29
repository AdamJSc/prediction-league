package logger

import (
	"fmt"
	"io"
	"os"
	"prediction-league/service/internal/domain"
	"runtime"
	"strings"
	"time"
)

const (
	prefixError = "ERROR"
	prefixInfo  = "INFO"
)

// Logger defines a standard Logger
type Logger struct {
	domain.Logger
	w  io.Writer
	cl domain.Clock
}

// Info implements domain.Logger
func (l *Logger) Info(msg string) {
	l.w.Write([]byte(l.prefixMsgArgs(prefixInfo, msg)))
}

// Infof implements domain.Logger
func (l *Logger) Infof(msg string, a ...interface{}) {
	l.w.Write([]byte(l.prefixMsgArgs(prefixInfo, msg, a...)))
}

// Info implements domain.Logger
func (l *Logger) Error(msg string) {
	l.w.Write([]byte(l.prefixMsgArgs(prefixError, msg)))
}

// Errorf implements domain.Logger
func (l *Logger) Errorf(msg string, a ...interface{}) {
	l.w.Write([]byte(l.prefixMsgArgs(prefixError, msg, a...)))
}

func (l *Logger) prefixMsgArgs(prefix, msg string, a ...interface{}) string {
	msgWithArgs := fmt.Sprintf(msg, a...)
	ref := getCallerRef()
	ts := l.cl.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s %s: %s%s\n", ts, prefix, ref, msgWithArgs)
}

// NewLogger returns a new Logger using the provided writer
func NewLogger(w io.Writer, cl domain.Clock) (*Logger, error) {
	if w == nil {
		return nil, fmt.Errorf("writer: %w", domain.ErrIsNil)
	}
	if cl == nil {
		return nil, fmt.Errorf("clock: %w", domain.ErrIsNil)
	}
	return &Logger{w: w, cl: cl}, nil
}

// getCallerRef returns the line and filename of the function that called an exported logger method
func getCallerRef() string {
	_, fpath, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	parts := strings.Split(fpath, string(os.PathSeparator))
	fdir := ""
	fname := parts[len(parts)-1]
	if len(parts) >= 2 {
		fdir = parts[len(parts)-2] + string(os.PathSeparator)
	}
	return fmt.Sprintf("[%s%s:%d] ", fdir, fname, line)
}
