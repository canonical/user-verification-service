package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/salesforce"
	"github.com/canonical/user-verification-service/internal/tracing"
	"github.com/canonical/user-verification-service/pkg/metrics"
	"github.com/canonical/user-verification-service/pkg/status"
	"github.com/canonical/user-verification-service/pkg/ui"
	userVerification "github.com/canonical/user-verification-service/pkg/user_verification"
	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

func parseBaseURL(baseUrl string) *url.URL {
	if baseUrl[len(baseUrl)-1] != '/' {
		baseUrl += "/"
	}

	// Check if has app suburl.
	u, err := url.Parse(baseUrl)
	if err != nil {
		panic(fmt.Errorf("invalid BASE_URL: %v", err))
	}

	return u
}

func NewRouter(
	errorUiUrl,
	supportEmail,
	token,
	uiBaseURL string,
	d salesforce.SalesforceAPI,
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

	var authMiddleware *userVerification.AuthMiddleware = nil
	if token != "" {
		authMiddleware = userVerification.NewAuthMiddleware(token, tracer, logger)
	}

	uiRouter := chi.NewMux()

	userVerification.NewAPI(userVerification.NewService(d, tracer, monitor, logger), authMiddleware, logger).RegisterEndpoints(router)
	ui.NewAPI(errorUiUrl, supportEmail, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)
	status.NewAPI(tracer, monitor, logger).RegisterEndpoints(router)

	if uiBaseURL != "" {
		ui.NewAPI(errorUiUrl, supportEmail, logger).RegisterEndpoints(uiRouter)
		u := parseBaseURL(uiBaseURL)
		router.Mount(u.Path, uiRouter)
	}

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
