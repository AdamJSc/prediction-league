package app

// HTTPAppContainer defines the http application dependencies
type HTTPAppContainer struct {
	DependencyInjector
}

// NewHTTPAppContainer returns a new instance of the http app container.
func NewHTTPAppContainer(deps DependencyInjector) *HTTPAppContainer {
	return &HTTPAppContainer{DependencyInjector: deps}
}
