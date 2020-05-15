package models

// ResourceIdentifier defines a generic interface for
// retrieving the value that identifies a resource
type ResourceIdentifier interface {
	Value() string
}
