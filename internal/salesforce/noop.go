package salesforce

import (
	"context"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/tracing"
)

type NoopClient struct {
	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (c *NoopClient) IsEmployee(ctx context.Context, mail string) (bool, error) {
	// always return false in order to prevent security vulnerabilities when using this client by accident
	return false, nil
}

func NewNoopClient(
	tracer tracing.TracingInterface,
	monitor monitoring.MonitorInterface,
	logger logging.LoggerInterface,
) *NoopClient {
	c := new(NoopClient)
	c.logger = logger
	c.monitor = monitor
	c.tracer = tracer
	return c
}
