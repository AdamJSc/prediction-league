package logger

import (
	"fmt"
	"io"
	"prediction-league/service/internal/domain"
)

const prefixInfo = "INFO"

// logger defines a standard logger
type logger struct {
	domain.Logger
	w io.Writer
}

// Info implements domain.Logger
func (l *logger) Info(msg string) {
	l.w.Write([]byte(prefixMsgArgs(prefixInfo, msg)))
}

// Infof implements domain.Logger
func (l *logger) Infof(msg string, a ...interface{}) {
	l.w.Write([]byte(prefixMsgArgs(prefixInfo, msg, a...)))
}

// NewLogger returns a new Logger using the provided writer
func NewLogger(w io.Writer) (*logger, error) {
	// TODO - add timestamp to logged message
	if w == nil {
		return nil, fmt.Errorf("writer: %w", domain.ErrIsNil)
	}
	return &logger{w: w}, nil
}

func prefixMsgArgs(prefix, msg string, a ...interface{}) string {
	msgWithArgs := fmt.Sprintf(msg, a...)
	return fmt.Sprintf("%s: %s\n", prefix, msgWithArgs)
}
