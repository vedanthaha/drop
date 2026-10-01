import { NextResponse } from 'next/server'

export async function POST(request: Request) {
  const { url } = await request.json()

  if (!url || typeof url !== 'string') {
    return NextResponse.json({ error: 'URL is required' }, { status: 400 })
  }

  // Basic URL validation
  try {
    const parsed = new URL(url)
    if (!['http:', 'https:'].includes(parsed.protocol)) {
      return NextResponse.json({ error: 'Invalid URL' }, { status: 400 })
    }
  } catch {
    return NextResponse.json({ error: 'Invalid URL' }, { status: 400 })
  }

  const workerUrl = process.env.WORKER_URL || 'http://localhost:8081'
  const workerSecret = process.env.WORKER_SECRET || ''

  try {
    const res = await fetch(`${workerUrl}/resolve`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${workerSecret}`,
      },
      body: JSON.stringify({ url }),
    })

    if (!res.ok) {
      const err = await res.json().catch(() => ({}))
      return NextResponse.json(
        {
          error: err.error || 'Could not resolve this link',
          details: typeof err.details === 'string' ? err.details : undefined,
        },
        { status: res.status }
      )
    }

    const data = await res.json()
    return NextResponse.json(data)
  } catch (e) {
    console.error('Resolve error:', e)
    return NextResponse.json(
      { error: 'Worker unavailable. Try again.' },
      { status: 502 }
    )
  }
}
