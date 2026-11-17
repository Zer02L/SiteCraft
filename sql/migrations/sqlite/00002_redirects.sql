-- +goose Up
CREATE TABLE IF NOT EXISTS redirects (
    id TEXT PRIMARY KEY,
    source_path TEXT NOT NULL UNIQUE,
    target_path TEXT NOT NULL,
    status_code INTEGER NOT NULL DEFAULT 301,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);
CREATE INDEX IF NOT EXISTS idx_redirects_source ON redirects (source_path);

-- +goose Down
DROP TABLE IF EXISTS redirects;
