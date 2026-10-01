ALTER TABLE download_jobs ADD COLUMN IF NOT EXISTS resolved_media_id UUID REFERENCES resolved_media(id);
