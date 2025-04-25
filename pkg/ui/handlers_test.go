package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

func TestHandleRegistrationError(t *testing.T) {
	type serviceResult struct {
		r   bool
		err error
	}

	tests := []struct {
		name string

		errorURL     string
		supportEmail string

		expectedURL string
	}{
		{
			name:         "Should include the email",
			supportEmail: "contact@support.com",
			errorURL:     "http://path/to/error",
			expectedURL:  "http://path/to/error?error=user_verification_failed&error_description=Account+could+not+be+verified.%0A%0APlease+try+to+log+in+again+or+contact+support+at+contact%40support.com",
		},
		{
			name:        "Should not include the email",
			errorURL:    "http://path/to/error",
			expectedURL: "http://path/to/error?error=user_verification_failed&error_description=Account+could+not+be+verified.%0A%0APlease+try+to+log+in+again+or+contact+support",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logging.NewLogger("DEBUG")

			req := httptest.NewRequest(http.MethodGet, "/ui/registration_error", nil)

			mux := chi.NewMux()
			NewAPI(test.errorURL, test.supportEmail, logger).RegisterEndpoints(mux)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode != http.StatusSeeOther {
				t.Fatalf("expected status to be %v not %v", http.StatusSeeOther, res.StatusCode)
			}

			loc, err := res.Location()
			if err != nil {
				t.Fatalf("Failed to parse location header: %v", err)
			}
			if loc.String() != test.expectedURL {
				t.Fatalf("expected Location to be %v not %v", test.expectedURL, loc.String())
			}
		})
	}
}
