'use client'
import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Music, Film } from 'lucide-react'

interface MediaFormat {
  id: string
  label: string
  width: number
  height: number
  ext: string
  hasAudio: boolean
  filesize: number
}

interface MediaData {
  id: string
  platform: string
  type: string
  title: string
  thumbnail: string
  sourceUrl: string
  width: number
  height: number
  duration: number
  formats: MediaFormat[]
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return ''
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatDuration(seconds: number): string {
  if (!seconds) return ''
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

export default function MediaPreview({ data, onReset }: { data: MediaData; onReset: () => void }) {
  const [selectedFormat, setSelectedFormat] = useState(data.formats[0]?.id || '')
  const [outputFormat, setOutputFormat] = useState<'mp4' | 'mp3' | 'jpg' | 'png' | 'webp'>(data.type === 'image' ? 'jpg' : 'mp4')
  const [downloadState, setDownloadState] = useState<'idle' | 'queued' | 'processing' | 'ready' | 'error'>('idle')
  const [jobId, setJobId] = useState<string | null>(null)
  const [fileUrl, setFileUrl] = useState<string | null>(null)
  const [errorMsg, setErrorMsg] = useState('')
  const [progress, setProgress] = useState(0)
  const [stage, setStage] = useState('')
  const [fileInfo, setFileInfo] = useState<{ filename?: string; filesize?: number }>({})

  const selectedFormatObj = data.formats.find(f => f.id === selectedFormat)

  const handleDownload = async () => {
    setDownloadState('queued')
    setErrorMsg('')
    try {
      const res = await fetch('/api/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          mediaId: data.id,
          formatId: selectedFormat,
          output: outputFormat,
        }),
      })
      const result = await res.json()
      if (result.error) {
        setDownloadState('error')
        setErrorMsg(result.error)
        return
      }
      setJobId(result.jobId)
    } catch {
      setDownloadState('error')
      setErrorMsg('Download failed. Try again.')
    }
  }

  useEffect(() => {
    if (!jobId || downloadState === 'ready' || downloadState === 'error') return
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/jobs/${jobId}`)
        const status = await res.json()

        if (status.status === 'ready') {
          setDownloadState('ready')
          setProgress(100)
          setFileInfo({ filename: status.filename, filesize: status.filesize })

          const fileRes = await fetch(`/api/files/${jobId}`)
          const fileData = await fileRes.json()
          if (fileData.url) {
            setFileUrl(fileData.url)
          }
          clearInterval(interval)
        } else if (status.status === 'error') {
          setDownloadState('error')
          setErrorMsg(status.error_message || 'Download failed. Try again.')
          clearInterval(interval)
        } else {
          setDownloadState('processing')
          setProgress(status.progress || 0)
          setStage(status.stage || 'processing')
        }
      } catch {
        // Network hiccup, keep polling
      }
    }, 2000)
    return () => clearInterval(interval)
  }, [jobId, downloadState])

  const stageLabel = (s: string) => {
    const labels: Record<string, string> = {
      queued: 'Preparing\u2026',
      extracting: 'Extracting\u2026',
      downloading: 'Downloading\u2026',
      processing: 'Processing\u2026',
      uploading: 'Uploading\u2026',
    }
    return labels[s] || 'Processing\u2026'
  }

  const downloadButtonText = () => {
    if (outputFormat === 'mp3') return 'Download MP3 →'
    if (selectedFormatObj) return `Download ${selectedFormatObj.label} →`
    return 'Download →'
  }

  return (
    <div className="bg-white dark:bg-[#222] border border-gray-100 dark:border-[#333] p-4 md:p-8 rounded-3xl shadow-sm text-left w-full">
      {/* Thumbnail */}
      {data.thumbnail && (
        <div className="aspect-video bg-gray-100 dark:bg-black rounded-xl mb-6 overflow-hidden relative">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={data.thumbnail} alt={data.title || 'Media thumbnail'} className="w-full h-full object-cover" />
        </div>
      )}

      {/* Title */}
      <h2 className="font-sans font-bold text-2xl leading-tight mb-2">{data.title || 'Untitled Media'}</h2>

      {/* Metadata */}
      <div className="flex gap-3 text-xs font-mono text-gray-400 mb-6">
        {data.duration > 0 && <span>{formatDuration(data.duration)}</span>}
        {data.width > 0 && data.height > 0 && <span>{data.width}&times;{data.height}</span>}
        <span className="uppercase">{data.platform}</span>
      </div>

      {/* Output format selector (Video / Audio) */}
      {data.type === 'video' && (
        <div className="flex gap-2 mb-4">
          <button
            onClick={() => setOutputFormat('mp4')}
            className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-colors ${outputFormat === 'mp4' ? 'bg-black text-white dark:bg-white dark:text-black' : 'bg-gray-100 dark:bg-[#333] text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-[#444]'}`}
          >
            <Film size={14} /> Video
          </button>
          <button
            onClick={() => setOutputFormat('mp3')}
            className={`flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-colors ${outputFormat === 'mp3' ? 'bg-black text-white dark:bg-white dark:text-black' : 'bg-gray-100 dark:bg-[#333] text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-[#444]'}`}
          >
            <Music size={14} /> Audio
          </button>
        </div>
      )}

      {/* Quality selector */}
      {(outputFormat === 'mp4' || data.type === 'image') && data.formats.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-6">
          {data.formats.map((f) => (
            <button
              key={f.id}
              onClick={() => setSelectedFormat(f.id)}
              className={`px-4 py-2 rounded-full font-mono text-sm transition-colors ${selectedFormat === f.id ? 'bg-black text-white dark:bg-white dark:text-black' : 'bg-gray-100 dark:bg-[#333] text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-[#444]'}`}
            >
              {f.label}
            </button>
          ))}
        </div>
      )}

      {/* Audio info */}
      {outputFormat === 'mp3' && (
        <div className="text-sm text-gray-500 dark:text-gray-400 mb-6 font-mono">
          Best available audio · MP3
        </div>
      )}

      {/* Selected format details */}
      {(outputFormat === 'mp4' || data.type === 'image') && selectedFormatObj && selectedFormatObj.filesize > 0 && (
        <div className="text-xs font-mono text-gray-400 mb-6">
          {selectedFormatObj.label} · {data.type === 'image' ? 'JPG' : 'MP4'} · {formatBytes(selectedFormatObj.filesize)}
        </div>
      )}

      {/* Action area */}
      <div className="flex items-center justify-between mt-auto">
        {downloadState === 'idle' && (
          <button
            onClick={handleDownload}
            className="w-full py-4 bg-black dark:bg-white text-white dark:text-black font-medium rounded-full hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
          >
            {downloadButtonText()}
          </button>
        )}

        {downloadState === 'error' && (
          <div className="w-full flex flex-col gap-3">
            <div className="text-center text-sm text-red-500">{errorMsg}</div>
            <button
              onClick={() => { setDownloadState('idle'); setErrorMsg(''); }}
              className="w-full py-4 bg-black dark:bg-white text-white dark:text-black font-medium rounded-full hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
            >
              Try again
            </button>
          </div>
        )}

        {(downloadState === 'queued' || downloadState === 'processing') && (
          <div className="w-full relative overflow-hidden rounded-full border border-gray-200 dark:border-[#444] bg-gray-50 dark:bg-[#333]">
            <motion.div 
              className="absolute top-0 left-0 h-full bg-black/10 dark:bg-white/20"
              initial={{ width: 0 }}
              animate={{ width: `${progress || 0}%` }}
              transition={{ duration: 0.5, ease: "easeOut" }}
            />
            <div className="w-full py-4 relative z-10 text-center text-gray-600 dark:text-gray-300 font-mono text-sm">
              {stageLabel(stage)}{progress > 0 && progress < 100 ? ` ${progress}%` : ''}
            </div>
          </div>
        )}

        {downloadState === 'ready' && (
          <motion.div
            initial={{ opacity: 0, y: 4 }}
            animate={{ opacity: 1, y: 0 }}
            className="w-full flex flex-col gap-4"
          >
            {fileInfo.filesize && fileInfo.filesize > 0 && (
              <div className="text-center text-xs font-mono text-gray-400">
                {outputFormat === 'mp3' ? 'MP3' : `${selectedFormatObj?.label || ''} · ${data.type === 'image' ? 'JPG' : 'MP4'}`} · {formatBytes(fileInfo.filesize)}
              </div>
            )}
            {fileUrl ? (
              <a
                href={fileUrl}
                download={fileInfo.filename}
                className="w-full block text-center py-4 bg-black dark:bg-white text-white dark:text-black font-medium rounded-full hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
              >
                Save File ↓
              </a>
            ) : (
              <div className="text-center text-sm text-gray-500">Generating link…</div>
            )}
            <button onClick={onReset} className="text-sm font-mono text-gray-500 dark:text-gray-400 underline text-center">
              Download another
            </button>
          </motion.div>
        )}
      </div>
    </div>
  )
}
