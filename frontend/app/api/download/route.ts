import { NextResponse } from 'next/server'
import { createClient } from '@supabase/supabase-js'

export async function POST(request: Request) {
  const { mediaId, formatId, output } = await request.json()

  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL
  const supabaseKey = process.env.SUPABASE_SERVICE_ROLE_KEY
  const workerUrl = process.env.WORKER_URL || 'http://localhost:8081'
  const workerSecret = process.env.WORKER_SECRET || ''

  if (!supabaseUrl || !supabaseKey) {
    return NextResponse.json({ error: 'Server misconfigured' }, { status: 500 })
  }

  const supabase = createClient(supabaseUrl, supabaseKey)

  // Validate the resolved media exists
  const { data: media, error: mediaError } = await supabase
    .from('resolved_media')
    .select('*')
    .eq('id', mediaId)
    .single()

  if (mediaError || !media) {
    return NextResponse.json({ error: 'Invalid or expired media' }, { status: 400 })
  }

  const jobId = crypto.randomUUID()

  // Create job in Supabase using trusted server data
  const { error: insertError } = await supabase.from('download_jobs').insert({
    id: jobId,
    source_url: media.source_url,
    platform: media.platform,
    media_type: media.media_type,
    title: media.title,
    resolved_media_id: mediaId,
    status: 'queued',
    stage: 'queued',
    selected_format: formatId,
    output_format: output || 'mp4',
    progress: 0,
  })

  if (insertError) {
    console.error('Job insert error:', insertError)
    return NextResponse.json({ error: 'Failed to create job' }, { status: 500 })
  }

  // Trigger worker (fire and forget)
  fetch(`${workerUrl}/process`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${workerSecret}`,
    },
    body: JSON.stringify({
      jobId,
    }),
  }).catch((e) => console.error('Worker trigger error:', e))

  return NextResponse.json({ jobId })
}
