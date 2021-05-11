package logger

import (
	"fmt"
	"io"
	"prediction-league/service/internal/domain"
)

const prefixInfo = "INFO"

// Logger defines a standard Logger
type Logger struct {
	domain.Logger
	w io.Writer
	c domain.Clock
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
	l.w.Write([]byte(l.prefixMsgArgs(prefixInfo, msg, a...)))
}

func (l *Logger) prefixMsgArgs(prefix, msg string, a ...interface{}) string {
	msgWithArgs := fmt.Sprintf(msg, a...)
	ts := l.c.Now().Format("2006/01/02 15:04:05")
	return fmt.Sprintf("%s %s: %s\n", ts, prefix, msgWithArgs)
}

// NewLogger returns a new Logger using the provided writer
func NewLogger(w io.Writer, c domain.Clock) (*Logger, error) {
	if w == nil {
		return nil, fmt.Errorf("writer: %w", domain.ErrIsNil)
	}
	if c == nil {
		return nil, fmt.Errorf("clock: %w", domain.ErrIsNil)
	}
	return &Logger{w: w, c: c}, nil
}
