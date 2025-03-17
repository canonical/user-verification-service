package user_verification

import (
	"context"

	directoryapi "github.com/canonical/user-verification-service/internal/directory_api"
	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/tracing"
)

type Service struct {
	directoryAPI directoryapi.DirectoryAPI

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func (s *Service) IsEmployee(ctx context.Context, email string) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "user_verification.Service.IsEmployee")
	defer span.End()

	return s.directoryAPI.IsEmployee(ctx, email)
}

func NewService(
	d directoryapi.DirectoryAPI,
	tracer tracing.TracingInterface,
	monitor monitoring.MonitorInterface,
	logger logging.LoggerInterface,
) *Service {
	s := new(Service)

	s.directoryAPI = d

	s.monitor = monitor
	s.tracer = tracer
	s.logger = logger

	return s
}
