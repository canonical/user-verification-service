package salesforce

import (
	"fmt"
	"testing"

	"github.com/canonical/user-verification-service/internal/logging"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -build_flags=--mod=mod -package salesforce -destination ./mock_monitor.go -source=../monitoring/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package salesforce -destination ./mock_tracing.go -source=../tracing/interfaces.go
//go:generate mockgen -build_flags=--mod=mod -package salesforce -destination ./mock_salesforce.go -source=./interfaces.go

func TestIsEmployee(t *testing.T) {
	err := fmt.Errorf("some error")
	tests := []struct {
		name string

		token            string
		email            string
		mockedHttpClient func(*MockSalesforceClientAPI) SalesforceClientAPI

		expectedResult bool
		expectedError  error
	}{
		{
			name: "should succeed",
			mockedHttpClient: func(c *MockSalesforceClientAPI) SalesforceClientAPI {
				r := []Record{{Employment_Record_Active__c: true}}
				c.EXPECT().Query(gomock.Any(), gomock.Any()).Times(1).Return(nil).SetArg(1, r)
				return c
			},
			expectedResult: true,
		},
		{
			name: "user not in directory",
			mockedHttpClient: func(c *MockSalesforceClientAPI) SalesforceClientAPI {
				r := []Record{{Employment_Record_Active__c: false}}
				c.EXPECT().Query(gomock.Any(), gomock.Any()).Times(1).Return(nil).SetArg(1, r)
				return c
			},
			expectedResult: false,
		},
		{
			name: "salesforce error",
			mockedHttpClient: func(c *MockSalesforceClientAPI) SalesforceClientAPI {
				c.EXPECT().Query(gomock.Any(), gomock.Any()).Times(1).Return(err)
				return c
			},
			expectedResult: false,
			expectedError:  err,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logging.NewNoopLogger()
			mockTracer := NewMockTracingInterface(ctrl)
			mockMonitor := NewMockMonitorInterface(ctrl)
			mockSalesforce := NewMockSalesforceClientAPI(ctrl)

			c := Client{
				salesforceClient: test.mockedHttpClient(mockSalesforce),
				tracer:           mockTracer,
				monitor:          mockMonitor,
				logger:           logger,
			}

			mockMonitor.EXPECT().SetSalesforceResponseTimeMetric(gomock.Any(), gomock.Any()).AnyTimes()

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
