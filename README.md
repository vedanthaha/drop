# drop. - Universal Media Downloader

A beautifully designed internet utility to download media from anywhere. Built with a Next.js minimal editorial frontend and a lightweight Go/Docker backend for processing.

## Architecture

This project is separated into a frontend Next.js app and a backend Go worker.

- **FRONTEND + LIGHT API**
  - **Vercel**
  - Next.js + TypeScript
  - Tailwind CSS + Framer Motion
  - Responsibilities: URL validation, platform detection, resolve orchestration, job creation, job status API, signed URL generation.

- **DATABASE + STORAGE**
  - **Supabase**
  - PostgreSQL + Storage
  - Responsibilities: Job metadata, state, temporary storage, expiration.

- **MEDIA WORKER**
  - **Render Free Web Service**
  - Docker + Go
  - Tools inside: yt-dlp, FFmpeg
  - Responsibilities: Media extraction, downloading, FFmpeg processing, image processing, upscaling, uploading artifacts.

## Data Flow

1. User pastes URL in the Vercel/Next.js frontend.
2. Vercel validates URL and resolves media metadata (via yt-dlp on worker or local adapters).
3. Frontend shows actual options.
4. User selects format/quality and triggers download.
5. Vercel API creates a job in Supabase database (\`download_jobs\`).
6. Vercel triggers Render worker (\`/process\`).
7. Render downloads, processes via FFmpeg/image processing, and uploads file to Supabase Storage.
8. Render updates job state in Supabase.
9. Next.js API generates a short-lived signed URL for the completed job.
10. Browser downloads directly from Supabase Storage.

## Setup Instructions

### 1. Supabase Setup
- Create a Supabase project.
- Go to SQL Editor and run the migration in \`supabase/migrations/001_initial_schema.sql\`.
- This creates the \`download_jobs\` table and a private \`media\` bucket.

### 2. Go Worker (Render)
- Deploy the \`worker\` directory to Render as a Web Service (Docker runtime).
- Use the included \`Dockerfile\`.
- Set Environment Variables:
  - \`PORT\`: (Render provides this)
  - \`SUPABASE_URL\`: Your Supabase Project URL
  - \`SUPABASE_SERVICE_ROLE_KEY\`: Your Supabase service role key
  - \`WORKER_SECRET\`: A random string for API authentication

### 3. Frontend (Vercel)
- Deploy the \`frontend\` directory to Vercel.
- Set Environment Variables:
  - \`NEXT_PUBLIC_SUPABASE_URL\`: Your Supabase Project URL
  - \`NEXT_PUBLIC_SUPABASE_ANON_KEY\`: Your Supabase anon key
  - \`SUPABASE_SERVICE_ROLE_KEY\`: Your Supabase service role key (for API routes)
  - \`WORKER_URL\`: The public URL of your Render worker (e.g. \`https://worker.onrender.com\`)
  - \`WORKER_SECRET\`: The same random string used above.

### 4. Verification
- Test worker \`/health\` endpoint.
- Submit a YouTube URL and verify Vercel correctly resolves.
- Perform a test download and ensure Supabase Storage receives the file.
- Check that the signed URL download works successfully.

## Extending the platform

### Adding a new Platform
1. Create \`worker/platforms/newplatform.go\` implementing \`Adapter\` interface.
2. Register in \`worker/platforms/detector.go\`.
3. Update Vercel's \`/api/resolve\` route to map platform appropriately.

### Cleanup and Retention
- Free instances and storage limits require maintenance.
- Jobs and objects must have an \`expires_at\` property.
- Worker automatically purges its local temp files.
- Scheduled DB/bucket cleanups can be run via external cron if needed.
