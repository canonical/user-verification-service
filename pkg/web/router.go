// Copyright 2026 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"net/http"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/salesforce"
	"github.com/canonical/user-verification-service/internal/tracing"
	"github.com/canonical/user-verification-service/pkg/metrics"
	"github.com/canonical/user-verification-service/pkg/status"
	userVerification "github.com/canonical/user-verification-service/pkg/user_verification"
	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	supportEmail,
	token string,
	sf salesforce.SalesforceAPI,
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

	userVerification.NewAPI(userVerification.NewService(sf, tracer, monitor, logger), supportEmail, authMiddleware, logger).RegisterEndpoints(router)
	metrics.NewAPI(logger).RegisterEndpoints(router)
	status.NewAPI(tracer, monitor, logger).RegisterEndpoints(router)

	return tracing.NewMiddleware(monitor, logger).OpenTelemetry(router)
}
