-- Apply after 001_initial_schema.sql. This migration is safe for the existing MVP table.
ALTER TABLE download_jobs ADD COLUMN IF NOT EXISTS resolved_media_id UUID;
ALTER TABLE download_jobs ADD COLUMN IF NOT EXISTS access_token_hash TEXT;
ALTER TABLE download_jobs ALTER COLUMN progress DROP DEFAULT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'download_jobs_resolved_media_id_fkey'
    ) THEN
        ALTER TABLE download_jobs
            ADD CONSTRAINT download_jobs_resolved_media_id_fkey
            FOREIGN KEY (resolved_media_id) REFERENCES resolved_media(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_download_jobs_resolved_media_id ON download_jobs (resolved_media_id);
CREATE INDEX IF NOT EXISTS idx_download_jobs_access_token_hash ON download_jobs (access_token_hash);

ALTER TABLE resolved_media ENABLE ROW LEVEL SECURITY;
ALTER TABLE download_jobs ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Service role full access" ON download_jobs;

CREATE OR REPLACE FUNCTION set_download_jobs_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS download_jobs_updated_at ON download_jobs;
CREATE TRIGGER download_jobs_updated_at
BEFORE UPDATE ON download_jobs
FOR EACH ROW EXECUTE FUNCTION set_download_jobs_updated_at();

CREATE OR REPLACE FUNCTION claim_download_job(
    p_job_id UUID,
    p_worker_id TEXT,
    p_lease_seconds INTEGER DEFAULT 900
)
RETURNS SETOF download_jobs
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    RETURN QUERY
    UPDATE download_jobs
    SET status = 'claimed',
        stage = 'extracting',
        locked_at = now(),
        locked_by = p_worker_id,
        retry_count = retry_count + 1,
        updated_at = now()
    WHERE id = p_job_id
      AND retry_count < 3
      AND (
          (status = 'queued' AND locked_at IS NULL)
          OR (status IN ('claimed', 'processing', 'uploading')
              AND locked_at < now() - make_interval(secs => p_lease_seconds))
      )
    RETURNING download_jobs.*;
END;
$$;

CREATE OR REPLACE FUNCTION claim_next_download_job(
    p_worker_id TEXT,
    p_lease_seconds INTEGER DEFAULT 900
)
RETURNS SETOF download_jobs
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    RETURN QUERY
    WITH candidate AS (
        SELECT id
        FROM download_jobs
        WHERE retry_count < 3
          AND (
              (status = 'queued' AND locked_at IS NULL)
              OR (status IN ('claimed', 'processing', 'uploading')
                  AND locked_at < now() - make_interval(secs => p_lease_seconds))
          )
        ORDER BY created_at
        FOR UPDATE SKIP LOCKED
        LIMIT 1
    )
    UPDATE download_jobs AS jobs
    SET status = 'claimed',
        stage = 'extracting',
        locked_at = now(),
        locked_by = p_worker_id,
        retry_count = jobs.retry_count + 1,
        updated_at = now()
    FROM candidate
    WHERE jobs.id = candidate.id
    RETURNING jobs.*;
END;
$$;

REVOKE ALL ON TABLE resolved_media FROM anon, authenticated;
REVOKE ALL ON TABLE download_jobs FROM anon, authenticated;
