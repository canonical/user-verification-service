package user_verification

import (
	"context"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/salesforce"
	"github.com/canonical/user-verification-service/internal/tracing"
)

type Service struct {
	salesforce salesforce.SalesforceAPI

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *Service) IsEmployee(ctx context.Context, email string) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "user_verification.Service.IsEmployee")
	defer span.End()

	return s.salesforce.IsEmployee(ctx, email)
}

func NewService(
	d salesforce.SalesforceAPI,
	tracer tracing.TracingInterface,
	monitor monitoring.MonitorInterface,
	logger logging.LoggerInterface,
) *Service {
	s := new(Service)

	s.salesforce = d

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
