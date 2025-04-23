package directoryapi

import (
	"net/http"
	"testing"

	"github.com/canonical/user-verification-service/internal/logging"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package directoryapi -destination ./mock_monitor.go -source=../monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package directoryapi -destination ./mock_tracing.go -source=../tracing/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package directoryapi -destination ./mock_directory_api.go -source=./interfaces.go

func TestIsEmployee(t *testing.T) {
	tests := []struct {
		name string

		token            string
		email            string
		mockedHttpClient func(*MockHttpClientInterface) HttpClientInterface

		expectedResult bool
		expectedError  error
	}{
		{
			name: "should succeed",
			mockedHttpClient: func(c *MockHttpClientInterface) HttpClientInterface {
				r := http.Response{}
				r.StatusCode = http.StatusOK
				c.EXPECT().Do(gomock.Any()).Times(1).Return(&r, nil)
				return c
			},
			expectedResult: true,
		},
		{
			name: "user not in directory",
			mockedHttpClient: func(c *MockHttpClientInterface) HttpClientInterface {
				r := http.Response{}
				r.StatusCode = http.StatusNotFound
				c.EXPECT().Do(gomock.Any()).Times(1).Return(&r, nil)
				return c
			},
			expectedResult: false,
		},
		{
			name: "invalid api token",
			mockedHttpClient: func(c *MockHttpClientInterface) HttpClientInterface {
				r := http.Response{}
				r.StatusCode = http.StatusUnauthorized
				c.EXPECT().Do(gomock.Any()).Times(1).Return(&r, nil)
				return c
			},
			expectedResult: false,
			expectedError:  ErrInvalidApiToken,
		},
		{
			name: "directory api error",
			mockedHttpClient: func(c *MockHttpClientInterface) HttpClientInterface {
				r := http.Response{}
				r.StatusCode = http.StatusInternalServerError
				c.EXPECT().Do(gomock.Any()).Times(1).Return(&r, nil)
				return c
			},
			expectedResult: false,
			expectedError:  ErrUnknownApiError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logging.NewNoopLogger()
			mockTracer := NewMockTracingInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockHttp := NewMockHttpClientInterface(ctrl)

			c := NewClient(false, "https://some-url", "token", mockTracer, mockMonitor, logger)
			c.http = test.mockedHttpClient(mockHttp)

			e, err := c.IsEmployee(t.Context(), "e@mail.com")

			if e != test.expectedResult {
				t.Fatalf("expected return value to be %v not %v", test.expectedResult, e)
			}

			if err != test.expectedError {
				t.Fatalf("expected error to be %v not %v", test.expectedError, err)
			}
		})
	}
}
