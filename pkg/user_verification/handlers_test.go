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

func TestSendWebhookError(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		errorID     ErrorID
		text        string
		instancePtr string
		wantStatus  int
		wantErrID   ErrorID
		wantText    string
		wantPtr     string
	}{
		{
			name:        "Send invalid payload error",
			statusCode:  http.StatusBadRequest,
			errorID:     InvalidPayload,
			text:        "Invalid payload",
			instancePtr: "",
			wantStatus:  http.StatusBadRequest,
			wantErrID:   InvalidPayload,
			wantText:    "Invalid payload",
			wantPtr:     "",
		},
		{
			name:        "Send API call failure error",
			statusCode:  http.StatusForbidden,
			errorID:     APICallFailure,
			text:        "Failed to call the salesforce API",
			instancePtr: "",
			wantStatus:  http.StatusForbidden,
			wantErrID:   APICallFailure,
			wantText:    "Failed to call the salesforce API",
			wantPtr:     "",
		},
		{
			name:        "Send not employee error with instance pointer",
			statusCode:  http.StatusForbidden,
			errorID:     NotFound,
			text:        "User is not an employee",
			instancePtr: "#/traits/email",
			wantStatus:  http.StatusForbidden,
			wantErrID:   NotFound,
			wantText:    "User is not an employee",
			wantPtr:     "#/traits/email",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			sendWebhookError(w, test.statusCode, test.errorID, test.text, test.instancePtr)

			if w.Code != test.wantStatus {
				t.Errorf("status code = %v, want %v", w.Code, test.wantStatus)
			}

			var response WebhookErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(response.Messages) != 1 {
				t.Fatalf("expected 1 message, got %d", len(response.Messages))
			}

			msg := response.Messages[0]
			if msg.InstancePtr != test.wantPtr {
				t.Errorf("instance_ptr = %v, want %v", msg.InstancePtr, test.wantPtr)
			}

			if len(msg.DetailedMessages) != 1 {
				t.Fatalf("expected 1 detailed message, got %d", len(msg.DetailedMessages))
			}

			detail := msg.DetailedMessages[0]
			if detail.ID != test.wantErrID {
				t.Errorf("error ID = %v, want %v", detail.ID, test.wantErrID)
			}
			if detail.Text != test.wantText {
				t.Errorf("text = %v, want %v", detail.Text, test.wantText)
			}
			if detail.Type != "error" {
				t.Errorf("type = %v, want error", detail.Type)
			}
		})
	}
}

func TestParsePayload(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantEmail string
		wantErr   bool
	}{
		{
			name:      "Valid payload",
			body:      `{"email":"test@example.com"}`,
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:    "Invalid JSON",
			body:    `{invalid json}`,
			wantErr: true,
		},
		{
			name:      "Empty email",
			body:      `{"email":""}`,
			wantEmail: "",
			wantErr:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify", bytes.NewBufferString(test.body))
			payload, err := parsePayload(req)

			if test.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if payload.Email != test.wantEmail {
				t.Errorf("email = %v, want %v", payload.Email, test.wantEmail)
			}
		})
	}
}

func TestVerifyEmployee(t *testing.T) {
	tests := []struct {
		name              string
		email             string
		serviceResult     bool
		serviceErr        error
		wantResult        bool
		wantErr           bool
		expectSecurityLog bool
	}{
		{
			name:              "Employee found",
			email:             "employee@example.com",
			serviceResult:     true,
			serviceErr:        nil,
			wantResult:        true,
			wantErr:           false,
			expectSecurityLog: false,
		},
		{
			name:              "Not an employee",
			email:             "notemployee@example.com",
			serviceResult:     false,
			serviceErr:        nil,
			wantResult:        false,
			wantErr:           false,
			expectSecurityLog: true,
		},
		{
			name:              "Service error",
			email:             "error@example.com",
			serviceResult:     false,
			serviceErr:        errors.New("service error"),
			wantResult:        false,
			wantErr:           true,
			expectSecurityLog: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockSecurityLogger := NewMockSecurityLoggerInterface(ctrl)
			mockService := NewMockServiceInterface(ctrl)

			mockService.EXPECT().IsEmployee(gomock.Any(), test.email).Return(test.serviceResult, test.serviceErr)

			if test.serviceErr != nil {
				mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			}

			if test.expectSecurityLog {
				mockLogger.EXPECT().Security().Return(mockSecurityLogger)
				mockSecurityLogger.EXPECT().AuthzFailureNotEmployee(test.email, gomock.Any()).Times(1)
			}

			api := NewAPI(mockService, nil, mockLogger)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/verify", nil)

			result, err := api.verifyEmployee(req, test.email)

			if test.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if result != test.wantResult {
				t.Errorf("result = %v, want %v", result, test.wantResult)
			}
		})
	}
}

func TestHandleVerify(t *testing.T) {
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
			NewAPI(mockService, nil, mockLogger).RegisterEndpoints(mux)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode != test.expectedStatus {
				t.Fatalf("expected status to be %v not %v", test.expectedStatus, res.StatusCode)
			}
		})
	}
}
