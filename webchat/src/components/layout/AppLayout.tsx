'use client'

import React, { useState, useEffect } from 'react'
import dynamic from 'next/dynamic'
import { Bot, PanelLeftClose, PanelLeftOpen, Sun, Moon, ChevronDown, Check, MessageSquare, SquareTerminal, FolderOpen, GitBranch } from 'lucide-react'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import Sidebar from '@/components/sidebar/Sidebar'
import ChatInterface from '@/components/chat/ChatInterface'
import { useLocale } from '@/contexts/LocaleContext'
import { LOCALES } from '@/i18n'

const ShellPanel = dynamic(() => import('@/components/shell/ShellPanel'), { ssr: false })
const FileTree = dynamic(() => import('@/components/file-tree/FileTree'), { ssr: false })
const GitPanel = dynamic(() => import('@/components/git-panel/GitPanel'), { ssr: false })

type Tab = 'chat' | 'terminal' | 'files' | 'git'

function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme()
  const { t } = useLocale()
  const [mounted, setMounted] = useState(false)

  useEffect(() => { setMounted(true) }, [])

  if (!mounted) return <div className="h-7 w-7" />

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-7 w-7 text-foreground/60 hover:text-foreground"
      onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
      title={resolvedTheme === 'dark' ? t.layout.theme.toLight : t.layout.theme.toDark}
    >
      {resolvedTheme === 'dark'
        ? <Sun className="w-4 h-4 text-amber-400" />
        : <Moon className="w-4 h-4 text-indigo-500" />
      }
    </Button>
  )
}

function LanguageSelector() {
  const { locale, setLocale } = useLocale()
  const [open, setOpen] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => { setMounted(true) }, [])

  if (!mounted) return <div className="h-7 w-16" />

  const current = LOCALES.find(l => l.id === locale)

  return (
    <div className="relative">
      <Button
        variant="ghost"
        size="sm"
        className="h-7 px-2 gap-1 text-muted-foreground hover:text-foreground"
        onClick={() => setOpen(v => !v)}
      >
        <span className="text-sm">{current?.flag}</span>
        <ChevronDown className={cn('w-3 h-3 transition-transform', open && 'rotate-180')} />
      </Button>

      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute right-0 top-full mt-1 bg-card border border-border rounded-lg shadow-lg z-50 py-1 w-36">
            {LOCALES.map(l => (
              <button
                key={l.id}
                onClick={() => { setLocale(l.id); setOpen(false) }}
                className={cn(
                  'flex items-center gap-2 w-full px-3 py-1.5 text-xs transition-colors',
                  locale === l.id
                    ? 'text-blue-400'
                    : 'text-foreground/80 hover:bg-muted hover:text-foreground'
                )}
              >
                <span>{l.flag}</span>
                <span>{l.label}</span>
                {locale === l.id && <Check className="w-3 h-3 ml-auto shrink-0" />}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

export default function AppLayout() {
  const { t } = useLocale()
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [activeTab, setActiveTab] = useState<Tab>('chat')
  const [mountedTabs, setMountedTabs] = useState<Set<Tab>>(new Set(['chat']))

  const TABS: { id: Tab; label: string; icon: React.ReactNode }[] = [
    { id: 'chat',     label: t.layout.tabs.chat,     icon: <MessageSquare  className="w-3.5 h-3.5 shrink-0 text-sky-400" /> },
    { id: 'terminal', label: t.layout.tabs.terminal, icon: <SquareTerminal className="w-3.5 h-3.5 shrink-0 text-emerald-400" /> },
    { id: 'files',    label: t.layout.tabs.files,    icon: <FolderOpen     className="w-3.5 h-3.5 shrink-0 text-amber-400" /> },
    { id: 'git',      label: t.layout.tabs.git,      icon: <GitBranch      className="w-3.5 h-3.5 shrink-0 text-purple-400" /> },
  ]

  const handleTabChange = (tab: Tab) => {
    setActiveTab(tab)
    setMountedTabs(prev => prev.has(tab) ? prev : new Set([...prev, tab]))
  }

  return (
    <div className="flex flex-col h-screen bg-background text-foreground overflow-hidden">
      {/* Top bar — z-20 so dropdown escapes above z-10 tab panels */}
      <header className="flex items-center bg-card border-b border-border shrink-0 z-20 relative">
        {/* Logo — same width as sidebar */}
        <div className={cn(
          'flex items-center gap-1.5 shrink-0 transition-all duration-200 overflow-hidden border-r border-border',
          sidebarOpen ? 'w-64 px-3 py-2' : 'w-0 px-0 py-2'
        )}>
          <Bot className="w-4 h-4 dark:text-blue-400 text-blue-600 shrink-0" />
          <span className="text-sm font-bold dark:text-blue-400 text-blue-600 whitespace-nowrap">JiKiME-ADK</span>
          <span className="text-[10px] text-muted-foreground whitespace-nowrap">Claude Code</span>
        </div>

        {/* Collapse icon + Tab buttons */}
        <div className="flex items-center gap-0.5 px-2 py-2 flex-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-foreground/60 hover:text-foreground"
            onClick={() => setSidebarOpen(v => !v)}
          >
            {sidebarOpen ? <PanelLeftClose className="w-4 h-4" /> : <PanelLeftOpen className="w-4 h-4" />}
          </Button>
          {TABS.map(tab => (
            <button
              key={tab.id}
              onClick={() => handleTabChange(tab.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-medium transition-colors',
                activeTab === tab.id
                  ? 'dark:bg-blue-500/25 dark:text-blue-200 dark:border-blue-500/30 bg-blue-600/15 text-blue-700 border-blue-600/40 border'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted border border-transparent'
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>

        {/* Language selector + Theme toggle — right side */}
        <div className="flex items-center gap-0.5 px-2 py-2 shrink-0">
          <LanguageSelector />
          <ThemeToggle />
        </div>
      </header>

      {/* Body */}
      <div className="flex flex-1 min-h-0 overflow-hidden">
        {/* Sidebar */}
        <aside
          className={cn(
            'shrink-0 transition-all duration-200 overflow-hidden border-r border-border',
            sidebarOpen ? 'w-64' : 'w-0'
          )}
        >
          <div className={cn('h-full', sidebarOpen ? 'block' : 'hidden')}>
            <Sidebar />
          </div>
        </aside>

        {/* Main content */}
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
