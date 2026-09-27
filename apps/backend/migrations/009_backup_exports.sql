-- 009 backup exports
CREATE TABLE backup_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status TEXT NOT NULL,
    file_name TEXT,
    storage_target TEXT,
    sha256 TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    CONSTRAINT chk_backup_exports_status CHECK (status IN ('started', 'success', 'failed'))
);
