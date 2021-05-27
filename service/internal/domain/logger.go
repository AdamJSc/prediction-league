package domain

// Logger defines a custom logger interface
type Logger interface {
	Info(msg string)
	Infof(msg string, a ...interface{})
	Error(msg string)
	Errorf(msg string, a ...interface{})
}
