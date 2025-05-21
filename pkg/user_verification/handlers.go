package user_verification

import (
	"encoding/json"
	"net/http"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/go-chi/chi/v5"
)

type ErrorID int

const (
	InvalidPayload ErrorID = 4200000 + iota
	APICallFailure
	NotFound
)

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
						Text: "Invalid payload",
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
						Text: "Failed to call the salesforce API",
						Type: "error",
					}},
				}},
			},
		)
		return
	}

	if !isEmployee {
		a.logger.Errorf("User '%v' is not an employee", payload.Email)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(
			WebhookErrorResponse{
				Messages: []errorMessage{{
					InstancePtr: "#/traits/email",
					DetailedMessages: []detailedMessage{{
						ID:   NotFound,
						Text: "User is not an employee",
						Type: "error",
					}},
				}},
			},
		)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func NewAPI(service ServiceInterface, middleware *AuthMiddleware, logger logging.LoggerInterface) *API {
	a := new(API)

	a.service = service
	if middleware != nil {
		a.middleware = middleware
	}

	a.logger = logger

	return a
}
