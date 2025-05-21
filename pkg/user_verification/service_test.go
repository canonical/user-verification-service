package user_verification

import (
	"context"
	"errors"
	"testing"

	"github.com/canonical/user-verification-service/internal/monitoring"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_logger.go -source=../../internal/logging/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_user_verification.go -source=./interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_monitor.go -source=../../internal/monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_tracing.go -source=../../internal/tracing/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package user_verification -destination ./mock_salesforce.go -source=../../internal/salesforce/interfaces.go

func TestServiceIsEmployee(t *testing.T) {
	dummyError := errors.New("some error")

	tests := []struct {
		name  string
		input string

		mockedServiceResult bool
		mockedServiceError  error

		expectedResult bool
		expectedError  error
	}{
		{
			name:                "Should fail because email is not an employee",
			input:               "a@a.com",
			mockedServiceResult: false,
			expectedResult:      false,
		},
		{
			name:               "Should fail because of error",
			input:              "a@a.com",
			mockedServiceError: dummyError,
			expectedError:      dummyError,
		},
		{
			name:           "Should fail because email is nil",
			expectedResult: false,
		},
		{
			name:                "Should succeed",
			input:               "a@a.com",
			mockedServiceResult: true,
			expectedResult:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := NewMockLoggerInterface(ctrl)
			mockTracer := NewMockTracingInterface(ctrl)
			mockMonitor := monitoring.NewMockMonitorInterface(ctrl)
			mockSalesforce := NewMockSalesforceAPI(ctrl)

			mockTracer.EXPECT().Start(gomock.Any(), "user_verification.Service.IsEmployee").Times(1).Return(context.TODO(), trace.SpanFromContext(context.TODO()))
			mockSalesforce.EXPECT().IsEmployee(gomock.Any(), test.input).Times(1).Return(test.mockedServiceResult, test.mockedServiceError)

			s := NewService(mockSalesforce, mockTracer, mockMonitor, mockLogger)

			b, err := s.IsEmployee(context.TODO(), test.input)

			if err != test.expectedError {
				t.Fatalf("expected error to be %v not %v", test.expectedError, err)
			}
			if b != test.expectedResult {
				t.Fatalf("expected return value to be %v not %v", test.expectedResult, b)
			}
		})
	}

}
