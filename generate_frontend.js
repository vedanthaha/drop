const fs = require('fs');
const path = require('path');

const files = {
"frontend/app/layout.tsx": `
import './globals.css'
import type { Metadata } from 'next'
import { Instrument_Serif, DM_Sans, IBM_Plex_Mono } from 'next/font/google'

const instrument = Instrument_Serif({ subsets: ['latin'], weight: '400', variable: '--font-instrument' })
const dmSans = DM_Sans({ subsets: ['latin'], variable: '--font-dm-sans' })
const plexMono = IBM_Plex_Mono({ subsets: ['latin'], weight: '400', variable: '--font-plex-mono' })

export const metadata: Metadata = {
  title: 'drop. | Download anything',
  description: 'A beautifully designed internet utility to download media.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={\`\${instrument.variable} \${dmSans.variable} \${plexMono.variable} font-sans bg-[#F9F9F8] text-[#111111] antialiased\`}>
        {children}
      </body>
    </html>
  )
}
`,

"frontend/app/globals.css": `
@import "tailwindcss";

@theme {
  --font-sans: var(--font-dm-sans), ui-sans-serif, system-ui, sans-serif;
  --font-serif: var(--font-instrument), ui-serif, Georgia, serif;
  --font-mono: var(--font-plex-mono), ui-monospace, SFMono-Regular, monospace;
}

body {
  background-color: #F9F9F8;
  color: #111111;
}

input:focus {
  outline: none;
  box-shadow: none;
}
`,

"frontend/app/page.tsx": `
'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Header from '@/components/Header'
import Hero from '@/components/Hero'
import UrlInput from '@/components/UrlInput'
import MediaPreview from '@/components/MediaPreview'
import Footer from '@/components/Footer'

export default function Home() {
  const [url, setUrl] = useState('')
  const [status, setStatus] = useState<'idle' | 'detecting' | 'resolved' | 'error'>('idle')
  const [mediaData, setMediaData] = useState<any>(null)
  
  const handleUrlSubmit = async (submittedUrl: string) => {
    setUrl(submittedUrl)
    setStatus('detecting')
    try {
      const res = await fetch('/api/resolve', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: submittedUrl })
      })
      if (!res.ok) throw new Error('Failed to resolve')
      const data = await res.json()
      setMediaData(data)
      setStatus('resolved')
    } catch (e) {
      console.error(e)
      setStatus('error')
    }
  }

  return (
    <div className="min-h-screen flex flex-col max-w-4xl mx-auto px-6 py-8">
      <Header />
      
      <main className="flex-1 flex flex-col items-center justify-center text-center mt-20 mb-20 w-full">
        <AnimatePresence mode="wait">
          {status === 'idle' && (
            <motion.div
              key="hero"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.3 }}
              className="w-full"
            >
              <Hero />
              <div className="mt-12 w-full max-w-lg mx-auto">
                <UrlInput onSubmit={handleUrlSubmit} />
              </div>
              <div className="mt-8 text-sm text-gray-400 font-mono tracking-tight">
                YouTube · Pinterest · Instagram · TikTok · X
              </div>
            </motion.div>
          )}

          {status === 'detecting' && (
            <motion.div
              key="detecting"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="font-sans text-lg text-gray-500"
            >
              Detecting...
            </motion.div>
          )}
          
          {status === 'error' && (
            <motion.div
              key="error"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="font-sans text-lg text-red-500"
            >
              Something went wrong.
              <button className="block mx-auto mt-4 underline text-black" onClick={() => setStatus('idle')}>Try again</button>
            </motion.div>
          )}

          {status === 'resolved' && mediaData && (
            <motion.div
              key="resolved"
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
              className="w-full max-w-lg mx-auto"
            >
              <div className="text-sm font-mono text-gray-500 mb-6 uppercase tracking-wider">
                {mediaData.platform}
              </div>
              <MediaPreview data={mediaData} onReset={() => setStatus('idle')} />
            </motion.div>
          )}
        </AnimatePresence>
      </main>
      
      <Footer />
    </div>
  )
}
`,

"frontend/components/Header.tsx": `
export default function Header() {
  return (
    <header className="flex justify-between items-center w-full">
      <div className="font-serif text-2xl lowercase tracking-tight">drop.</div>
      <div className="text-sm font-mono text-gray-400">Supported platforms</div>
    </header>
  )
}
`,

"frontend/components/Hero.tsx": `
export default function Hero() {
  return (
    <div>
      <h1 className="font-serif text-6xl md:text-8xl tracking-tight leading-none mb-6">
        Download anything.
      </h1>
      <p className="font-sans text-xl md:text-2xl text-gray-500 font-light">
        Paste a link. We'll handle the rest.
      </p>
    </div>
  )
}
`,

"frontend/components/UrlInput.tsx": `
'use client'
import { useState } from 'react'
import { ArrowRight, Clipboard } from 'lucide-react'

export default function UrlInput({ onSubmit }: { onSubmit: (url: string) => void }) {
  const [val, setVal] = useState('')

  const handlePaste = async () => {
    try {
      const text = await navigator.clipboard.readText()
      setVal(text)
      onSubmit(text)
    } catch (e) {
      console.error(e)
    }
  }

  return (
    <div className="relative flex items-center border border-gray-200 rounded-full bg-white shadow-sm overflow-hidden p-2 transition-all hover:border-gray-300 focus-within:border-black focus-within:ring-1 focus-within:ring-black">
      <input 
        type="url"
        className="flex-1 bg-transparent px-4 py-2 text-lg w-full font-sans placeholder-gray-300"
        placeholder="Paste a link"
        value={val}
        onChange={e => setVal(e.target.value)}
        onKeyDown={e => e.key === 'Enter' && val && onSubmit(val)}
      />
      {val ? (
        <button onClick={() => onSubmit(val)} className="p-3 bg-black text-white rounded-full ml-2">
          <ArrowRight size={18} />
        </button>
      ) : (
        <button onClick={handlePaste} className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-500 hover:text-black transition-colors rounded-full">
          <Clipboard size={16} /> Paste
        </button>
      )}
    </div>
  )
}
`,

"frontend/components/MediaPreview.tsx": `
'use client'
import { useState, useEffect } from 'react'

export default function MediaPreview({ data, onReset }: { data: any, onReset: () => void }) {
  const [selectedFormat, setSelectedFormat] = useState(data.formats[0]?.id)
  const [downloadState, setDownloadState] = useState<'idle' | 'queued' | 'processing' | 'ready'>('idle')
  const [jobId, setJobId] = useState<string | null>(null)
  const [fileUrl, setFileUrl] = useState<string | null>(null)

  const handleDownload = async () => {
    setDownloadState('queued')
    try {
      const res = await fetch('/api/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          url: data.sourceUrl,
          platform: data.platform,
          formatId: selectedFormat,
          output: 'mp4'
        })
      })
      const { jobId } = await res.json()
      setJobId(jobId)
    } catch (e) {
      console.error(e)
    }
  }

  useEffect(() => {
    if (!jobId || downloadState === 'ready') return
    const interval = setInterval(async () => {
      const res = await fetch(\`/api/jobs/\${jobId}\`)
      const status = await res.json()
      if (status.status === 'ready') {
        setDownloadState('ready')
        const fileRes = await fetch(\`/api/files/\${jobId}\`)
        const { url } = await fileRes.json()
        setFileUrl(url)
        clearInterval(interval)
      } else {
        setDownloadState('processing')
      }
    }, 2000)
    return () => clearInterval(interval)
  }, [jobId, downloadState])

  return (
    <div className="bg-white border border-gray-100 p-4 md:p-8 rounded-3xl shadow-sm text-left w-full">
      <div className="aspect-video bg-gray-100 rounded-xl mb-6 overflow-hidden relative">
         {/* Placeholder for thumbnail, normally an img tag */}
         {data.thumbnail ? <img src={data.thumbnail} className="w-full h-full object-cover" /> : <div className="w-full h-full bg-gray-200"></div>}
      </div>
      
      <h2 className="font-serif text-3xl leading-tight mb-6">{data.title || 'Untitled Media'}</h2>
      
      <div className="flex flex-wrap gap-2 mb-8">
        {data.formats.map((f: any) => (
          <button 
            key={f.id}
            onClick={() => setSelectedFormat(f.id)}
            className={\`px-4 py-2 rounded-full font-mono text-sm transition-colors \${selectedFormat === f.id ? 'bg-black text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}\`}
          >
            {f.label}
          </button>
        ))}
      </div>

      <div className="flex items-center justify-between mt-auto">
        {downloadState === 'idle' ? (
           <button onClick={handleDownload} className="w-full py-4 bg-black text-white font-medium rounded-full hover:bg-gray-800 transition-colors">
             Download →
           </button>
        ) : downloadState === 'ready' && fileUrl ? (
           <div className="w-full flex flex-col gap-4">
              <a href={fileUrl} className="w-full block text-center py-4 bg-black text-white font-medium rounded-full hover:bg-gray-800 transition-colors">
                Save File
              </a>
              <button onClick={onReset} className="text-sm font-mono text-gray-500 underline text-center">Download another</button>
           </div>
        ) : (
           <div className="w-full py-4 text-center border border-gray-200 text-gray-600 font-mono text-sm rounded-full bg-gray-50">
             {downloadState === 'queued' ? 'Preparing...' : 'Processing...'}
           </div>
        )}
      </div>
    </div>
  )
}
`,

"frontend/components/Footer.tsx": `
export default function Footer() {
  return (
    <footer className="w-full border-t border-gray-200 pt-6 flex justify-between items-center text-xs font-mono text-gray-400">
      <div>Built for the web.</div>
      <div className="flex gap-4">
        <a href="#">Privacy</a>
        <a href="#">Terms</a>
      </div>
    </footer>
  )
}
`,

"frontend/app/api/resolve/route.ts": `
import { NextResponse } from 'next/server'

export async function POST(request: Request) {
  const { url } = await request.json()
  
  // Real implementation: proxy to worker or parse URL locally
  // For MVP: Detect platform from string
  let platform = 'unknown'
  if (url.includes('youtube') || url.includes('youtu.be')) platform = 'youtube'
  else if (url.includes('pinterest') || url.includes('pin.it')) platform = 'pinterest'
  else if (url.includes('instagram')) platform = 'instagram'
  else if (url.includes('tiktok')) platform = 'tiktok'
  else if (url.includes('x.com') || url.includes('twitter')) platform = 'x'

  return NextResponse.json({
    id: 'test-id',
    sourceUrl: url,
    platform,
    type: 'video',
    title: 'Sample ' + platform + ' content',
    formats: [
      { id: '1080p', label: '1080p' },
      { id: '720p', label: '720p' },
      { id: '480p', label: '480p' }
    ]
  })
}
`,

"frontend/app/api/download/route.ts": `
import { NextResponse } from 'next/server'
import { createClient } from '@supabase/supabase-js'

export async function POST(request: Request) {
  const { url, platform, formatId, output } = await request.json()
  
  // Real implementation: create job in supabase, notify worker
  // For MVP: We return a fake job ID, worker processing is simulated or configured to real worker URL

  const jobId = crypto.randomUUID()
  
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
  const supabaseKey = process.env.SUPABASE_SERVICE_ROLE_KEY!
  
  if (supabaseUrl && supabaseKey) {
    const supabase = createClient(supabaseUrl, supabaseKey)
    await supabase.from('download_jobs').insert({
      id: jobId,
      source_url: url,
      platform,
      media_type: 'video',
      status: 'queued'
    })
    
    const workerUrl = process.env.WORKER_URL || 'http://localhost:8080'
    fetch(workerUrl + '/process', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ jobId, sourceUrl: url, platform, formatId, output })
    }).catch(console.error)
  }
  
  return NextResponse.json({ jobId })
}
`,

"frontend/app/api/jobs/[jobId]/route.ts": `
import { NextResponse } from 'next/server'
import { createClient } from '@supabase/supabase-js'

export async function GET(request: Request, { params }: { params: { jobId: string } }) {
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
  const supabaseKey = process.env.SUPABASE_SERVICE_ROLE_KEY!
  
  if (supabaseUrl && supabaseKey) {
    const supabase = createClient(supabaseUrl, supabaseKey)
    const { data } = await supabase.from('download_jobs').select('status, stage, progress').eq('id', params.jobId).single()
    if (data) return NextResponse.json(data)
  }
  
  // Fake fallback
  return NextResponse.json({ status: 'ready', stage: 'ready' })
}
`,

"frontend/app/api/files/[jobId]/route.ts": `
import { NextResponse } from 'next/server'
import { createClient } from '@supabase/supabase-js'

export async function GET(request: Request, { params }: { params: { jobId: string } }) {
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
  const supabaseKey = process.env.SUPABASE_SERVICE_ROLE_KEY!
  
  if (supabaseUrl && supabaseKey) {
    const supabase = createClient(supabaseUrl, supabaseKey)
    // Create signed url for media bucket
    const { data } = await supabase.storage.from('media').createSignedUrl(\`downloads/\${params.jobId}/video.mp4\`, 3600)
    if (data) return NextResponse.json({ url: data.signedUrl })
  }

  return NextResponse.json({ url: 'https://example.com/fake-download' })
}
`
};

for (const [filepath, content] of Object.entries(files)) {
  const fullPath = path.join(__dirname, filepath);
  fs.mkdirSync(path.dirname(fullPath), { recursive: true });
  fs.writeFileSync(fullPath, content.trim() + '\n');
}
console.log('Frontend files generated successfully.');
