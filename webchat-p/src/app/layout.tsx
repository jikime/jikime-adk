import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'JiKiME Web Chat',
  description: 'Claude Code web chat via claude -p',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ko" className="dark">
      <body className="antialiased">{children}</body>
    </html>
  )
}
