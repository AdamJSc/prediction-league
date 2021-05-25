package logger

import (
	"fmt"
	"io"
	"os"
	"prediction-league/service/internal/domain"
	"runtime"
	"strings"
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

// Errorf implements domain.Logger
func (l *Logger) Errorf(msg string, a ...interface{}) {
	l.w.Write([]byte(l.prefixMsgArgs(prefixError, msg, a...)))
}

func (l *Logger) prefixMsgArgs(prefix, msg string, a ...interface{}) string {
	msgWithArgs := fmt.Sprintf(msg, a...)
	ref := getCallerRef()
	ts := l.cl.Now().Format("2006-01-02T15:04:05Z07:00")
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
	_, fpath, line, ok := runtime.Caller(4)
	if !ok {
		return ""
	}
	parts := strings.Split(fpath, string(os.PathSeparator))
	fname := parts[len(parts)-1]
	return fmt.Sprintf("[%s:%d] ", fname, line)
}
