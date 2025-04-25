package ui

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/go-chi/chi/v5"
)

const UserVerificationErrorDescription = "Account could not be verified.\n\nPlease try to log in again or contact support"

type API struct {
	uiURL string

	logger logging.LoggerInterface
}

func (a *API) RegisterEndpoints(mux *chi.Mux) {
	mux.Get("/ui/registration_error", a.handleRegistration)
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

func NewAPI(errorUiUrl, supportEmail string, logger logging.LoggerInterface) *API {
	a := new(API)

	a.uiURL = a.registrationURL(errorUiUrl, supportEmail)

	a.logger = logger

	return a
}
