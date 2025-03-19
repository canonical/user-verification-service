package user_verification

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_user_verification.go -source=./interfaces.go

func TestHandleDeviceUserCodeAcceptSuccess(t *testing.T) {
	type serviceResult struct {
		r   bool
		err error
	}

	tests := []struct {
		name  string
		input string

		result *serviceResult

		expectedStatus int
	}{
		{
			name:           "Should fail because no email provided",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Should fail because is not employee",
			input:          "not@employee.com",
			result:         &serviceResult{r: false},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Should fail because of service error",
			input:          "not@employee.com",
			result:         &serviceResult{err: errors.New("some error")},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Should succeed",
			input:          "not@employee.com",
			result:         &serviceResult{r: true},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logging.NewLogger("DEBUG")
			mockService := NewMockServiceInterface(ctrl)

			if test.result != nil {
				mockService.EXPECT().IsEmployee(gomock.Any(), test.input).Times(1).Return(test.result.r, test.result.err)
			}

			body := []byte("")
			if test.input != "" {
				body, _ = json.Marshal(WebhookPayload{Email: test.input})
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify", bytes.NewBuffer(body))

			mux := chi.NewMux()
			NewAPI(mockService, logger).RegisterEndpoints(mux)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode != test.expectedStatus {
				t.Fatalf("expected status to be %v not %v", test.expectedStatus, res.StatusCode)
			}
		})
	}
}
