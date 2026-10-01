const { Client } = require('pg');
const fs = require('fs');
const path = require('path');

async function main() {
  const client = new Client({
    connectionString: 'postgresql://postgres:vedant233445566@db.jwnksdmbaexaicjxlxrv.supabase.co:5432/postgres',
    ssl: { rejectUnauthorized: false }
  });

  await client.connect();
  console.log('Connected to Supabase PostgreSQL');

  // Read and run migration (only the table + indexes part, bucket creation uses Storage API)
  const migrationSQL = `
    CREATE TABLE IF NOT EXISTS download_jobs (
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

    CREATE INDEX IF NOT EXISTS idx_download_jobs_status ON download_jobs (status);
    CREATE INDEX IF NOT EXISTS idx_download_jobs_created_at ON download_jobs (created_at);
    CREATE INDEX IF NOT EXISTS idx_download_jobs_expires_at ON download_jobs (expires_at);

    -- Enable RLS
    ALTER TABLE download_jobs ENABLE ROW LEVEL SECURITY;

    -- Allow service role full access
    CREATE POLICY "Service role full access" ON download_jobs
      FOR ALL
      USING (true)
      WITH CHECK (true);
  `;

  await client.query(migrationSQL);
  console.log('Migration complete: download_jobs table created with RLS enabled');

  await client.end();
}

main().catch(e => { console.error(e); process.exit(1); });
