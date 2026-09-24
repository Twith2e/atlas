package auditlog

import "atlas/internal/domain"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateAuditLog(log *domain.AuditLog) error {
	return s.repo.CreateAuditLog(log)
}
