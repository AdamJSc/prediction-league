package httph

import "prediction-league/service/internal/app"

// HTTPAppContainer defines the http application dependencies
type HTTPAppContainer struct {
	app.DependencyInjector
}

// NewHTTPAppContainer returns a new instance of the http app container.
func NewHTTPAppContainer(dependencies app.DependencyInjector) *HTTPAppContainer {
	return &HTTPAppContainer{DependencyInjector: dependencies}
}
