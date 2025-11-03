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

// sendWebhookError writes a WebhookErrorResponse with the given status code
func sendWebhookError(w http.ResponseWriter, statusCode int, errorID ErrorID, text string, instancePtr string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(
		WebhookErrorResponse{
			Messages: []errorMessage{{
				InstancePtr: instancePtr,
				DetailedMessages: []detailedMessage{{
					ID:   errorID,
					Text: text,
					Type: "error",
				}},
			}},
		},
	)
}

// parsePayload decodes the WebhookPayload from the request body
func parsePayload(r *http.Request) (*WebhookPayload, error) {
	var payload = new(WebhookPayload)
	err := json.NewDecoder(r.Body).Decode(payload)
	return payload, err
}

// verifyEmployee checks if the email belongs to an employee and handles security logging
func (a *API) verifyEmployee(r *http.Request, email string) (bool, error) {
	isEmployee, err := a.service.IsEmployee(r.Context(), email)
	if err != nil {
		a.logger.Errorf("Failed to check if user '%v' is employee: %v", email, err)
		return false, err
	}

	if !isEmployee {
		a.logger.Security().AuthzFailureNotEmployee(email, logging.WithRequest(r))
	}

	return isEmployee, nil
}

func (a *API) handleVerify(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Parse the payload
	payload, err := parsePayload(r)
	if err != nil {
		a.logger.Error("Failed to parse payload: ", err)
		sendWebhookError(w, http.StatusBadRequest, InvalidPayload, "Invalid payload", "")
		return
	}

	// Verify the user is an employee
	isEmployee, err := a.verifyEmployee(r, payload.Email)
	if err != nil {
		sendWebhookError(w, http.StatusForbidden, APICallFailure, "Failed to call the salesforce API", "")
		return
	}

	if !isEmployee {
		sendWebhookError(w, http.StatusForbidden, NotFound, "User is not an employee", "#/traits/email")
		return
	}

	// Success: echo the original payload
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
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
