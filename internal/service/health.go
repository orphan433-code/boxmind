package service

import (
	"context"

	"pet-link/internal/domain"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

type HealthService interface {
	Check(ctx context.Context) domain.HealthStatus
}

type healthService struct {
	serviceName string
	db          DBPinger
}

func NewHealthService(serviceName string, db DBPinger) HealthService {
	return &healthService{
		serviceName: serviceName,
		db:          db,
	}
}

func (s *healthService) Check(ctx context.Context) domain.HealthStatus {
	checks := make(map[string]string)

	if s.db == nil {
		checks["database"] = "not configured"
		return domain.HealthStatus{
			Status:  "degraded",
			Service: s.serviceName,
			Checks:  checks,
		}
	}

	if err := s.db.Ping(ctx); err != nil {
		checks["database"] = "unavailable"
		return domain.HealthStatus{
			Status:  "degraded",
			Service: s.serviceName,
			Checks:  checks,
		}
	}

	checks["database"] = "ok"
	return domain.HealthStatus{
		Status:  "ok",
		Service: s.serviceName,
		Checks:  checks,
	}
}
