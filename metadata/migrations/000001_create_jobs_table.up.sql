CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    job_name TEXT NOT NULL,
    command TEXT NOT NULL, 
    status INTEGER NOT NULL, 
    exit_code INTEGER DEFAULT 0,
    pid INTEGER NOT NULL DEFAULT 0,
    metadata TEXT,
    created_at DATETIME NOT NULL DEFAULT current_timestamp,
    updated_at DATETIME NOT NULL DEFAULT current_timestamp
);

CREATE INDEX IF NOT EXISTS idx_jobs_timestamp ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_name ON jobs(job_name);

CREATE TRIGGER IF NOT EXISTS jobs_update_updated_at
    AFTER UPDATE ON jobs
    FOR EACH ROW
BEGIN
    UPDATE jobs
    SET updated_at = current_timestamp
    WHERE id = OLD.id;
END;
