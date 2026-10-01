// Copyright 2026 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0-only

package user_verification

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_user_verification.go -source=./interfaces.go

func TestHandleVerify(t *testing.T) {
	type serviceResult struct {
		r   bool
		err error
	}

	tests := []struct {
		name         string
		input        string
		supportEmail string

		result *serviceResult

		expectedStatus int
		expectedError  *detailedMessage
	}{
		{
			name:           "Should fail because no email provided",
			expectedStatus: http.StatusBadRequest,
			expectedError:  &detailedMessage{ID: InvalidPayload, Text: errorDescription, Type: "error"},
		},
		{
			name:           "Should fail because is not employee",
			input:          "not@employee.com",
			result:         &serviceResult{r: false},
			expectedStatus: http.StatusForbidden,
			expectedError:  &detailedMessage{ID: NotFound, Text: errorDescription, Type: "error"},
		},
		{
			name:           "Should fail with the support email in the error",
			input:          "not@employee.com",
			supportEmail:   "support@example.com",
			result:         &serviceResult{r: false},
			expectedStatus: http.StatusForbidden,
			expectedError:  &detailedMessage{ID: NotFound, Text: errorDescription + " at support@example.com", Type: "error"},
		},
		{
			name:           "Should fail because of service error",
			input:          "not@employee.com",
			result:         &serviceResult{err: errors.New("some error")},
			expectedStatus: http.StatusForbidden,
			expectedError:  &detailedMessage{ID: APICallFailure, Text: errorDescription, Type: "error"},
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

			mockLogger := NewMockLoggerInterface(ctrl)
			mockSecurityLogger := NewMockSecurityLoggerInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			if test.result != nil {
				mockService.EXPECT().IsEmployee(gomock.Any(), test.input).Times(1).Return(test.result.r, test.result.err)
				if !test.result.r && test.result.err == nil {
					mockLogger.EXPECT().Security().Return(mockSecurityLogger)
					mockSecurityLogger.EXPECT().AuthzFailureNotEmployee(gomock.Any(), gomock.Any()).AnyTimes()
				}
			}

			if test.expectedStatus != http.StatusOK {
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
			}

			body := []byte("")
			if test.input != "" {
				body, _ = json.Marshal(WebhookPayload{Email: test.input})
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify", bytes.NewBuffer(body))

			mux := chi.NewMux()
			NewAPI(mockService, test.supportEmail, nil, mockLogger).RegisterEndpoints(mux)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode != test.expectedStatus {
				t.Fatalf("expected status to be %v not %v", test.expectedStatus, res.StatusCode)
			}

			if test.expectedError == nil {
				return
			}

			var errorResponse WebhookErrorResponse
			if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("unexpected decode error: %v", err)
			}
			if len(errorResponse.Messages) != 1 || len(errorResponse.Messages[0].DetailedMessages) != 1 {
				t.Fatalf("expected a single error message, got %v", errorResponse.Messages)
			}
			if actual := errorResponse.Messages[0].DetailedMessages[0]; actual.ID != test.expectedError.ID ||
				actual.Text != test.expectedError.Text || actual.Type != test.expectedError.Type {
				t.Fatalf("expected error to be %v not %v", *test.expectedError, actual)
			}
		})
	}
}
