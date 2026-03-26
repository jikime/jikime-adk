'use client'

import { useState, useEffect, useCallback, memo } from 'react'
import {
  Folder, FolderOpen, MessageSquare, Plus, RefreshCw,
  ChevronDown, ChevronRight, X, Check,
  Server, Trash2, Edit2, Globe, AlertTriangle, Eye, EyeOff, KeyRound,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogAction, AlertDialogCancel,
} from '@/components/ui/alert-dialog'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
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
  gitPat?: string
}

const SETTINGS_KEY = 'webchat_settings'

const DEFAULT_SETTINGS: AppSettings = {
  model: 'claude-sonnet-4-6',
  permissionMode: 'bypassPermissions',
  gitPat: '',
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
      <Input
        value={name} onChange={e => setName(e.target.value)}
        placeholder={t.sidebar.serverName}
        className="h-7 text-base"
      />
      <Input
        value={host} onChange={e => handleHostChange(e.target.value)}
        placeholder={t.sidebar.serverHost}
        className="h-7 text-base font-mono"
      />
      <div className="flex items-center gap-2">
        <Checkbox
          id="secure-connection"
          checked={secure}
          onCheckedChange={(v) => setSecure(v as boolean)}
        />
        <Label htmlFor="secure-connection" className="text-base text-muted-foreground cursor-pointer">
          {t.sidebar.secureConnection}
        </Label>
      </div>
      <div className="flex gap-2 pt-1">
        <Button
          size="sm"
          onClick={() => valid && onSave({ name: name.trim(), host: host.trim(), secure })}
          disabled={!valid}
          className="flex-1 h-7 text-base"
        >
          {t.sidebar.save}
        </Button>
        <Button
          size="sm" variant="outline"
          onClick={onCancel}
          className="flex-1 h-7 text-base"
        >
          {t.sidebar.cancel}
        </Button>
      </div>
    </div>
  )
}

