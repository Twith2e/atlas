CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    public_id TEXT NOT NULL UNIQUE,
    resource_type TEXT NOT NULL INDEX CONSTRAINT valid_resource_type CHECK (resource_type IN ('')),
    resource_id TEXT NOT NULL INDEX,
    actor_type TEXT NOT NULL INDEX,
    actor_id TEXT NOT NULL INDEX,
    action TEXT NOT NULL,
    ip_address TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata jsonb
)