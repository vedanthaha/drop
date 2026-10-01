-- Create download_jobs table
CREATE TABLE download_jobs (
    id UUID PRIMARY KEY,
    platform TEXT NOT NULL,
    source_url TEXT NOT NULL,
    media_type TEXT NOT NULL,
    status TEXT NOT NULL,
    stage TEXT NULL,
    title TEXT NULL,
    thumbnail_url TEXT NULL,
    storage_path TEXT NULL,
    filename TEXT NULL,
    filesize BIGINT NULL,
    selected_format TEXT NULL,
    output_format TEXT NULL,
    progress INTEGER DEFAULT 0,
    error_code TEXT NULL,
    error_message TEXT NULL,
    retry_count INTEGER DEFAULT 0,
    locked_at TIMESTAMPTZ NULL,
    locked_by TEXT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ NULL
);

-- Create indexes
CREATE INDEX idx_download_jobs_status ON download_jobs (status);
CREATE INDEX idx_download_jobs_created_at ON download_jobs (created_at);
CREATE INDEX idx_download_jobs_expires_at ON download_jobs (expires_at);

-- Create private media bucket
insert into storage.buckets (id, name, public)
values ('media', 'media', false);

-- Set bucket private configuration
update storage.buckets
set public = false
where id = 'media';