// ── 설정 모달 ─────────────────────────────────────────────────────
export function SettingsModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useLocale()
  const [settings, setSettings] = useState<AppSettings>(() => loadSettings())
  const [showPat, setShowPat] = useState(false)

  const update = (patch: Partial<AppSettings>) => {
    const next = { ...settings, ...patch }
    setSettings(next)
    saveSettings(next)
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-xs sm:max-w-sm p-0 gap-0 overflow-hidden" showCloseButton={false}>
        <DialogHeader className="flex flex-row items-center justify-between px-4 py-3 border-b border-border">
          <DialogTitle className="text-lg font-semibold">{t.sidebar.settingsTitle}</DialogTitle>
          <Button variant="ghost" size="icon-sm" onClick={onClose} className="h-6 w-6">
            <X className="w-4 h-4" />
          </Button>
        </DialogHeader>

        <div className="space-y-5 p-4 max-h-[70vh] overflow-y-auto">
          {/* ── 기본 모델 ── */}
          <div className="space-y-2">
            <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">{t.sidebar.defaultModel}</p>
            <div className="flex flex-col gap-1">
              {MODELS.map(m => (
                <button
                  key={m.id}
                  onClick={() => update({ model: m.id })}
                  className={cn(
                    'flex items-center gap-2 px-3 py-2 rounded-lg text-base text-left transition-colors',
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
                  'flex-1 px-3 py-2 rounded-lg text-base font-medium transition-colors',
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
                  'flex-1 px-3 py-2 rounded-lg text-base font-medium transition-colors',
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

          {/* ── Git PAT ── */}
          <div className="space-y-2">
            <Label className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-1.5">
              <KeyRound className="w-3 h-3" />
              {t.sidebar.gitPatTitle}
            </Label>
            <div className="relative">
              <Input
                type={showPat ? 'text' : 'password'}
                value={settings.gitPat ?? ''}
                onChange={e => update({ gitPat: e.target.value })}
                placeholder={t.sidebar.gitPatPlaceholder}
                className="h-8 text-base font-mono pr-8"
              />
              <Button
                type="button" variant="ghost" size="icon-sm"
                onClick={() => setShowPat(v => !v)}
                className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
              >
                {showPat ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
              </Button>
            </div>
            <p className="text-[10px] dark:text-muted-foreground/50 text-muted-foreground/70 leading-relaxed">
              {t.sidebar.gitPatDesc}
            </p>
          </div>

          {/* 현재 설정 요약 */}
          <div className="text-[10px] dark:text-muted-foreground/50 text-muted-foreground/70 pt-3">
            {t.sidebar.activeModel}: <span className="text-foreground/70">{MODELS.find(m => m.id === settings.model)?.label}</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}


// ── 세션 아이템 ───────────────────────────────────────────────────
const SessionItem = memo(function SessionItem({
  sessionId, project, isActive, isDeleting,
  onSelect, onDelete,
}: {
  sessionId: string
  project: Project
  isActive: boolean
  isDeleting: boolean
  onSelect: (project: Project, sessionId: string) => void
  onDelete: (project: Project, sessionId: string) => void
}) {
  const { t } = useLocale()
  return (
    <div className="flex items-center group/session">
      <button
        className={cn(
          'flex flex-1 items-center gap-2 pl-2 pr-1 py-1 text-base rounded-md transition-colors min-w-0',
          isActive
            ? 'bg-accent text-foreground'
            : 'text-muted-foreground hover:text-foreground/80 hover:bg-muted'
        )}
        onClick={() => onSelect(project, sessionId)}
      >
        {isDeleting
          ? <RefreshCw className="w-3 h-3 shrink-0 animate-spin" />
          : <MessageSquare className="w-3 h-3 shrink-0" />
        }
        <span className="truncate font-mono">{sessionId.slice(0, 20)}...</span>
      </button>
      <button
        onClick={e => { e.stopPropagation(); onDelete(project, sessionId) }}
        disabled={isDeleting}
        className="p-1 mr-0.5 rounded opacity-0 group-hover/session:opacity-100 text-muted-foreground/50 hover:text-red-400 hover:bg-red-400/10 transition-all shrink-0"
        title={t.sidebar.deleteSession}
      >
        <Trash2 className="w-3 h-3" />
      </button>
    </div>
  )
})

// ── 프로젝트 아이템 ───────────────────────────────────────────────
const ProjectItem = memo(function ProjectItem({
  project, isOpen, isActive, deletingId, deletingSessionId, activeSessionId,
  onToggle, onSelect, onDeleteProject, onNewChat, onSelectSession, onDeleteSession,
}: {
  project: Project
  isOpen: boolean
  isActive: boolean
  deletingId: string | null
  deletingSessionId: string | null
  activeSessionId: string | null
  onToggle: (id: string) => void
  onSelect: (p: Project) => void
  onDeleteProject: (p: Project) => void
  onNewChat: (p: Project) => void
  onSelectSession: (project: Project, sessionId: string) => void
  onDeleteSession: (project: Project, sessionId: string) => void
}) {
  const { t } = useLocale()
  return (
    <div>
      <div className="flex items-center group">
        <button
          onClick={() => onToggle(project.id)}
          className="p-0.5 ml-1 text-muted-foreground hover:text-foreground/80"
        >
          {isOpen
            ? <ChevronDown className="w-3 h-3" />
            : <ChevronRight className="w-3 h-3" />}
        </button>
        <button
          className={cn(
            'flex flex-1 items-center gap-2 px-2 py-1.5 text-left rounded-md ml-1 text-lg transition-colors min-w-0',
            isActive ? 'bg-accent text-foreground' : 'text-foreground/80 hover:bg-muted'
          )}
          onClick={() => onSelect(project)}
        >
          {deletingId === project.id
            ? <RefreshCw className="w-3.5 h-3.5 text-muted-foreground shrink-0 animate-spin" />
            : isOpen
              ? <FolderOpen className="w-3.5 h-3.5 text-blue-400 shrink-0" />
              : <Folder className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
          }
          <span className="truncate text-base">{project.name}</span>
        </button>
        <button
          onClick={e => { e.stopPropagation(); onDeleteProject(project) }}
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
            className="flex items-center gap-2 px-3 py-1 text-base text-muted-foreground hover:text-foreground/80 w-full hover:bg-muted rounded-md"
            onClick={() => onNewChat(project)}
          >
            <Plus className="w-3 h-3" />
            {t.sidebar.newChat}
          </button>
          {project.sessions.slice(0, 10).map(sessionId => (
            <SessionItem
              key={sessionId}
              sessionId={sessionId}
              project={project}
              isActive={activeSessionId === sessionId}
              isDeleting={deletingSessionId === sessionId}
              onSelect={onSelectSession}
              onDelete={onDeleteSession}
            />
          ))}
        </div>
      )}
    </div>
  )
})

// ── 메인 사이드바 ─────────────────────────────────────────────────
function Sidebar() {
  const { t } = useLocale()
  const { projects, activeProject, activeSessionId, setActiveProject, setActiveSessionId, navigateToSession, refreshProjects } = useProject()
  const { servers, activeServer, setActiveServerId, addServer, updateServer, removeServer, getApiUrl } = useServer()
  const { isConnected } = useWebSocket()

  const [loading,          setLoading]         = useState(false)
  const [openProjects,     setOpenProjects]     = useState<Set<string>>(new Set())

  // URL 복원으로 activeProject가 설정되면 해당 프로젝트 트리를 자동으로 열기
  useEffect(() => {
    if (activeProject) {
      setOpenProjects(prev => prev.has(activeProject.id) ? prev : new Set([...prev, activeProject.id]))
    }
  }, [activeProject?.id])
  const [showServerPicker, setShowServerPicker] = useState(false)
  const [addingServer,     setAddingServer]     = useState(false)
  const [editingServerId,  setEditingServerId]  = useState<string | null>(null)
  const [deletingProject,  setDeletingProject]  = useState<Project | null>(null)
  const [deletingId,       setDeletingId]       = useState<string | null>(null)
  // 세션 삭제
  type SessionTarget = { project: Project; sessionId: string }
  const [deletingSession,    setDeletingSession]    = useState<SessionTarget | null>(null)
  const [deletingSessionId,  setDeletingSessionId]  = useState<string | null>(null)
  const [showAddProject,   setShowAddProject]   = useState(false)
  const [addProjectPath,   setAddProjectPath]   = useState('')
  const [addProjectBusy,   setAddProjectBusy]   = useState(false)
  const [addProjectError,  setAddProjectError]  = useState('')

  const handleAddProject = useCallback(async () => {
    const p = addProjectPath.trim()
    if (!p) return
    setAddProjectBusy(true)
    setAddProjectError('')
    try {
      const res = await fetch(getApiUrl('/api/ws/project'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: p }),
      })
      const data = await res.json() as { ok?: boolean; error?: string; project?: { id: string; name: string; path: string } }
      if (!res.ok) { setAddProjectError(data.error || '등록 실패'); return }
      await refreshProjects()
      setShowAddProject(false)
      setAddProjectPath('')
    } catch (e) {
      setAddProjectError(String(e))
    } finally {
      setAddProjectBusy(false)
    }
  }, [addProjectPath, getApiUrl, refreshProjects])

  const handleRefresh = useCallback(async () => {
    setLoading(true)
    await refreshProjects()
    setLoading(false)
  }, [refreshProjects])

  const handleDeleteConfirm = useCallback(async () => {
    if (!deletingProject) return
    const id = deletingProject.id
    setDeletingProject(null)
    setDeletingId(id)
    try {
      const res = await fetch(getApiUrl(`/api/ws/project?id=${encodeURIComponent(id)}`), { method: 'DELETE' })
      if (res.ok) {
        if (activeProject?.id === id) {
          setActiveProject(null)
          setActiveSessionId(null)
        }
        await refreshProjects()
      }
    } catch { /* */ } finally {
      setDeletingId(null)
    }
  }, [deletingProject, getApiUrl, activeProject?.id, setActiveProject, setActiveSessionId, refreshProjects])

  const handleSessionDeleteConfirm = useCallback(async () => {
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
  }, [deletingSession, getApiUrl, activeSessionId, setActiveSessionId, refreshProjects])

  const toggleProject = useCallback((id: string) => {
    setOpenProjects(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const selectProject = useCallback((p: Project) => {
    // 프로젝트 폴더 클릭: URL 변경 없이 상태만 업데이트 + 하위 세션 목록 열기
    setActiveProject(p)
    setActiveSessionId(null)
    setOpenProjects(prev => prev.has(p.id) ? prev : new Set([...prev, p.id]))
  }, [setActiveProject, setActiveSessionId])

  const handleNewChat = useCallback(async (p: Project) => {
    try {
      const res = await fetch(getApiUrl('/api/ws/session/new'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectId: p.id }),
      })
      if (!res.ok) return
      const { sessionId } = await res.json() as { sessionId: string }
      navigateToSession(p, sessionId)
      await refreshProjects()
    } catch { /* */ }
  }, [getApiUrl, navigateToSession, refreshProjects])

  return (
    <div className="relative flex flex-col h-full bg-white dark:bg-muted">

      {/* ── 서버 선택기 ── */}
      <div className="shrink-0">
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
            <p className="text-base font-medium text-foreground truncate">{activeServer?.name ?? t.sidebar.noServer}</p>
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
          <div>
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
                          <p className={cn('text-base truncate leading-tight font-medium', activeServer?.id === s.id ? 'text-foreground' : 'dark:text-muted-foreground text-foreground/70')}>
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
                          <Button
                            variant="ghost" size="icon-sm"
                            onClick={() => setEditingServerId(s.id)}
                            className="h-6 w-6 text-muted-foreground/50 hover:text-foreground"
                          >
                            <Edit2 className="w-3 h-3" />
                          </Button>
                          <Button
                            variant="ghost" size="icon-sm"
                            onClick={() => removeServer(s.id)}
                            className="h-6 w-6 text-muted-foreground/50 hover:text-red-400"
                          >
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* 서버 추가 */}
            <div className="p-2">
              {addingServer ? (
                <ServerForm
                  onSave={(s) => { addServer(s); setAddingServer(false) }}
                  onCancel={() => setAddingServer(false)}
                />
              ) : (
                <Button
                  variant="outline" size="sm"
                  onClick={() => setAddingServer(true)}
                  className="flex items-center gap-2 w-full h-7 text-base border-dashed text-muted-foreground hover:text-foreground"
                >
                  <Plus className="w-3 h-3" />
                  {t.sidebar.addServer}
                </Button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* ── 프로젝트 헤더 ── */}
      <div className="flex items-center justify-between px-3 py-2.5 shrink-0">
        <span className="text-base font-semibold text-muted-foreground uppercase tracking-wider">{t.sidebar.projects}</span>
        <div className="flex items-center gap-0.5">
          <Button
            variant="ghost" size="icon"
            className="h-6 w-6 text-muted-foreground hover:text-foreground"
            onClick={() => { setShowAddProject(true); setAddProjectPath(''); setAddProjectError('') }}
            title="경로 등록"
          >
            <Plus className="w-3 h-3" />
          </Button>
          <Button
            variant="ghost" size="icon"
            className="h-6 w-6 text-muted-foreground hover:text-foreground"
            onClick={handleRefresh}
            disabled={loading}
          >
            <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
          </Button>
        </div>
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
              <p className="text-base text-muted-foreground/50">{t.sidebar.noProjects}</p>
              <p className="text-base text-muted-foreground/30 mt-1">{t.sidebar.noProjectsHint}</p>
            </div>
          )}

          {projects.slice(0, 50).map((project) => (
            <ProjectItem
              key={project.id}
              project={project}
              isOpen={openProjects.has(project.id)}
              isActive={activeProject?.id === project.id && activeSessionId === null}
              deletingId={deletingId}
              deletingSessionId={deletingSessionId}
              activeSessionId={activeSessionId}
              onToggle={toggleProject}
              onSelect={selectProject}
              onDeleteProject={setDeletingProject}
              onNewChat={handleNewChat}
              onSelectSession={(p, sId) => {
                // 세션 클릭: URL을 /session/{sessionId}로 이동
                navigateToSession(p, sId || '')
              }}
              onDeleteSession={(p, sId) => setDeletingSession({ project: p, sessionId: sId })}
            />
          ))}
        </div>
      </div>


      {/* ── 프로젝트 삭제 확인 ── */}
      <AlertDialog open={!!deletingProject} onOpenChange={(open) => { if (!open) setDeletingProject(null) }}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <div className="flex items-center gap-2 mb-1">
              <div className="w-7 h-7 rounded-full bg-red-500/15 flex items-center justify-center shrink-0">
                <AlertTriangle className="w-4 h-4 text-red-400" />
              </div>
              <AlertDialogTitle className="text-lg">{t.sidebar.deleteProjectTitle}</AlertDialogTitle>
            </div>
            <AlertDialogDescription className="text-base space-y-1">
              <span className="font-medium text-foreground">{deletingProject?.name}</span>{' '}
              {t.sidebar.deleteProjectDesc}
              {(deletingProject?.sessions.length ?? 0) > 0 && (
                <span className="block text-amber-500 mt-1">
                  {t.sidebar.deleteSessionCount(deletingProject!.sessions.length)}
                </span>
              )}
              <span className="block text-muted-foreground/50 mt-1">{t.sidebar.deleteProjectWarning}</span>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel size="sm" className="text-base">{t.sidebar.cancel}</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              className="text-base bg-destructive text-white hover:bg-destructive/90"
              size="sm"
            >
              {t.sidebar.delete}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* ── 프로젝트 경로 등록 ── */}
      <Dialog open={showAddProject} onOpenChange={(v) => { if (!v) setShowAddProject(false) }}>
        <DialogContent className="w-[400px] max-w-[95vw] flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-lg">
              <Plus className="w-4 h-4 text-blue-400" />
              프로젝트 경로 등록
            </DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <label className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
              절대 경로 <span className="text-red-400">*</span>
            </label>
            <Input
              value={addProjectPath}
              onChange={(e) => setAddProjectPath(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') handleAddProject() }}
              placeholder="/Users/me/my-project"
              className="font-mono text-base"
              autoFocus
            />
            <p className="text-[11px] text-muted-foreground">
              폴더가 없어도 등록됩니다. Team 패널에서 해당 경로를 작업 디렉터리로 사용합니다.
            </p>
          </div>
          {addProjectError && (
            <div className="text-[12px] text-red-400 bg-red-950/30 border border-red-900/40 rounded-lg px-3 py-2">
              {addProjectError}
            </div>
          )}
          <div className="flex gap-2 justify-end">
            <Button variant="outline" size="sm" className="text-base" onClick={() => setShowAddProject(false)} disabled={addProjectBusy}>
              취소
            </Button>
            <Button
              size="sm" className="text-base"
              onClick={handleAddProject}
              disabled={addProjectBusy || !addProjectPath.trim()}
            >
              {addProjectBusy ? '등록 중…' : '등록'}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* ── 세션 삭제 확인 ── */}
      <AlertDialog open={!!deletingSession} onOpenChange={(open) => { if (!open) setDeletingSession(null) }}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <div className="flex items-center gap-2 mb-1">
              <div className="w-7 h-7 rounded-full bg-red-500/15 flex items-center justify-center shrink-0">
                <AlertTriangle className="w-4 h-4 text-red-400" />
              </div>
              <AlertDialogTitle className="text-lg">{t.sidebar.deleteSessionTitle}</AlertDialogTitle>
            </div>
            <AlertDialogDescription className="text-base space-y-1">
              <span className="font-mono text-foreground">{deletingSession?.sessionId.slice(0, 12)}...</span>{' '}
              {t.sidebar.deleteSessionDesc}
              <span className="block text-muted-foreground/50 mt-1">{t.sidebar.deleteSessionWarning}</span>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel size="sm" className="text-base">{t.sidebar.cancel}</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleSessionDeleteConfirm}
              className="text-base bg-destructive text-white hover:bg-destructive/90"
              size="sm"
            >
              {t.sidebar.delete}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

export default memo(Sidebar)
