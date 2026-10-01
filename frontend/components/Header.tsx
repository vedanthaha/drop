'use client'

import { useState, useEffect } from 'react'
import { useTheme } from 'next-themes'
import { Moon, Sun } from 'lucide-react'

export default function Header() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <header className="flex justify-between items-center w-full">
      <div className="flex items-center">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src="/drop%20black.png" alt="drop." className="dark:hidden h-10 w-auto object-contain" />
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src="/drop%20white.png" alt="drop." className="hidden dark:block h-10 w-auto object-contain" />
      </div>
      <button 
        onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
        className="p-2 rounded-full hover:bg-black/5 dark:hover:bg-white/10 transition-colors h-10 w-10 flex items-center justify-center"
      >
        {mounted ? (
          theme === 'dark' ? <Sun size={20} /> : <Moon size={20} />
        ) : (
          <div className="w-5 h-5" />
        )}
      </button>
    </header>
  )
}
