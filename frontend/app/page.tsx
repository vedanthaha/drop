'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Header from '@/components/Header'
import Hero from '@/components/Hero'
import UrlInput from '@/components/UrlInput'
import MediaPreview from '@/components/MediaPreview'
import Footer from '@/components/Footer'
import { SiYoutube, SiPinterest, SiInstagram, SiTiktok, SiX } from 'react-icons/si'

export default function Home() {
  const [status, setStatus] = useState<'idle' | 'detecting' | 'resolved' | 'error'>('idle')
  const [errorMessage, setErrorMessage] = useState('')
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [mediaData, setMediaData] = useState<any>(null)
  
  const handleUrlSubmit = async (submittedUrl: string) => {
    setStatus('detecting')
    setErrorMessage('')
    try {
      const res = await fetch('/api/resolve', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: submittedUrl })
      })
      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.details || data.error || 'Failed to resolve')
      }
      setMediaData(data)
      setStatus('resolved')
    } catch (e) {
      console.error(e)
      setErrorMessage(e instanceof Error ? e.message : 'Failed to resolve')
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
              <div className="mt-8 flex gap-6 justify-center items-center text-gray-400">
                <SiYoutube size={20} className="hover:text-[#FF0000] transition-colors" />
                <SiPinterest size={20} className="hover:text-[#E60023] transition-colors" />
                <SiInstagram size={20} className="hover:text-[#E1306C] transition-colors" />
                <SiTiktok size={20} className="hover:text-black dark:hover:text-white transition-colors" />
                <SiX size={18} className="hover:text-black dark:hover:text-white transition-colors" />
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
              <div>{errorMessage || 'Something went wrong.'}</div>
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
