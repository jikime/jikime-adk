'use client'

import { useState } from 'react'
import {
  Folder, FolderOpen, MessageSquare, Plus, RefreshCw,
  ChevronDown, ChevronRight, Settings, X, Check,
  Server, Trash2, Edit2, Globe, AlertTriangle,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useProject, type Project } from '@/contexts/ProjectContext'
import { useServer, type RemoteServer } from '@/contexts/ServerContext'
import { useWebSocket } from '@/contexts/WebSocketContext'
import { useLocale } from '@/contexts/LocaleContext'


// ── 설정 타입 & 유틸 ──────────────────────────────────────────────
export const MODELS = [
  { id: 'claude-opus-4-6',            label: 'Opus 4.6',   descKey: 'opusDesc'    as const },
  { id: 'claude-sonnet-4-6',          label: 'Sonnet 4.6', descKey: 'sonnetDesc'  as const },
  { id: 'claude-opus-4-5',            label: 'Opus 4.5',   descKey: 'opus45Desc'  as const },
  { id: 'claude-sonnet-4-5',          label: 'Sonnet 4.5', descKey: 'sonnet45Desc' as const },
  { id: 'claude-haiku-4-5-20251001',  label: 'Haiku 4.5',  descKey: 'haikuDesc'   as const },
] as const

export type ModelId = typeof MODELS[number]['id']
export type PermissionMode = 'bypassPermissions' | 'default'

export interface AppSettings {
  model: ModelId
  permissionMode: PermissionMode
}

const SETTINGS_KEY = 'webchat_settings'

const DEFAULT_SETTINGS: AppSettings = {
  model: 'claude-sonnet-4-6',
  permissionMode: 'bypassPermissions',
}

export function loadSettings(): AppSettings {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (raw) return { ...DEFAULT_SETTINGS, ...JSON.parse(raw) }
  } catch { /* */ }
  return DEFAULT_SETTINGS
}

export function saveSettings(s: AppSettings) {
  try { localStorage.setItem(SETTINGS_KEY, JSON.stringify(s)) } catch { /* */ }
}

