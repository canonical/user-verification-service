package user_verification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/go-chi/chi/v5"
)

type ErrorID int

const (
	InvalidPayload ErrorID = 4200000 + iota
	APICallFailure
	NotInDirectory

	UserVerificationErrorDescription = "Account could not be verified.\n\nContact support"
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
	service ServiceInterface

	uiURL string

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Post("/api/v0/verify", a.handleVerify)
	mux.Get("/ui/registration_error", a.handleRegistration)
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
		a.logger.Errorf("Failed to check if user is employee: %v", payload.Email, err)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(
			WebhookErrorResponse{
				Messages: []errorMessage{{
					DetailedMessages: []detailedMessage{{
						ID:   APICallFailure,
						Text: "Failed to call the directory API",
						Type: "error",
					}},
				}},
			},
		)
		return
	}

	if !isEmployee {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(
			WebhookErrorResponse{
				Messages: []errorMessage{{
					InstancePtr: "#/traits/email",
					DetailedMessages: []detailedMessage{{
						ID:   NotInDirectory,
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

func (a *API) registrationURL(ErrorUiUrl, supportEmail string) string {
	u, err := url.Parse(ErrorUiUrl)
	if err != nil {
		panic(fmt.Errorf("invalid config login_ui_base_url: %v", err))
	}

	q := u.Query()
	var errorDescription string
	if supportEmail == "" {
		errorDescription = UserVerificationErrorDescription
	} else {
		errorDescription = fmt.Sprintf("%v at %v", UserVerificationErrorDescription, supportEmail)
	}

	q.Add("error_description", errorDescription)
	q.Add("error", "user_verification_failed")
	u.RawQuery = q.Encode()
	return u.String()
}

func (a *API) handleRegistration(w http.ResponseWriter, r *http.Request) {
	http.Redirect(
		w,
		r,
		a.uiURL,
		http.StatusSeeOther,
	)
}

func NewAPI(service ServiceInterface, ErrorUiUrl, supportEmail string, logger logging.LoggerInterface) *API {
	a := new(API)

	a.service = service
	a.uiURL = a.registrationURL(ErrorUiUrl, supportEmail)

	a.logger = logger

	return a
}
