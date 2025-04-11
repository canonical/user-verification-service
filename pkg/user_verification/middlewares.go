package user_verification

import (
	"net/http"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/tracing"
)

type AuthMiddleware struct {
	token string

	tracer tracing.TracingInterface
	logger logging.LoggerInterface
}

func (m *AuthMiddleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if m.token != token {
			m.logger.Error("Got invalid authorization header, rejecting request")
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewAuthMiddleware(token string, tracer tracing.TracingInterface, logger logging.LoggerInterface) *AuthMiddleware {
	m := new(AuthMiddleware)

	m.token = token

	m.tracer = tracer
	m.logger = logger
	return m
}