// ── 서버 추가/편집 폼 ─────────────────────────────────────────────
function ServerForm({
  initial, onSave, onCancel,
}: {
  initial?: Partial<RemoteServer>
  onSave: (s: Omit<RemoteServer, 'id'>) => void
  onCancel: () => void
}) {
  const { t } = useLocale()
  const [name,   setName]   = useState(initial?.name   ?? '')
  const [host,   setHost]   = useState(initial?.host   ?? '')
  const [secure, setSecure] = useState(initial?.secure ?? false)

  const valid = name.trim() && host.trim()

  // ws:// wss:// http:// https:// 프로토콜 prefix 자동 제거
  const handleHostChange = (raw: string) => {
    const stripped = raw.replace(/^(wss?|https?):\/\//i, '').replace(/\/$/, '')
    setHost(stripped)
  }

  return (
    <div className="space-y-2 p-3 bg-muted/60 rounded-lg border border-border">
      <input
        value={name} onChange={e => setName(e.target.value)}
        placeholder={t.sidebar.serverName}
        className="w-full bg-background border border-border rounded px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground/50 outline-none focus:border-blue-500/50"
      />
      <input
        value={host} onChange={e => handleHostChange(e.target.value)}
        placeholder={t.sidebar.serverHost}
        className="w-full bg-background border border-border rounded px-2 py-1.5 text-xs text-foreground placeholder:text-muted-foreground/50 outline-none focus:border-blue-500/50 font-mono"
      />
      <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer select-none">
        <input
          type="checkbox" checked={secure} onChange={e => setSecure(e.target.checked)}
          className="accent-blue-500"
        />
        {t.sidebar.secureConnection}
      </label>
      <div className="flex gap-2 pt-1">
        <button
          onClick={() => valid && onSave({ name: name.trim(), host: host.trim(), secure })}
          disabled={!valid}
          className={cn(
            'flex-1 py-1 rounded text-xs font-medium transition-colors',
            valid ? 'bg-blue-600 text-white hover:bg-blue-500' : 'bg-muted text-muted-foreground cursor-default'
          )}
        >
          {t.sidebar.save}
        </button>
        <button
          onClick={onCancel}
          className="flex-1 py-1 rounded text-xs font-medium bg-muted text-foreground/70 hover:bg-accent transition-colors"
        >
          {t.sidebar.cancel}
        </button>
      </div>
    </div>
  )
}

// ── 설정 모달 ─────────────────────────────────────────────────────
function SettingsModal({ onClose }: { onClose: () => void }) {
  const { t } = useLocale()
  const [settings, setSettings] = useState<AppSettings>(() => loadSettings())

  const update = (patch: Partial<AppSettings>) => {
    const next = { ...settings, ...patch }
    setSettings(next)
    saveSettings(next)
  }

  return (
    <div
      className="absolute inset-0 z-50 flex items-end bg-black/60 backdrop-blur-[1px]"
      onClick={onClose}
    >
      <div
        className="w-full bg-card border-t border-border rounded-t-xl p-4 space-y-5 max-h-[85vh] overflow-y-auto"
        onClick={e => e.stopPropagation()}
      >
        {/* 헤더 */}
        <div className="flex items-center justify-between">
          <span className="text-sm font-semibold text-foreground">{t.sidebar.settingsTitle}</span>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground transition-colors">
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* ── 기본 모델 ── */}
        <div className="space-y-2">
          <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">{t.sidebar.defaultModel}</p>
          <div className="flex flex-col gap-1">
            {MODELS.map(m => (
              <button
                key={m.id}
                onClick={() => update({ model: m.id })}
                className={cn(
                  'flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-left transition-colors',
                  settings.model === m.id
                    ? 'bg-blue-600/20 border border-blue-500/40 dark:text-blue-300 text-blue-700 font-medium'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground border border-transparent'
                )}
              >
                {settings.model === m.id
                  ? <Check className="w-3 h-3 dark:text-blue-400 text-blue-600 shrink-0" />
                  : <span className="w-3 h-3 shrink-0" />
                }
                {m.label}
              </button>
            ))}
          </div>
        </div>

        {/* ── 권한 모드 ── */}
        <div className="space-y-2">
          <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">{t.sidebar.permissionMode}</p>
          <div className="flex gap-2">
            <button
              onClick={() => update({ permissionMode: 'bypassPermissions' })}
              className={cn(
                'flex-1 px-3 py-2 rounded-lg text-xs font-medium transition-colors',
                settings.permissionMode === 'bypassPermissions'
                  ? 'bg-blue-600/20 border border-blue-500/40 dark:text-blue-300 text-blue-700'
                  : 'bg-muted border border-border text-muted-foreground hover:text-foreground'
              )}
            >
              {t.sidebar.autoAllow}
            </button>
            <button
              onClick={() => update({ permissionMode: 'default' })}
              className={cn(
                'flex-1 px-3 py-2 rounded-lg text-xs font-medium transition-colors',
                settings.permissionMode === 'default'
                  ? 'bg-blue-600/20 border border-blue-500/40 dark:text-blue-300 text-blue-700'
                  : 'bg-muted border border-border text-muted-foreground hover:text-foreground'
              )}
            >
              {t.sidebar.confirmEach}
            </button>
          </div>
          <p className="text-[10px] dark:text-muted-foreground/50 text-muted-foreground/70 leading-relaxed">
            {settings.permissionMode === 'bypassPermissions'
              ? t.sidebar.autoAllowDesc
              : t.sidebar.confirmEachDesc}
          </p>
        </div>

        {/* 현재 설정 요약 */}
        <div className="text-[10px] dark:text-muted-foreground/50 text-muted-foreground/70 border-t border-border pt-3">
          {t.sidebar.activeModel}: <span className="text-foreground/70">{MODELS.find(m => m.id === settings.model)?.label}</span>
        </div>
      </div>
    </div>
  )
}

