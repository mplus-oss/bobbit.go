CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    job_name TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT current_timestamp,
    command TEXT NOT NULL, 
    status INTEGER NOT NULL, 
    exit_code INTEGER DEFAULT 0,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_jobs_timestamp ON jobs(timestamp);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_name ON jobs(job_name);
