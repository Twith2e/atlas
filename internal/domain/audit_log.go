package domain

type AuditLog struct {
	ID           int64  `json:"id"`
	PublicID     string `json:"public_id"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	ActorType    string `json:"actor_type"`
	ActorID      string `json:"actor_id"`
	Action       string `json:"action"`
	IPAddress    string `json:"ip_address"`
	Timestamp    string `json:"timestamp"`
	Metadata     string `json:"metadata"`
}
