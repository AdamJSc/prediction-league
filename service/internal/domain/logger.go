package domain

// Logger defines a custom logger interface
type Logger interface {
	Info(msg string)
	Infof(msg string, a ...interface{})
	Errorf(msg string, a ...interface{})
}
