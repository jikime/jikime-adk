'use client'

import { useState } from 'react'
import dynamic from 'next/dynamic'
import { Bot, PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import Sidebar from '@/components/sidebar/Sidebar'
import ChatInterface from '@/components/chat/ChatInterface'

const ShellPanel = dynamic(() => import('@/components/shell/ShellPanel'), { ssr: false })
const FileTree = dynamic(() => import('@/components/file-tree/FileTree'), { ssr: false })
const GitPanel = dynamic(() => import('@/components/git-panel/GitPanel'), { ssr: false })

type Tab = 'chat' | 'terminal' | 'files' | 'git'

const TABS: { id: Tab; label: string }[] = [
  { id: 'chat',     label: '채팅' },
  { id: 'terminal', label: '터미널' },
  { id: 'files',    label: '파일' },
  { id: 'git',      label: 'Git' },
]

export default function AppLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [activeTab, setActiveTab] = useState<Tab>('chat')
  // 한 번이라도 활성화된 탭만 마운트 유지 (lazy mount)
  const [mountedTabs, setMountedTabs] = useState<Set<Tab>>(new Set(['chat']))

  const handleTabChange = (tab: Tab) => {
    setActiveTab(tab)
    setMountedTabs(prev => prev.has(tab) ? prev : new Set([...prev, tab]))
  }

  return (
    <div className="flex flex-col h-screen bg-zinc-950 text-zinc-100 overflow-hidden">
      {/* Top bar */}
      <header className="flex items-center gap-2 px-3 py-2 bg-zinc-900 border-b border-zinc-800 shrink-0 z-10">
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7 text-zinc-400 hover:text-zinc-100"
          onClick={() => setSidebarOpen(v => !v)}
        >
          {sidebarOpen ? <PanelLeftClose className="w-4 h-4" /> : <PanelLeftOpen className="w-4 h-4" />}
        </Button>
        <div className="flex items-center gap-1.5">
          <Bot className="w-4 h-4 text-blue-400" />
          <span className="text-sm font-bold text-blue-400">JiKiME</span>
          <span className="text-xs text-zinc-500">Claude Code</span>
        </div>

        {/* Tab buttons in header */}
        <div className="flex items-center gap-0.5 ml-4">
          {TABS.map(tab => (
            <button
              key={tab.id}
              onClick={() => handleTabChange(tab.id)}
              className={cn(
                'px-3 py-1 rounded-md text-xs font-medium transition-colors',
                activeTab === tab.id
                  ? 'bg-zinc-700 text-white'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800'
              )}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </header>

      {/* Body */}
      <div className="flex flex-1 min-h-0 overflow-hidden">
        {/* Sidebar — CSS 숨김으로 unmount 방지 */}
        <aside
          className={cn(
            'shrink-0 transition-all duration-200 overflow-hidden border-r border-zinc-800',
            sidebarOpen ? 'w-64' : 'w-0'
          )}
        >
          <div className={cn('h-full', sidebarOpen ? 'block' : 'hidden')}>
            <Sidebar />
          </div>
        </aside>

        {/* Main content
              - 첫 방문 전: 마운트하지 않음 (API 호출 없음)
              - 첫 방문 후: 항상 마운트 유지 + visibility 전환
                visibility:hidden → 레이아웃 크기 보존 (xterm.js fit 정상 동작) */}
        <main className="flex-1 min-w-0 overflow-hidden relative">
          <div className={cn('absolute inset-0 p-2 transition-none',
            activeTab === 'chat' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
            <ChatInterface />
          </div>

          {mountedTabs.has('terminal') && (
            <div className={cn('absolute inset-0 p-2 transition-none',
              activeTab === 'terminal' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
              <ShellPanel />
            </div>
          )}

          {mountedTabs.has('files') && (
            <div className={cn('absolute inset-0 p-2 transition-none',
              activeTab === 'files' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
              <FileTree />
            </div>
          )}

          {mountedTabs.has('git') && (
            <div className={cn('absolute inset-0 p-2 transition-none',
              activeTab === 'git' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
              <GitPanel />
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
