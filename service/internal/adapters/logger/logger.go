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
	prefixDebug = "DEBUG"
	prefixError = "ERROR"
	prefixInfo  = "INFO"
)

const (
	LevelDebug level = iota
	LevelInfo
	LevelError
)

type level int

func (l level) isValid() bool {
	return l >= LevelDebug && l <= LevelError
}

// Logger defines a standard Logger
type Logger struct {
	lvl level
	w   io.Writer
	cl  domain.Clock
}

// Debugf implements domain.Logger
func (l *Logger) Debugf(msg string, a ...interface{}) {
	if l.lvl <= LevelDebug {
		l.w.Write([]byte(l.prefixMsgArgs(prefixDebug, msg, a...)))
	}
}

// Info implements domain.Logger
func (l *Logger) Info(msg string) {
	if l.lvl <= LevelInfo {
		l.w.Write([]byte(l.prefixMsgArgs(prefixInfo, msg)))
	}
}

// Infof implements domain.Logger
func (l *Logger) Infof(msg string, a ...interface{}) {
	if l.lvl <= LevelInfo {
		l.w.Write([]byte(l.prefixMsgArgs(prefixInfo, msg, a...)))
	}
}

// Info implements domain.Logger
func (l *Logger) Error(msg string) {
	if l.lvl <= LevelError {
		l.w.Write([]byte(l.prefixMsgArgs(prefixError, msg)))
	}
}

// Errorf implements domain.Logger
func (l *Logger) Errorf(msg string, a ...interface{}) {
	if l.lvl <= LevelError {
		l.w.Write([]byte(l.prefixMsgArgs(prefixError, msg, a...)))
	}
}

func (l *Logger) prefixMsgArgs(prefix, msg string, a ...interface{}) string {
	msgWithArgs := fmt.Sprintf(msg, a...)
	ref := getCallerRef()
	ts := l.cl.Now().Format(time.RFC3339)
	return fmt.Sprintf("%s %s: %s%s\n", ts, prefix, ref, msgWithArgs)
}

// NewLogger returns a new Logger using the provided writer
func NewLogger(l level, w io.Writer, cl domain.Clock) (*Logger, error) {
	if !l.isValid() {
		return nil, fmt.Errorf("level: %w", domain.ErrIsInvalid)
	}
	if w == nil {
		return nil, fmt.Errorf("writer: %w", domain.ErrIsNil)
	}
	if cl == nil {
		return nil, fmt.Errorf("clock: %w", domain.ErrIsNil)
	}
	return &Logger{l, w, cl}, nil
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
