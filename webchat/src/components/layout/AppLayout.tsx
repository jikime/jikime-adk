'use client'

import React, { Component, ErrorInfo, ReactNode, useState, useEffect, useCallback, useMemo, memo } from 'react'

// ── Error Boundary ───────────────────────────────────────────────
class ErrorBoundary extends Component<{ children: ReactNode }, { hasError: boolean; error: Error | null }> {
  constructor(props: { children: ReactNode }) {
    super(props)
    this.state = { hasError: false, error: null }
  }
  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }
  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack)
  }
  render() {
    if (this.state.hasError) {
      return (
        <div className="h-full flex flex-col items-center justify-center gap-3 text-center p-8">
          <p className="text-base font-medium text-destructive">패널을 로드하는 중 오류가 발생했습니다.</p>
          <p className="text-base text-muted-foreground font-mono">{this.state.error?.message}</p>
          <button
            className="text-base text-blue-500 hover:underline"
            onClick={() => this.setState({ hasError: false, error: null })}
          >
            다시 시도
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
import dynamic from 'next/dynamic'
import { Bot, Menu, Sun, Moon, ChevronDown, Check, MessageSquare, SquareTerminal, FolderOpen, GitBranch, Settings, Zap, Users, Layers } from 'lucide-react'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import Sidebar, { SettingsModal } from '@/components/sidebar/Sidebar'
import ChatInterface from '@/components/chat/ChatInterface'
import { useLocale } from '@/contexts/LocaleContext'
import { LOCALES } from '@/i18n'
import { useServer } from '@/contexts/ServerContext'
import { useProject } from '@/contexts/ProjectContext'
import { ResizablePanelGroup, ResizablePanel, ResizableHandle, usePanelRef } from '@/components/ui/resizable'

const PanelSkeleton = () => (
  <div className="h-full flex items-center justify-center text-muted-foreground text-sm animate-pulse">
    Loading…
  </div>
)

const ShellPanel  = dynamic(() => import('@/components/shell/ShellPanel'),  { ssr: false, loading: PanelSkeleton })
const FileTree    = dynamic(() => import('@/components/file-tree/FileTree'), { ssr: false, loading: PanelSkeleton })
const GitPanel    = dynamic(() => import('@/components/git-panel/GitPanel'), { ssr: false, loading: PanelSkeleton })
const BoardPanel  = dynamic(() => import('@/components/team/BoardPanel'),   { ssr: false, loading: PanelSkeleton })

type Tab = 'chat' | 'terminal' | 'files' | 'git' | 'board'

const ThemeToggle = memo(function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme()
  const { t } = useLocale()
  const [mounted, setMounted] = useState(false)

  useEffect(() => { setMounted(true) }, [])

  if (!mounted) return <div className="h-7 w-7" />

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-7 w-7 text-white/70 hover:text-white hover:bg-transparent"
      onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
      title={resolvedTheme === 'dark' ? t.layout.theme.toLight : t.layout.theme.toDark}
    >
      {resolvedTheme === 'dark'
        ? <Sun className="w-4 h-4 text-amber-400" />
        : <Moon className="w-4 h-4 text-white/90" />
      }
    </Button>
  )
})

const LanguageSelector = memo(function LanguageSelector() {
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
        className="h-7 px-2 gap-1 text-white/70 hover:text-white hover:bg-transparent dark:text-muted-foreground dark:hover:text-foreground dark:hover:bg-transparent"
        onClick={() => setOpen(v => !v)}
      >
        <span className="text-lg">{current?.flag}</span>
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
                  'flex items-center gap-2 w-full px-3 py-1.5 text-base transition-colors',
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
})

export default function AppLayout() {
  const { t } = useLocale()
  const { getApiUrl } = useServer()
  const { activeProject, activeSessionId } = useProject()
  const sidebarPanelRef = usePanelRef()
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [sidebarPx, setSidebarPx] = useState(256)
  const [activeTab, setActiveTab] = useState<Tab>('chat')
  const [showSettings, setShowSettings] = useState(false)
  const [mountedTabs, setMountedTabs] = useState<Set<Tab>>(new Set(['chat']))
  const [harnessRunning, setHarnessRunning] = useState(false)
  const [harnessActiveCount, setHarnessActiveCount] = useState(0)

  const TABS = useMemo<{ id: Tab; label: string; icon: React.ReactNode }[]>(() => [
    { id: 'chat',     label: t.layout.tabs.chat,     icon: <MessageSquare  className="w-3.5 h-3.5 shrink-0 text-white/80" /> },
    { id: 'terminal', label: t.layout.tabs.terminal, icon: <SquareTerminal className="w-3.5 h-3.5 shrink-0 text-white/80" /> },
    { id: 'files',    label: t.layout.tabs.files,    icon: <FolderOpen     className="w-3.5 h-3.5 shrink-0 text-white/80" /> },
    { id: 'git',      label: t.layout.tabs.git,      icon: <GitBranch      className="w-3.5 h-3.5 shrink-0 text-white/80" /> },
    { id: 'board',    label: 'Team',                 icon: <Users          className="w-3.5 h-3.5 shrink-0 text-white/80" /> },
  ], [t.layout.tabs])

  const handleTabChange = useCallback((tab: Tab) => {
    setActiveTab(tab)
    setMountedTabs(prev => prev.has(tab) ? prev : new Set([...prev, tab]))
  }, [])

  // Harness 상태 폴링 (5초 간격) — 프로젝트 선택 시에만
  useEffect(() => {
    const projectPath = activeProject?.path
    if (!projectPath) { setHarnessRunning(false); setHarnessActiveCount(0); return }
    const check = async () => {
      try {
        const res = await fetch(getApiUrl(`/api/harness/status?projectPath=${encodeURIComponent(projectPath)}`))
        if (res.ok) {
          const data = await res.json() as { status: string; activeCount?: number }
          setHarnessRunning(data.status === 'running')
          setHarnessActiveCount(data.activeCount ?? 0)
        }
      } catch { /* ignore */ }
    }
    check()
    const id = setInterval(check, 5000)
    return () => clearInterval(id)
  }, [activeProject?.path, getApiUrl])

  // 사이드바 토글 — ResizablePanel collapse/expand 연동
  const handleToggleSidebar = useCallback(() => {
    if (sidebarOpen) {
      sidebarPanelRef.current?.collapse()
    } else {
      sidebarPanelRef.current?.expand()
    }
    setSidebarOpen(v => !v)
  }, [sidebarOpen, sidebarPanelRef])

  return (
    <div className="flex flex-col h-screen bg-muted dark:bg-muted text-foreground overflow-hidden">
      {/* Top bar — z-20 so dropdown escapes above z-10 tab panels */}
      <header className="flex items-center bg-primary border-b border-border shrink-0 z-20 relative">
        {/* Logo — same width as sidebar */}
        <div
          className="flex items-center gap-1.5 shrink-0 transition-all duration-200 overflow-hidden py-2"
          style={{ width: sidebarOpen ? sidebarPx : 0, paddingLeft: sidebarOpen ? 12 : 0, paddingRight: sidebarOpen ? 12 : 0 }}
        >
          <Bot className="w-4 h-4 text-white/90 shrink-0" />
          <span className="text-lg font-bold text-white whitespace-nowrap">JiKiME-ADK</span>
          <span className="text-[10px] text-white/60 whitespace-nowrap">Claude Code</span>
        </div>

        {/* Collapse icon + Tab buttons */}
        <div className="flex items-center gap-0.5 px-2 py-2 flex-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-white/70 hover:text-white hover:bg-transparent"
            onClick={handleToggleSidebar}
          >
            <Menu className="w-4 h-4" />
          </Button>
          {TABS.map(tab => (
            <button
              key={tab.id}
              onClick={() => handleTabChange(tab.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1 rounded-md text-base font-medium transition-colors',
                activeTab === tab.id
                  ? 'bg-white/20 text-white border border-white/30'
                  : 'text-white/70 hover:text-white hover:bg-white/15 border border-transparent'
              )}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>

        {/* 프로젝트 컨텍스트 배지 — 항상 동일한 위치에 표시 */}
        {activeProject && (
          <div
            className="flex items-center gap-1 px-2 py-0.5 rounded bg-white/10 border border-white/20 text-white/80 text-[11px] font-mono shrink-0 max-w-[240px] cursor-default"
            title={activeProject.path}
          >
            <Layers className="w-3 h-3 shrink-0 text-white/60" />
            <span className="truncate">{activeProject.name}</span>
          </div>
        )}

        {/* Harness 진행 상태 뱃지 */}
        {harnessRunning && (
          <div className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-500/20 border border-emerald-500/30 text-emerald-300 text-[10px] font-medium shrink-0">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse shrink-0" />
            <Zap className="w-2.5 h-2.5 shrink-0" />
            Harness
            {harnessActiveCount > 0 && (
              <span className="bg-emerald-500 text-white rounded-full w-3.5 h-3.5 flex items-center justify-center text-[9px] font-bold shrink-0">
                {harnessActiveCount}
              </span>
            )}
          </div>
        )}

        {/* Settings + Language selector + Theme toggle — right side */}
        <div className="flex items-center gap-0.5 px-2 py-2 shrink-0">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-white/70 hover:text-white hover:bg-transparent"
            onClick={() => setShowSettings(true)}
          >
            <Settings className="w-4 h-4" />
          </Button>
          <LanguageSelector />
          <ThemeToggle />
        </div>
      </header>
      <SettingsModal open={showSettings} onClose={() => setShowSettings(false)} />

      {/* Body */}
      <ResizablePanelGroup id="app-layout" orientation="horizontal" className="flex-1 min-h-0 bg-muted dark:bg-muted">
        {/* Sidebar */}
        <ResizablePanel
          id="sidebar-panel"
          panelRef={sidebarPanelRef}
          defaultSize="256px"
          minSize="160px"
          maxSize="480px"
          collapsible
          collapsedSize="0px"
          onResize={(size) => {
            setSidebarOpen(size.asPercentage > 0)
            if (size.inPixels > 0) setSidebarPx(size.inPixels)
          }}
          className="min-w-0 bg-white dark:bg-muted"
        >
          <div className="h-full p-2 pr-0">
            <div className="h-full rounded-lg overflow-hidden">
              <Sidebar />
            </div>
          </div>
        </ResizablePanel>

        <ResizableHandle id="main-handle" withHandle={true} />

        {/* Main content */}
        <ResizablePanel id="content-panel" minSize="320px" className="min-w-0 bg-white dark:bg-muted">
          <main className="h-full overflow-hidden relative">
            {/* 세션 미선택 시 통합 가이드 페이지 — 탭(z-10)보다 위, 헤더(z-20)보다 아래 */}
            {!activeSessionId && (
              <div className="absolute inset-0 z-[11] p-2">
                <div className="flex flex-col items-center justify-center h-full bg-muted dark:bg-background rounded-lg border border-zinc-300 dark:border-zinc-600 text-center gap-6 px-8">
                  <div className="w-20 h-20 rounded-full bg-blue-500/10 border border-blue-500/20 flex items-center justify-center">
                    <MessageSquare className="w-9 h-9 text-blue-400/70" />
                  </div>
                  <div className="space-y-2">
                    <p className="text-xl font-semibold text-foreground/80">{t.chat.selectSession}</p>
                    <p className="text-base text-muted-foreground max-w-xs leading-relaxed">{t.chat.selectSessionHint}</p>
                  </div>
                </div>
              </div>
            )}
            <div className={cn('absolute inset-0 p-2 transition-none bg-white dark:bg-muted',
              activeTab === 'chat' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
              <div className="h-full rounded-lg bg-muted dark:bg-background"><ErrorBoundary><ChatInterface /></ErrorBoundary></div>
            </div>

            {mountedTabs.has('terminal') && (
              <div className={cn('absolute inset-0 p-2 transition-none bg-white dark:bg-muted',
                activeTab === 'terminal' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
                <div className="h-full rounded-lg bg-muted dark:bg-background"><ErrorBoundary><ShellPanel /></ErrorBoundary></div>
              </div>
            )}

            {mountedTabs.has('files') && (
              <div className={cn('absolute inset-0 p-2 transition-none bg-white dark:bg-muted',
                activeTab === 'files' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
              <div className="h-full rounded-lg bg-muted dark:bg-background"><ErrorBoundary><FileTree /></ErrorBoundary></div>
            </div>
            )}

            {mountedTabs.has('git') && (
              <div className={cn('absolute inset-0 p-2 transition-none bg-white dark:bg-muted',
                activeTab === 'git' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
                <div className="h-full rounded-lg bg-muted dark:bg-background"><ErrorBoundary><GitPanel /></ErrorBoundary></div>
              </div>
            )}

            {mountedTabs.has('board') && (
              <div className={cn('absolute inset-0 p-2 transition-none bg-white dark:bg-muted',
                activeTab === 'board' ? 'visible pointer-events-auto z-10' : 'invisible pointer-events-none z-0')}>
                <div className="h-full rounded-lg bg-muted dark:bg-background overflow-hidden"><ErrorBoundary><BoardPanel /></ErrorBoundary></div>
              </div>
            )}
          </main>
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  )
}
