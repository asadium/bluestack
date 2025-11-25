package core

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Service is the interface that all Azure service emulators must implement.
// This allows for a clean, extensible architecture where new services can be
// added by simply implementing this interface and registering themselves.
type Service interface {
	// Name returns the unique identifier for this service (e.g., "blob", "queue", "keyvault").
	// This is used for service discovery and configuration.
	Name() string

	// RegisterRoutes sets up HTTP routes for this service on the provided router.
	// The router is typically a sub-router scoped to this service's path prefix.
	// Services should register their routes following Azure REST API patterns.
	RegisterRoutes(router chi.Router)
}

// serviceRegistry holds all registered services.
// This is a simple in-memory registry that can be extended later with
// dynamic loading, plugin support, etc.
type serviceRegistry struct {
	services []Service
}

var registry = &serviceRegistry{
	services: make([]Service, 0),
}

// RegisterService adds a service to the global registry.
// This should be called during application initialization, typically from
// the main function or an initialization function.
func RegisterService(s Service) {
	registry.services = append(registry.services, s)
}

// GetRegisteredServices returns all currently registered services.
// This is used by the edge router to set up routes for each service.
func GetRegisteredServices() []Service {
	return registry.services
}

// RequestContext provides common context information for service handlers.
// This can be extended with authentication, request ID, etc. as needed.
type RequestContext struct {
	Request  *http.Request
	Response http.ResponseWriter
}

