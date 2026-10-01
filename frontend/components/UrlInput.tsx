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
    <div className="relative flex items-center border border-gray-200 dark:border-[#333] rounded-full bg-white dark:bg-[#222] shadow-sm overflow-hidden p-2 transition-all hover:border-gray-300 dark:hover:border-gray-500 focus-within:border-black dark:focus-within:border-white focus-within:ring-1 focus-within:ring-black dark:focus-within:ring-white">
      <input 
        type="url"
        className="flex-1 bg-transparent px-4 py-2 text-lg w-full font-sans placeholder-gray-400 dark:text-white"
        placeholder="Paste a link"
        value={val}
        onChange={e => setVal(e.target.value)}
        onKeyDown={e => e.key === 'Enter' && val && onSubmit(val)}
      />
      {val ? (
        <button onClick={() => onSubmit(val)} className="p-3 bg-black dark:bg-white text-white dark:text-black rounded-full ml-2">
          <ArrowRight size={18} />
        </button>
      ) : (
        <button onClick={handlePaste} className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-gray-500 hover:text-black dark:hover:text-white transition-colors rounded-full">
          <Clipboard size={16} /> Paste
        </button>
      )}
    </div>
  )
}
