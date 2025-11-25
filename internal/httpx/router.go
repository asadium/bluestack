package httpx

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/asad/bluestack/internal/config"
	"github.com/asad/bluestack/internal/core"
	"github.com/asad/bluestack/internal/logging"
)

// EdgeRouter is the main HTTP router that receives all incoming requests
// and dispatches them to the appropriate service modules.
// It acts as a single entry point, similar to LocalStack's edge service.
type EdgeRouter struct {
	router chi.Router
	cfg    *config.Config
	logger logging.Logger
}

// NewEdgeRouter creates and configures a new edge router instance.
// It sets up middleware for logging, request ID, recovery, etc.
// Services should be registered via core.RegisterService() before calling this.
func NewEdgeRouter(cfg *config.Config, logger logging.Logger) http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLoggingMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint - always available regardless of enabled services
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"bluestack"}`))
	})

	// Register routes for each enabled service
	services := core.GetRegisteredServices()
	for _, service := range services {
		if cfg.IsServiceEnabled(service.Name()) {
			logger.Info("registering service routes",
				logging.String("service", service.Name()),
			)

			// Each service gets its own sub-router
			// For now, we use a simple prefix pattern. In the future, this could
			// be more sophisticated (e.g., host-based routing for account-specific endpoints)
			r.Route("/"+service.Name(), func(r chi.Router) {
				service.RegisterRoutes(r)
			})
		} else {
			logger.Info("skipping service (not enabled)",
				logging.String("service", service.Name()),
			)
		}
	}

	return &EdgeRouter{
		router: r,
		cfg:    cfg,
		logger: logger,
	}
}

// ServeHTTP implements http.Handler interface.
func (er *EdgeRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	er.router.ServeHTTP(w, r)
}

// requestLoggingMiddleware creates middleware that logs HTTP requests with
// structured logging including method, path, status code, and latency.
func requestLoggingMiddleware(logger logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process request
			next.ServeHTTP(ww, r)

			// Log request details
			duration := time.Since(start)
			logger.Info("request completed",
				logging.String("method", r.Method),
				logging.String("path", r.URL.Path),
				logging.String("query", r.URL.RawQuery),
				logging.Int("status", ww.Status()),
				logging.Duration("latency_ms", duration.Milliseconds()),
				logging.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

