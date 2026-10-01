import { NextResponse } from 'next/server'
import { createClient } from '@supabase/supabase-js'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ jobId: string }> }
) {
  const { jobId } = await params
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL
  const supabaseKey = process.env.SUPABASE_SERVICE_ROLE_KEY

  if (!supabaseUrl || !supabaseKey) {
    return NextResponse.json({ error: 'Server misconfigured' }, { status: 500 })
  }

  const supabase = createClient(supabaseUrl, supabaseKey)

  // Get job to verify it's ready and not expired
  const { data: job, error } = await supabase
    .from('download_jobs')
    .select('status, storage_path, filename, expires_at')
    .eq('id', jobId)
    .single()

  if (error || !job) {
    return NextResponse.json({ error: 'Job not found' }, { status: 404 })
  }

  if (job.status !== 'ready') {
    return NextResponse.json({ error: 'File not ready' }, { status: 400 })
  }

  if (job.expires_at && new Date(job.expires_at) < new Date()) {
    return NextResponse.json({ error: 'Download expired' }, { status: 410 })
  }

  if (!job.storage_path) {
    return NextResponse.json({ error: 'File path missing' }, { status: 500 })
  }

  // Generate a signed URL (valid for 1 hour)
  const { data: signedData, error: signError } = await supabase
    .storage
    .from('media')
    .createSignedUrl(job.storage_path, 3600, {
      download: job.filename || 'download',
    })

  if (signError || !signedData) {
    console.error('Signed URL error:', signError)
    return NextResponse.json({ error: 'Could not generate download link' }, { status: 500 })
  }

  return NextResponse.json({ url: signedData.signedUrl, filename: job.filename })
}
