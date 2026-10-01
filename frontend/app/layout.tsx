import './globals.css'
import type { Metadata } from 'next'
import { ThemeProvider } from '@/components/ThemeProvider'
import { Instrument_Serif, IBM_Plex_Mono } from 'next/font/google'
import localFont from 'next/font/local'

const typoGrotesk = localFont({
  src: [
    { path: './fonts/Typo Grotesk Thin Demo.otf', weight: '100', style: 'normal' },
    { path: './fonts/Typo Grotesk Thin Italic Demo.otf', weight: '100', style: 'italic' },
    { path: './fonts/Typo Grotesk Demo.otf', weight: '400', style: 'normal' },
    { path: './fonts/Typo Grotesk Italic Demo.otf', weight: '400', style: 'italic' },
    { path: './fonts/Typo Grotesk Bold Demo.otf', weight: '700', style: 'normal' },
    { path: './fonts/Typo Grotesk Bold Italic Demo.otf', weight: '700', style: 'italic' },
    { path: './fonts/Typo Grotesk Black Demo.otf', weight: '900', style: 'normal' },
    { path: './fonts/Typo Grotesk Black Italic Demo.otf', weight: '900', style: 'italic' }
  ],
  variable: '--font-typo-grotesk'
})

const instrument = Instrument_Serif({ subsets: ['latin'], weight: '400', variable: '--font-instrument' })
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
    <html lang="en" suppressHydrationWarning>
      <body className={`${instrument.variable} ${typoGrotesk.variable} ${plexMono.variable} font-sans antialiased`}>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
          {children}
        </ThemeProvider>
      </body>
    </html>
  )
}
