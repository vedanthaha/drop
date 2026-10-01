-- Core downloader state. Run this migration with Supabase service-role/admin access.
CREATE TABLE IF NOT EXISTS resolved_media (
    id UUID PRIMARY KEY,
    source_url TEXT NOT NULL,
    platform TEXT NOT NULL,
    media_type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT 'media',
    thumbnail_url TEXT NULL,
    width INTEGER NULL,
    height INTEGER NULL,
    duration INTEGER NULL,
    formats JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS download_jobs (
    id UUID PRIMARY KEY,
    resolved_media_id UUID NULL REFERENCES resolved_media(id) ON DELETE SET NULL,
    access_token_hash TEXT NULL,
    platform TEXT NOT NULL,
    source_url TEXT NOT NULL,
    media_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    stage TEXT NULL,
    title TEXT NULL,
    thumbnail_url TEXT NULL,
    storage_path TEXT NULL,
    filename TEXT NULL,
    filesize BIGINT NULL,
    selected_format TEXT NULL,
    output_format TEXT NULL,
    progress INTEGER NULL,
    error_code TEXT NULL,
    error_message TEXT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    locked_at TIMESTAMPTZ NULL,
    locked_by TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_resolved_media_expires_at ON resolved_media (expires_at);
CREATE INDEX IF NOT EXISTS idx_download_jobs_status ON download_jobs (status);
CREATE INDEX IF NOT EXISTS idx_download_jobs_created_at ON download_jobs (created_at);
CREATE INDEX IF NOT EXISTS idx_download_jobs_expires_at ON download_jobs (expires_at);
CREATE INDEX IF NOT EXISTS idx_download_jobs_locked_at ON download_jobs (locked_at);

INSERT INTO storage.buckets (id, name, public)
VALUES ('media', 'media', false)
ON CONFLICT (id) DO UPDATE SET public = false;

ALTER TABLE resolved_media ENABLE ROW LEVEL SECURITY;
ALTER TABLE download_jobs ENABLE ROW LEVEL SECURITY;

-- There are intentionally no anonymous/authenticated policies. Server routes and
-- the worker use the service-role key, which bypasses RLS.
