package auditlog

import (
	"atlas/internal/domain"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAuditLog(log *domain.AuditLog) error {
	query := `INSERT INTO audit_logs (public_id, resource_type, resource_id, actor_type, actor_id, action, ip_address, timestamp, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query, log.PublicID, log.ResourceType, log.ResourceID, log.ActorType, log.ActorID, log.Action, log.IPAddress, log.Timestamp, log.Metadata)
	return err
}
