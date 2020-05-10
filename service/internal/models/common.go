package models

// ClientResourceIdentifier defines a generic interface for retrieving the value that
// identifies a resource for interacting with a client (i.e. external data source)
type ClientResourceIdentifier interface {
	Value() string
}