// ── 프로젝트 삭제 확인 다이얼로그 ─────────────────────────────────
function DeleteConfirmDialog({
  project, onConfirm, onCancel,
}: {
  project: Project
  onConfirm: () => void
  onCancel: () => void
}) {
  const { t } = useLocale()
  const sessionCount = project.sessions.length
  return (
    <div
      className="absolute inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-[2px] p-4"
      onClick={onCancel}
    >
      <div
        className="w-full max-w-xs bg-card border border-border rounded-xl p-4 space-y-3 shadow-2xl"
        onClick={e => e.stopPropagation()}
      >
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-full bg-red-500/15 flex items-center justify-center shrink-0">
            <AlertTriangle className="w-4 h-4 text-red-400" />
          </div>
          <span className="text-sm font-semibold text-foreground">{t.sidebar.deleteProjectTitle}</span>
        </div>

        <div className="space-y-1.5">
          <p className="text-xs text-foreground/80">
            <span className="font-medium text-foreground">{project.name}</span> {t.sidebar.deleteProjectDesc}
          </p>
          {sessionCount > 0 && (
            <p className="text-xs text-amber-400">
              {t.sidebar.deleteSessionCount(sessionCount)}
            </p>
          )}
          <p className="text-[11px] text-muted-foreground/50">
            {t.sidebar.deleteProjectWarning}
          </p>
        </div>

        <div className="flex gap-2 pt-1">
          <button
            onClick={onCancel}
            className="flex-1 py-1.5 rounded-lg text-xs font-medium bg-muted text-foreground/70 hover:bg-accent transition-colors"
          >
            {t.sidebar.cancel}
          </button>
          <button
            onClick={onConfirm}
            className="flex-1 py-1.5 rounded-lg text-xs font-medium bg-red-600 text-white hover:bg-red-500 transition-colors"
          >
            {t.sidebar.delete}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── 메인 사이드바 ─────────────────────────────────────────────────
export default function Sidebar() {
  const { t } = useLocale()
  const { projects, activeProject, activeSessionId, setActiveProject, setActiveSessionId, refreshProjects } = useProject()
  const { servers, activeServer, setActiveServerId, addServer, updateServer, removeServer, getApiUrl } = useServer()
  const { isConnected } = useWebSocket()

  const [loading,          setLoading]         = useState(false)
  const [openProjects,     setOpenProjects]     = useState<Set<string>>(new Set())
  const [showSettings,     setShowSettings]     = useState(false)
  const [showServerPicker, setShowServerPicker] = useState(false)
  const [addingServer,     setAddingServer]     = useState(false)
  const [editingServerId,  setEditingServerId]  = useState<string | null>(null)
  const [deletingProject,  setDeletingProject]  = useState<Project | null>(null)
  const [deletingId,       setDeletingId]       = useState<string | null>(null)
  // 세션 삭제
  type SessionTarget = { project: Project; sessionId: string }
  const [deletingSession,    setDeletingSession]    = useState<SessionTarget | null>(null)
  const [deletingSessionId,  setDeletingSessionId]  = useState<string | null>(null)

  const handleRefresh = async () => {
    setLoading(true)
    await refreshProjects()
    setLoading(false)
  }

  const handleDeleteConfirm = async () => {
    if (!deletingProject) return
    const id = deletingProject.id
    setDeletingProject(null)
    setDeletingId(id)
    try {
      const res = await fetch(getApiUrl(`/api/ws/project?id=${encodeURIComponent(id)}`), { method: 'DELETE' })
      if (res.ok) {
        // 삭제된 프로젝트가 활성화 상태면 초기화
        if (activeProject?.id === id) {
          setActiveProject(null)
          setActiveSessionId(null)
        }
        await refreshProjects()
      }
    } catch { /* */ } finally {
      setDeletingId(null)
    }
  }

  const handleSessionDeleteConfirm = async () => {
    if (!deletingSession) return
    const { project, sessionId } = deletingSession
    setDeletingSession(null)
    setDeletingSessionId(sessionId)
    try {
      const res = await fetch(
        getApiUrl(`/api/ws/session?projectId=${encodeURIComponent(project.id)}&sessionId=${encodeURIComponent(sessionId)}`),
        { method: 'DELETE' },
      )
      if (res.ok) {
        if (activeSessionId === sessionId) setActiveSessionId(null)
        await refreshProjects()
      }
    } catch { /* */ } finally {
      setDeletingSessionId(null)
    }
  }

  const toggleProject = (id: string) => {
    setOpenProjects(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const selectProject = (p: Project) => {
    setActiveProject(p)
    setActiveSessionId(null)
    if (!openProjects.has(p.id)) toggleProject(p.id)
  }

  return (
    <div className="relative flex flex-col h-full bg-card">

      {/* ── 서버 선택기 ── */}
      <div className="shrink-0 border-b border-border">
        <button
          onClick={() => setShowServerPicker(v => !v)}
          className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-muted transition-colors"
        >
          <div className="relative shrink-0">
            <Server className="w-3.5 h-3.5 dark:text-blue-400 text-blue-600" />
            <span className={cn(
              'absolute -top-0.5 -right-0.5 w-1.5 h-1.5 rounded-full border border-card',
              isConnected ? 'bg-green-400' : 'bg-muted-foreground'
            )} />
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium text-foreground truncate">{activeServer?.name ?? t.sidebar.noServer}</p>
            <p className="text-[10px] font-mono truncate">
              <span className={isConnected ? 'dark:text-green-500 text-green-600' : 'text-muted-foreground/60'}>
                {isConnected ? t.sidebar.connected : t.sidebar.connecting}
              </span>
              <span className="dark:text-muted-foreground/30 text-muted-foreground/80 ml-1">
                {activeServer ? `${activeServer.secure ? 'wss://' : 'ws://'}${activeServer.host}` : ''}
              </span>
            </p>
          </div>
          <ChevronDown className={cn('w-3 h-3 text-muted-foreground shrink-0 transition-transform', showServerPicker && 'rotate-180')} />
        </button>

        {/* 서버 드롭다운 */}
        {showServerPicker && (
          <div className="border-t border-border">
            {/* 서버 목록 */}
            <div className="py-1">
              {servers.map(s => (
                <div key={s.id}>
                  {editingServerId === s.id ? (
                    <div className="px-2 py-1.5">
                      <ServerForm
                        initial={s}
                        onSave={(patch) => { updateServer(s.id, patch); setEditingServerId(null) }}
                        onCancel={() => setEditingServerId(null)}
                      />
                    </div>
                  ) : (
                    <div className={cn(
                      'flex items-center gap-1 px-2 py-1 group transition-colors',
                      activeServer?.id === s.id ? 'bg-accent/60' : 'hover:bg-muted'
                    )}>
                      {/* 서버 선택 버튼 */}
                      <button
                        onClick={() => { setActiveServerId(s.id); setShowServerPicker(false) }}
                        className="flex items-center gap-2 flex-1 min-w-0 text-left py-0.5"
                      >
                        <Globe className={cn('w-3 h-3 shrink-0', activeServer?.id === s.id ? 'dark:text-blue-400 text-blue-600' : 'text-muted-foreground')} />
                        <div className="min-w-0">
                          <p className={cn('text-xs truncate leading-tight font-medium', activeServer?.id === s.id ? 'text-foreground' : 'dark:text-muted-foreground text-foreground/70')}>
                            {s.name}
                          </p>
                          <p className="text-[10px] dark:text-muted-foreground/50 text-muted-foreground/60 font-mono truncate">
                            {s.secure ? 'wss://' : 'ws://'}{s.host}
                          </p>
                        </div>
                        {activeServer?.id === s.id && <Check className="w-3 h-3 dark:text-blue-400 text-blue-600 shrink-0 ml-auto" />}
                      </button>
                      {/* 편집/삭제 — 로컬 서버는 편집만 숨김 */}
                      {s.id !== '__local__' && (
                        <div className="flex gap-0.5 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => setEditingServerId(s.id)}
                            className="p-1 text-muted-foreground/50 hover:text-foreground transition-colors rounded"
                          >
                            <Edit2 className="w-3 h-3" />
                          </button>
                          <button
                            onClick={() => removeServer(s.id)}
                            className="p-1 text-muted-foreground/50 hover:text-red-400 transition-colors rounded"
                          >
                            <Trash2 className="w-3 h-3" />
                          </button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* 서버 추가 */}
            <div className="border-t border-border p-2">
              {addingServer ? (
                <ServerForm
                  onSave={(s) => { addServer(s); setAddingServer(false) }}
                  onCancel={() => setAddingServer(false)}
                />
              ) : (
                <button
                  onClick={() => setAddingServer(true)}
                  className="flex items-center gap-2 w-full px-2 py-1.5 rounded-md border border-dashed border-border text-xs text-muted-foreground hover:text-foreground hover:border-foreground/50 transition-colors"
                >
                  <Plus className="w-3 h-3" />
                  {t.sidebar.addServer}
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* ── 프로젝트 헤더 ── */}
      <div className="flex items-center justify-between px-3 py-2.5 border-b border-border shrink-0">
        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t.sidebar.projects}</span>
        <Button
          variant="ghost" size="icon"
          className="h-6 w-6 text-muted-foreground hover:text-foreground"
          onClick={handleRefresh}
          disabled={loading}
        >
          <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
        </Button>
      </div>

      {/* ── 프로젝트 목록 ── */}
      <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden
        [&::-webkit-scrollbar]:w-[2px]
        [&::-webkit-scrollbar]:h-[2px]
        [&::-webkit-scrollbar-track]:bg-transparent
        [&::-webkit-scrollbar-thumb]:bg-border
        [&::-webkit-scrollbar-thumb]:rounded-full
        hover:[&::-webkit-scrollbar-thumb]:bg-muted-foreground/40">
        <div className="py-1">
          {projects.length === 0 && (
            <div className="px-3 py-4 text-center">
              <p className="text-xs text-muted-foreground/50">{t.sidebar.noProjects}</p>
              <p className="text-xs text-muted-foreground/30 mt-1">{t.sidebar.noProjectsHint}</p>
            </div>
          )}

          {projects.map((project) => {
            const isOpen = openProjects.has(project.id)
            return (
              <div key={project.id}>
                <div className="flex items-center group">
                  <button
                    onClick={() => toggleProject(project.id)}
                    className="p-0.5 ml-1 text-muted-foreground hover:text-foreground/80"
                  >
                    {isOpen
                      ? <ChevronDown className="w-3 h-3" />
                      : <ChevronRight className="w-3 h-3" />}
                  </button>
                  <button
                    className={cn(
                      'flex flex-1 items-center gap-2 px-2 py-1.5 text-left rounded-md ml-1 text-sm transition-colors min-w-0',
                      activeProject?.id === project.id && activeSessionId === null
                        ? 'bg-accent text-foreground'
                        : 'text-foreground/80 hover:bg-muted'
                    )}
                    onClick={() => selectProject(project)}
                  >
                    {deletingId === project.id
                      ? <RefreshCw className="w-3.5 h-3.5 text-muted-foreground shrink-0 animate-spin" />
                      : isOpen
                        ? <FolderOpen className="w-3.5 h-3.5 text-blue-400 shrink-0" />
                        : <Folder className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                    }
                    <span className="truncate text-xs">{project.name}</span>
                  </button>
                  <button
                    onClick={e => { e.stopPropagation(); setDeletingProject(project) }}
                    disabled={deletingId === project.id}
                    className="mr-1.5 p-1 rounded opacity-0 group-hover:opacity-100 text-muted-foreground/50 hover:text-red-400 hover:bg-red-400/10 transition-all shrink-0"
                    title={t.sidebar.deleteProject}
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>

                {isOpen && (
                  <div className="ml-5">
                    <button
                      className="flex items-center gap-2 px-3 py-1 text-xs text-muted-foreground hover:text-foreground/80 w-full hover:bg-muted rounded-md"
                      onClick={() => { setActiveProject(project); setActiveSessionId(null) }}
                    >
                      <Plus className="w-3 h-3" />
                      {t.sidebar.newChat}
                    </button>

                    {project.sessions.slice(0, 10).map((sessionId) => (
                      <div key={sessionId} className="flex items-center group/session">
                        <button
                          className={cn(
                            'flex flex-1 items-center gap-2 pl-2 pr-1 py-1 text-xs rounded-md transition-colors min-w-0',
                            activeSessionId === sessionId
                              ? 'bg-accent text-foreground'
                              : 'text-muted-foreground hover:text-foreground/80 hover:bg-muted'
                          )}
                          onClick={() => { setActiveProject(project); setActiveSessionId(sessionId) }}
                        >
                          {deletingSessionId === sessionId
                            ? <RefreshCw className="w-3 h-3 shrink-0 animate-spin" />
                            : <MessageSquare className="w-3 h-3 shrink-0" />
                          }
                          <span className="truncate font-mono">{sessionId.slice(0, 20)}...</span>
                        </button>
                        <button
                          onClick={e => { e.stopPropagation(); setDeletingSession({ project, sessionId }) }}
                          disabled={deletingSessionId === sessionId}
                          className="p-1 mr-0.5 rounded opacity-0 group-hover/session:opacity-100 text-muted-foreground/50 hover:text-red-400 hover:bg-red-400/10 transition-all shrink-0"
                          title={t.sidebar.deleteSession}
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* ── 하단 설정 버튼 ── */}
      <div className="shrink-0 border-t border-border px-2 py-2">
        <button
          onClick={() => setShowSettings(true)}
          className="flex items-center gap-2 w-full px-2 py-1.5 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
        >
          <Settings className="w-3.5 h-3.5" />
          {t.sidebar.settings}
        </button>
      </div>

      {showSettings && <SettingsModal onClose={() => setShowSettings(false)} />}

      {deletingProject && (
        <DeleteConfirmDialog
          project={deletingProject}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeletingProject(null)}
        />
      )}

      {deletingSession && (
        <div
          className="absolute inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-[2px] p-4"
          onClick={() => setDeletingSession(null)}
        >
          <div
            className="w-full max-w-xs bg-card border border-border rounded-xl p-4 space-y-3 shadow-2xl"
            onClick={e => e.stopPropagation()}
          >
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-full bg-red-500/15 flex items-center justify-center shrink-0">
                <AlertTriangle className="w-4 h-4 text-red-400" />
              </div>
              <span className="text-sm font-semibold text-foreground">{t.sidebar.deleteSessionTitle}</span>
            </div>
            <div className="space-y-1.5">
              <p className="text-xs text-foreground/80">
                <span className="font-mono text-foreground">{deletingSession.sessionId.slice(0, 12)}...</span> {t.sidebar.deleteSessionDesc}
              </p>
              <p className="text-[11px] text-muted-foreground/50">{t.sidebar.deleteSessionWarning}</p>
            </div>
            <div className="flex gap-2 pt-1">
              <button
                onClick={() => setDeletingSession(null)}
                className="flex-1 py-1.5 rounded-lg text-xs font-medium bg-muted text-foreground/70 hover:bg-accent transition-colors"
              >
                {t.sidebar.cancel}
              </button>
              <button
                onClick={handleSessionDeleteConfirm}
                className="flex-1 py-1.5 rounded-lg text-xs font-medium bg-red-600 text-white hover:bg-red-500 transition-colors"
              >
                {t.sidebar.delete}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
