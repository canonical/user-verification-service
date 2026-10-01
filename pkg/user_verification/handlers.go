// Copyright 2026 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0-only

package user_verification

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/go-chi/chi/v5"
)

type ErrorID int

// The login UI treats messages with an ID from 4200000 to 4299999 as a rejected
// registration and shows their text on its error page.
const (
	InvalidPayload ErrorID = 4200000 + iota
	APICallFailure
	NotFound
)

const errorDescription = "Account could not be verified.\n\nPlease try to log in again or contact support"

type WebhookPayload struct {
	Email string `json:"email"`
}

type detailedMessage struct {
	ID      ErrorID         `json:"id"`
	Text    string          `json:"text"`
	Type    string          `json:"type"`
	Context json.RawMessage `json:"context,omitempty"`
}

type errorMessage struct {
	InstancePtr      string            `json:"instance_ptr"`
	DetailedMessages []detailedMessage `json:"messages"`
}

// Taken from https://github.com/ory/kratos/blob/v1.3.1/selfservice/hook/web_hook.go#L106
type WebhookErrorResponse struct {
	Messages []errorMessage `json:"messages"`
}

type API struct {
	service    ServiceInterface
	middleware *AuthMiddleware

	// errorText is shown to the user, the cause of the failure is only logged.
	errorText string

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	if a.middleware != nil {
		mux = mux.With(a.middleware.AuthMiddleware).(*chi.Mux)
	}
	mux.Post("/api/v0/verify", a.handleVerify)
}

func (a *API) handleVerify(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload = new(WebhookPayload)

	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		a.logger.Error("Failed to parse payload: ", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			WebhookErrorResponse{
				Messages: []errorMessage{{
					DetailedMessages: []detailedMessage{{
						ID:   InvalidPayload,
						Text: a.errorText,
						Type: "error",
					}},
				}},
			},
		)
		return
	}

	isEmployee, err := a.service.IsEmployee(r.Context(), payload.Email)
	if err != nil {
		a.logger.Errorf("Failed to check if user '%v' is employee: %v", payload.Email, err)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(
			WebhookErrorResponse{
				Messages: []errorMessage{{
					DetailedMessages: []detailedMessage{{
						ID:   APICallFailure,
						Text: a.errorText,
						Type: "error",
					}},
				}},
			},
		)
		return
	}

	if !isEmployee {
		a.logger.Security().AuthzFailureNotEmployee(payload.Email, logging.WithRequest(r))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(
			WebhookErrorResponse{
				Messages: []errorMessage{{
					InstancePtr: "#/traits/email",
					DetailedMessages: []detailedMessage{{
						ID:   NotFound,
						Text: a.errorText,
						Type: "error",
					}},
				}},
			},
		)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
}

func NewAPI(service ServiceInterface, supportEmail string, middleware *AuthMiddleware, logger logging.LoggerInterface) *API {
	a := new(API)

	a.service = service
	a.errorText = errorDescription
	if supportEmail != "" {
		a.errorText = fmt.Sprintf("%v at %v", errorDescription, supportEmail)
	}
	if middleware != nil {
		a.middleware = middleware
	}

	a.logger = logger

	return a
}
