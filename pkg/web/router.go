package web

import (
	"net/http"

	directoryapi "github.com/canonical/user-verification-service/internal/directory_api"
	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/tracing"
	"github.com/canonical/user-verification-service/pkg/metrics"
	userVerification "github.com/canonical/user-verification-service/pkg/user_verification"
	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	errorUiUrl,
	supportEmail string,
	d directoryapi.DirectoryAPI,
	tracer tracing.TracingInterface,
	monitor monitoring.MonitorInterface,
	logger logging.LoggerInterface,
) http.Handler {
	router := chi.NewMux()

	middlewares := make(chi.Middlewares, 0)
	middlewares = append(
		middlewares,
		middleware.RequestID,
		monitoring.NewMiddleware(monitor, logger).ResponseTime(),
		middlewareCORS([]string{"*"}),
	)

	if true {
		middlewares = append(
			middlewares,
			middleware.RequestLogger(logging.NewLogFormatter(logger)), // LogFormatter will only work if logger is set to DEBUG level
		)
	}

	router.Use(middlewares...)

	userVerification.NewAPI(userVerification.NewService(d, tracer, monitor, logger), errorUiUrl, supportEmail, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
