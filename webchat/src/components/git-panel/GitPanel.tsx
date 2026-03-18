'use client'

import { useState, useEffect, useCallback } from 'react'
import { GitBranch, RefreshCw, GitCommit, Check, Plus, Minus, ArrowUp, ArrowDown, CircleDot, FolderOpen, X, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogAction,
} from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '@/components/ui/resizable'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { useLocale } from '@/contexts/LocaleContext'
import { loadSettings } from '@/components/sidebar/Sidebar'
import IssuesPanel from './IssuesPanel'

// ── API 헬퍼 ──────────────────────────────────────────────────────
async function gitCmd(apiUrl: string, cwd: string, body: Record<string, unknown>, pat?: string): Promise<string> {
  const res = await fetch(apiUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ cwd, ...(pat ? { pat } : {}), ...body }),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'Git error')
  if (data.notGit) throw new Error('NOT_GIT_REPO')
  return (data.output as string) ?? ''
}

// ── 타입 ──────────────────────────────────────────────────────────
interface GitFile {
  status: string   // XY (two-char git status)
  path: string
}

type Tab = 'changes' | 'log' | 'branches' | 'issues'

// ── git status --short 파싱 ────────────────────────────────────────
function parseStatus(raw: string): GitFile[] {
  return raw
    .split('\n')
    .filter(Boolean)
    .map(line => ({ status: line.slice(0, 2), path: line.slice(3).trim() }))
}

// 상태 코드 → 라벨/색상
function statusBadge(xy: string): { label: string; color: string } {
  const x = xy[0], y = xy[1]
  if (xy === '??') return { label: 'U', color: 'text-zinc-400 bg-zinc-700' }
  if (x === 'A')   return { label: 'A', color: 'text-emerald-400 bg-emerald-900/40' }
  if (x === 'D' || y === 'D') return { label: 'D', color: 'text-red-400 bg-red-900/40' }
  if (x === 'R')   return { label: 'R', color: 'text-blue-400 bg-blue-900/40' }
  return { label: 'M', color: 'text-amber-400 bg-amber-900/40' }
}

// ── Diff 컬러 렌더러 ──────────────────────────────────────────────
function DiffViewer({ content, noChangesLabel }: { content: string; noChangesLabel: string }) {
  if (!content.trim()) {
    return <p className="text-xs text-muted-foreground text-center py-8">{noChangesLabel}</p>
  }
  return (
    <div className="font-mono text-xs leading-5 pt-4 pb-16">
      {content.split('\n').map((line, i) => {
        const isAdd    = line.startsWith('+') && !line.startsWith('+++')
        const isRemove = line.startsWith('-') && !line.startsWith('---')
        const isHunk   = line.startsWith('@@')
        const isMeta   = line.startsWith('diff') || line.startsWith('index') || line.startsWith('---') || line.startsWith('+++')
        return (
          <div key={i} className={cn(
            'px-2 whitespace-pre-wrap break-all',
            isAdd    && 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
            isRemove && 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-300',
            isHunk   && 'bg-blue-50 text-blue-700 font-semibold dark:bg-blue-950/30 dark:text-blue-300',
            isMeta   && 'text-muted-foreground',
            !isAdd && !isRemove && !isHunk && !isMeta && 'text-foreground/60',
          )}>
            {isAdd    && <Plus  className="inline w-3 h-3 mr-1 shrink-0" />}
            {isRemove && <Minus className="inline w-3 h-3 mr-1 shrink-0" />}
            {line}
          </div>
        )
      })}
    </div>
  )
}

// ── 메인 컴포넌트 ─────────────────────────────────────────────────
export default function GitPanel() {
  const { t } = useLocale()
  const { activeProject } = useProject()
  const { getApiUrl }     = useServer()
  const cwd = activeProject?.path ?? ''
  const gitUrl = getApiUrl('/api/ws/git')

  const pat = loadSettings().gitPat ?? ''

  const [tab, setTab]               = useState<Tab>('changes')
  const [loading, setLoading]       = useState(false)
  const [isGitRepo, setIsGitRepo]   = useState(true)
  const [pushing, setPushing]       = useState(false)
  const [pulling, setPulling]       = useState(false)
  const [remoteUrl, setRemoteUrl]   = useState<string | null>(null)
  const [errorMsg, setErrorMsg]     = useState<string | null>(null)

  // 변경사항 탭
  const [files, setFiles]           = useState<GitFile[]>([])
  const [checked, setChecked]       = useState<Set<string>>(new Set())
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [diff, setDiff]             = useState('')
  const [diffLoading, setDiffLoading] = useState(false)
  const [commitMsg, setCommitMsg]   = useState('')
  const [committing, setCommitting] = useState(false)
  const [commitPanelOpen, setCommitPanelOpen] = useState(false)

  // 로그 탭
  const [log, setLog]               = useState('')
  const [selectedCommit, setSelectedCommit] = useState<string | null>(null)
  const [commitDiff, setCommitDiff]         = useState('')
  const [commitDiffLoading, setCommitDiffLoading] = useState(false)

  // 브랜치 탭
  const [branches, setBranches]     = useState<string[]>([])
  const [currentBranch, setCurrentBranch] = useState('')

  // ── 새로고침 ──────────────────────────────────────────────────
  const refresh = useCallback(async () => {
    if (!cwd) return
    setLoading(true)
    setIsGitRepo(true)
    try {
      const [statusRaw, logRaw, branchRaw] = await Promise.all([
        gitCmd(gitUrl, cwd, { action: 'status' }),
        gitCmd(gitUrl, cwd, { action: 'log'    }),
        gitCmd(gitUrl, cwd, { action: 'branch' }),
      ])
      gitCmd(gitUrl, cwd, { action: 'custom', args: ['remote', 'get-url', 'origin'] })
        .then(url => setRemoteUrl(url.trim()))
        .catch(() => setRemoteUrl(null))
      setFiles(parseStatus(statusRaw))
      setLog(logRaw)

      const branchList = branchRaw.split('\n').filter(Boolean)
      setBranches(branchList)
      const cur = branchList.find(b => b.startsWith('* '))?.replace('* ', '') ?? ''
      setCurrentBranch(cur)

      if (selectedFile) loadDiff(selectedFile)
    } catch (e: unknown) {
      if ((e as Error).message === 'NOT_GIT_REPO') setIsGitRepo(false)
    } finally {
      setLoading(false)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cwd, selectedFile, gitUrl])

  useEffect(() => { refresh() }, [refresh])

  // ── 커밋 diff 로드 ────────────────────────────────────────────
  const loadCommitDiff = useCallback(async (hash: string) => {
    if (!cwd) return
    setCommitDiffLoading(true)
    try {
      const out = await gitCmd(gitUrl, cwd, { action: 'custom', args: ['show', hash] })
      setCommitDiff(out)
    } catch { setCommitDiff('') }
    finally { setCommitDiffLoading(false) }
  }, [cwd, gitUrl])

  // ── 파일 diff 로드 ────────────────────────────────────────────
  const loadDiff = useCallback(async (filePath: string) => {
    if (!cwd) return
    setDiffLoading(true)
    try {
      const out = await gitCmd(gitUrl, cwd, { action: 'file_diff', file: filePath })
      setDiff(out)
    } catch { setDiff('') }
    finally { setDiffLoading(false) }
  }, [cwd, gitUrl])

  const handleSelectFile = (filePath: string) => {
    setSelectedFile(filePath)
    loadDiff(filePath)
  }

  // ── 체크박스 ─────────────────────────────────────────────────
  const toggleCheck = (filePath: string) => {
    setChecked(prev => {
      const next = new Set(prev)
      next.has(filePath) ? next.delete(filePath) : next.add(filePath)
      return next
    })
  }
  const toggleAll = () => {
    setChecked(prev => prev.size === files.length ? new Set() : new Set(files.map(f => f.path)))
  }

  // ── 커밋 ────────────────────────────────────────────────────
  const handleCommit = async () => {
    if (!cwd || !commitMsg.trim() || checked.size === 0) return
    setCommitting(true)
    try {
      await gitCmd(gitUrl, cwd, { action: 'add', files: Array.from(checked) })
      await gitCmd(gitUrl, cwd, { action: 'commit', message: commitMsg.trim() })
      setCommitMsg('')
      setChecked(new Set())
      setSelectedFile(null)
      setDiff('')
      setCommitPanelOpen(false)
      await refresh()
    } catch (e: unknown) {
      setErrorMsg((e as Error).message)
    } finally {
      setCommitting(false)
    }
  }

  // ── Push ────────────────────────────────────────────────────
  const handlePush = async () => {
    if (!cwd) return
    setPushing(true)
    try {
      await gitCmd(gitUrl, cwd, { action: 'push' }, pat || undefined)
      await refresh()
    } catch (e: unknown) {
      setErrorMsg((e as Error).message)
    } finally {
      setPushing(false)
    }
  }

  // ── Pull ────────────────────────────────────────────────────
  const handlePull = async () => {
    if (!cwd) return
    setPulling(true)
    try {
      await gitCmd(gitUrl, cwd, { action: 'pull' }, pat || undefined)
      await refresh()
    } catch (e: unknown) {
      setErrorMsg((e as Error).message)
    } finally {
      setPulling(false)
    }
  }

  // ── 브랜치 체크아웃 ──────────────────────────────────────────
  const handleCheckout = async (branch: string) => {
    const name = branch.replace('* ', '').replace('remotes/origin/', '').trim()
    if (!cwd || name === currentBranch) return
    try {
      await gitCmd(gitUrl, cwd, { action: 'checkout', branch: name })
      await refresh()
    } catch (e: unknown) {
      setErrorMsg((e as Error).message)
    }
  }

  // ── Not a git repo ────────────────────────────────────────────
  if (!activeProject) {
    return (
      <div className="flex items-center justify-center flex-1 h-full text-xs text-muted-foreground/50">
        {t.git.selectProject}
      </div>
    )
  }

  if (!isGitRepo) {
    return (
      <div className="flex flex-col h-full bg-background rounded-lg overflow-hidden border border-zinc-300 dark:border-zinc-600">
        <GitHeader currentBranch="" />
        <div className="flex flex-col items-center justify-center flex-1 gap-2 text-center px-4">
          <GitBranch className="w-8 h-8 text-muted-foreground/30" />
          <p className="text-sm text-muted-foreground">{t.git.notGitRepo}</p>
          <p className="text-xs text-muted-foreground/50 font-mono break-all">{cwd}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full bg-background rounded-lg overflow-hidden border border-zinc-300 dark:border-zinc-600">
      <GitHeader currentBranch={currentBranch} />

      {/* 탭 */}
      <div className="flex gap-0 px-2 pt-2 shrink-0 border-b border-border">
        {(['changes', 'log', 'branches', 'issues'] as Tab[]).map(tabId => (
          <button key={tabId} onClick={() => setTab(tabId)}
            className={cn(
              'px-3 py-1.5 text-xs font-medium rounded-t transition-colors flex items-center gap-1',
              tab === tabId ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground/80'
            )}>
            {tabId === 'issues' && <CircleDot className="w-3 h-3" />}
            {tabId === 'changes' ? `${t.git.changes}${files.length > 0 ? ` (${files.length})` : ''}` :
             tabId === 'log'     ? t.git.log :
             tabId === 'branches'? t.git.branch :
             t.git.issues}
          </button>
        ))}
      </div>

      {/* Remote URL 인포바 — 모든 탭 공통, 고정 높이 h-7 */}
      {remoteUrl && (
        <div className="flex items-center gap-1.5 px-3 h-7 border-b border-border/50 bg-muted/30 shrink-0">
          <GitBranch className="w-3 h-3 text-muted-foreground/50 shrink-0" />
          <span className="text-[11px] font-mono text-muted-foreground truncate flex-1 min-w-0" title={remoteUrl}>
            {remoteUrl}
          </span>
          {tab === 'changes' && (
            <div className="flex items-center gap-1 shrink-0">
              <Button
                variant="ghost" size="sm"
                className="h-5 px-2 text-xs gap-1 text-sky-600 dark:text-sky-400 hover:bg-sky-500/10 disabled:opacity-40"
                onClick={handlePull} disabled={pulling || loading}
              >
                <ArrowDown className={cn('w-3 h-3', pulling && 'animate-bounce')} />
                Pull
              </Button>
              <Button
                variant="ghost" size="sm"
                className="h-5 px-2 text-xs gap-1 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-500/10 disabled:opacity-40"
                onClick={handlePush} disabled={pushing || loading}
              >
                <ArrowUp className={cn('w-3 h-3', pushing && 'animate-bounce')} />
                Push
              </Button>
              <div className="w-px h-3.5 bg-border mx-0.5" />
              <Button
                variant="ghost" size="sm"
                className={cn(
                  'h-5 px-2 text-xs gap-1 disabled:opacity-40',
                  commitPanelOpen
                    ? 'bg-primary/10 text-primary'
                    : 'text-foreground/70 hover:bg-muted'
                )}
                onClick={() => setCommitPanelOpen(v => !v)}
                disabled={files.length === 0}
              >
                <GitCommit className="w-3 h-3" />
                {t.git.commit}
                {checked.size > 0 && (
                  <span className="ml-0.5 bg-primary text-primary-foreground text-[9px] font-bold rounded-full w-3.5 h-3.5 flex items-center justify-center">
                    {checked.size}
                  </span>
                )}
              </Button>
            </div>
          )}
        </div>
      )}

      {/* 변경사항 탭 */}
      {tab === 'changes' && (
        <div className="flex flex-1 min-h-0 overflow-hidden">
          {/* 좌측: 파일 목록 + Diff */}
          <ResizablePanelGroup orientation="vertical" className="flex-1 min-w-0">
            {/* 파일 목록 */}
            <ResizablePanel defaultSize="40" minSize="80px" className="flex flex-col overflow-hidden border-b border-border">
              <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50 shrink-0">
                <Checkbox
                  checked={files.length > 0 && checked.size === files.length}
                  onCheckedChange={() => toggleAll()}
                  className="w-3 h-3"
                />
                <span className="text-xs text-muted-foreground flex-1">{t.git.selectAll}</span>
                <button
                  onClick={refresh}
                  disabled={loading}
                  className="text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted disabled:opacity-40"
                  title="Refresh"
                >
                  <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
                </button>
              </div>
              <ScrollArea className="flex-1">
                {files.length === 0
                  ? <p className="text-xs text-muted-foreground/50 text-center py-4">{t.git.noChanges}</p>
                  : files.map(f => {
                      const badge = statusBadge(f.status)
                      return (
                        <div key={f.path}
                          onClick={() => handleSelectFile(f.path)}
                          className={cn(
                            'flex items-center gap-2 px-3 py-1.5 cursor-pointer hover:bg-muted/50 transition-colors',
                            selectedFile === f.path && 'bg-muted'
                          )}>
                          <Checkbox
                            checked={checked.has(f.path)}
                            onCheckedChange={() => toggleCheck(f.path)}
                            onClick={(e: React.MouseEvent) => e.stopPropagation()}
                            className="w-3 h-3 shrink-0"
                          />
                          <span className={cn('text-xs font-bold rounded px-1 shrink-0', badge.color)}>
                            {badge.label}
                          </span>
                          <span className="text-xs text-foreground/80 truncate font-mono">{f.path}</span>
                        </div>
                      )
                    })
                }
              </ScrollArea>
            </ResizablePanel>

            <ResizableHandle withHandle orientation="horizontal" />

            {/* Diff 뷰어 */}
            <ResizablePanel minSize="80px">
              <ScrollArea className="h-full">
                {diffLoading
                  ? <p className="text-xs text-muted-foreground text-center py-4">{t.git.loading}</p>
                  : selectedFile
                  ? <DiffViewer content={diff} noChangesLabel={t.git.noChanges} />
                  : <p className="text-xs text-muted-foreground/50 text-center py-8">{t.git.selectFileDiff}</p>
                }
              </ScrollArea>
            </ResizablePanel>
          </ResizablePanelGroup>

          {/* 우측: 커밋 사이드 패널 (슬라이드인) */}
          <div className={cn(
            'flex flex-col shrink-0 border-l border-border bg-background overflow-hidden transition-all duration-200',
            commitPanelOpen ? 'w-60' : 'w-0'
          )}>
            {commitPanelOpen && (
              <>
                {/* 패널 헤더 — 좌측 파일 목록 헤더와 높이/테두리 일치 */}
                <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50 shrink-0">
                  <GitCommit className="w-3 h-3 text-primary shrink-0" />
                  <span className="text-xs font-semibold text-foreground flex-1">
                    {t.git.commit}
                  </span>
                  {checked.size > 0 && (
                    <span className="text-[10px] font-bold bg-primary/10 text-primary rounded-full px-1.5 leading-none py-[3px]">
                      {checked.size}
                    </span>
                  )}
                  <button
                    onClick={() => setCommitPanelOpen(false)}
                    className="text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted"
                    title="닫기"
                  >
                    <X className="w-3 h-3" />
                  </button>
                </div>

                {/* 스테이징 파일 목록 */}
                <div className="px-2 pt-2 pb-1.5 border-b border-border/50 shrink-0">
                  <div className="flex items-center justify-between px-1 mb-1.5">
                    <span className="text-[10px] font-medium text-muted-foreground">
                      {checked.size > 0 ? `${checked.size}개 파일 스테이징됨` : '파일을 선택하세요'}
                    </span>
                    {files.length > 0 && (
                      <button
                        onClick={toggleAll}
                        className="text-[10px] text-primary hover:text-primary/80 transition-colors"
                      >
                        {checked.size === files.length ? '전체 해제' : '전체 선택'}
                      </button>
                    )}
                  </div>
                  <ScrollArea className="max-h-32">
                    {checked.size === 0 ? (
                      <div className="flex flex-col items-center justify-center py-3 gap-1.5">
                        <ChevronRight className="w-4 h-4 text-muted-foreground/30" />
                        <p className="text-[10px] text-muted-foreground/40 text-center">
                          왼쪽 목록에서<br/>파일을 체크하세요
                        </p>
                      </div>
                    ) : (
                      Array.from(checked).map(fp => {
                        const f = files.find(x => x.path === fp)
                        const badge = f ? statusBadge(f.status) : { label: 'M', color: 'text-amber-400 bg-amber-900/40' }
                        return (
                          <div key={fp} className="flex items-center gap-1.5 px-1 py-0.5 group">
                            <span className={cn('text-[10px] font-bold rounded px-1 shrink-0', badge.color)}>
                              {badge.label}
                            </span>
                            <span className="text-[10px] font-mono text-foreground/70 truncate flex-1" title={fp}>{fp}</span>
                            <button
                              onClick={() => toggleCheck(fp)}
                              className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-all shrink-0"
                            >
                              <X className="w-2.5 h-2.5" />
                            </button>
                          </div>
                        )
                      })
                    )}
                  </ScrollArea>
                </div>

                {/* 커밋 메시지 입력 */}
                <div className="flex-1 flex flex-col px-2 py-2 gap-2 min-h-0">
                  <p className="text-[10px] font-medium text-muted-foreground px-0.5">커밋 메시지</p>
                  <textarea
                    value={commitMsg}
                    onChange={e => setCommitMsg(e.target.value)}
                    onKeyDown={e => { if (e.key === 'Enter' && e.metaKey) handleCommit() }}
                    placeholder={t.git.commitPlaceholder}
                    disabled={committing}
                    className="flex-1 w-full resize-none rounded-md border border-input bg-muted/30 px-2.5 py-2 text-xs text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-1 focus:ring-primary focus:bg-background disabled:opacity-40 transition-colors min-h-0"
                  />
                  <button
                    onClick={handleCommit}
                    disabled={!commitMsg.trim() || checked.size === 0 || committing}
                    className={cn(
                      'w-full h-7 rounded-md text-xs font-medium flex items-center justify-center gap-1.5 transition-colors',
                      'bg-primary text-primary-foreground hover:bg-primary/90',
                      'disabled:opacity-40 disabled:cursor-not-allowed'
                    )}
                  >
                    {committing
                      ? <RefreshCw className="w-3 h-3 animate-spin" />
                      : <GitCommit className="w-3 h-3" />
                    }
                    {committing ? '커밋 중...' : t.git.commit}
                  </button>
                  <p className="text-[10px] text-muted-foreground/40 text-center">⌘ + Enter</p>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* 로그 탭 */}
      {tab === 'log' && (
        <div className="flex flex-col flex-1 min-h-0">
          {/* 헤더 */}
          <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50 shrink-0">
            <GitCommit className="w-3 h-3 text-muted-foreground/50 shrink-0" />
            <span className="text-xs text-muted-foreground flex-1">{t.git.log}</span>
            {log && (
              <span className="text-[10px] text-muted-foreground/50">
                {log.split('\n').filter(Boolean).length}개
              </span>
            )}
            <button
              onClick={refresh}
              disabled={loading}
              className="text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted disabled:opacity-40"
              title="Refresh"
            >
              <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
            </button>
          </div>

          <ResizablePanelGroup orientation="vertical" className="flex-1 min-h-0">
            {/* 커밋 목록 */}
            <ResizablePanel defaultSize="40" minSize="80px" className="overflow-hidden border-b border-border">
              <ScrollArea className="h-full">
                <div className="px-2">
                  {log.split('\n').filter(Boolean).map((line, i) => {
                    const [hash, ...rest] = line.split(' ')
                    return (
                      <button
                        key={i}
                        onClick={() => { setSelectedCommit(hash); loadCommitDiff(hash) }}
                        className={cn(
                          'flex items-start gap-2 w-full py-1.5 px-1 rounded text-left border-b border-border/40 last:border-0 hover:bg-muted/50 transition-colors',
                          selectedCommit === hash && 'bg-muted'
                        )}
                      >
                        <span className="text-xs font-mono text-blue-400 shrink-0">{hash}</span>
                        <span className="text-xs text-foreground/80 truncate">{rest.join(' ')}</span>
                      </button>
                    )
                  })}
                  {!log && <p className="text-xs text-muted-foreground/50 text-center py-8">{t.git.noCommits}</p>}
                </div>
              </ScrollArea>
            </ResizablePanel>

            <ResizableHandle withHandle orientation="horizontal" />

            {/* Diff 뷰어 */}
            <ResizablePanel minSize="80px">
              {selectedCommit ? (
                <ScrollArea className="h-full">
                  {commitDiffLoading
                    ? <p className="text-xs text-muted-foreground text-center py-4">{t.git.loading}</p>
                    : <DiffViewer content={commitDiff} noChangesLabel={t.git.noChanges} />
                  }
                </ScrollArea>
              ) : (
                <div className="flex items-center justify-center h-full">
                  <p className="text-xs text-muted-foreground/40">커밋을 클릭하면 diff를 볼 수 있어요</p>
                </div>
              )}
            </ResizablePanel>
          </ResizablePanelGroup>
        </div>
      )}

      {/* 브랜치 탭 */}
      {tab === 'branches' && (
        <div className="flex flex-col flex-1 min-h-0">
          <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50 shrink-0">
            <GitBranch className="w-3 h-3 text-muted-foreground/50 shrink-0" />
            <span className="text-xs text-muted-foreground flex-1">{t.git.branch}</span>
            {branches.length > 0 && (
              <span className="text-[10px] text-muted-foreground/50">
                {branches.filter(Boolean).length}개
              </span>
            )}
            <button
              onClick={refresh}
              disabled={loading}
              className="text-muted-foreground hover:text-foreground transition-colors rounded hover:bg-muted disabled:opacity-40"
              title="Refresh"
            >
              <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
            </button>
          </div>
          <ScrollArea className="flex-1 min-h-0">
            <div className="p-2">
              {branches.filter(Boolean).map((b, i) => {
                const isCurrent = b.startsWith('* ')
                const name = b.replace('* ', '').trim()
                const isRemote = name.startsWith('remotes/')
                return (
                  <button key={i} onClick={() => handleCheckout(b)}
                    disabled={isCurrent || isRemote}
                    className={cn(
                      'flex items-center gap-2 w-full px-2 py-1.5 rounded text-xs text-left transition-colors',
                      isCurrent  ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 cursor-default' : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                      isRemote   && 'opacity-50 cursor-default',
                    )}>
                    {isCurrent && <Check className="w-3 h-3 shrink-0 text-purple-500 dark:text-purple-400" />}
                    {!isCurrent && <GitBranch className="w-3 h-3 shrink-0 text-muted-foreground/50" />}
                    <span className="font-mono truncate">{name}</span>
                    {isRemote && <span className="ml-auto text-muted-foreground/50 text-xs">remote</span>}
                  </button>
                )
              })}
              {branches.length === 0 && <p className="text-xs text-muted-foreground/50 text-center py-8">{t.git.noBranches}</p>}
            </div>
          </ScrollArea>
        </div>
      )}

      {/* Issues 탭 */}
      {tab === 'issues' && (
        <div className="flex-1 min-h-0 overflow-hidden">
          <IssuesPanel />
        </div>
      )}

      {/* ── 에러 알림 ── */}
      <AlertDialog open={!!errorMsg} onOpenChange={(open) => { if (!open) setErrorMsg(null) }}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle className="text-sm text-destructive">Git 오류</AlertDialogTitle>
            <AlertDialogDescription className="text-xs font-mono whitespace-pre-wrap break-all">
              {errorMsg}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={() => setErrorMsg(null)} size="sm" className="text-xs">
              확인
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

// ── 헤더 서브컴포넌트 ─────────────────────────────────────────────
function GitHeader({ currentBranch }: {
  currentBranch: string
}) {
  const { activeProject } = useProject()
  const projectPath = activeProject?.path ?? ''

  return (
    <div className="flex items-center px-3 py-2.5 bg-white dark:bg-accent border-b border-border shrink-0">
      <div className="flex items-center gap-2 min-w-0">
        <GitBranch className="w-4 h-4 text-purple-400 shrink-0" />
        <span className="text-sm font-medium text-foreground shrink-0">Git</span>
        {currentBranch && (
          <span className="text-xs text-muted-foreground font-mono bg-muted px-1.5 py-0.5 rounded shrink-0">
            {currentBranch}
          </span>
        )}
        {projectPath && (
          <span
            className="inline-flex items-center gap-1 text-[11px] font-mono bg-violet-50 dark:bg-violet-500/10 text-violet-600 dark:text-violet-400 border border-violet-200 dark:border-violet-500/30 rounded px-1.5 py-0.5 truncate"
            title={projectPath}
          >
            <FolderOpen className="w-3 h-3 shrink-0" />
            <span className="truncate">{projectPath}</span>
          </span>
        )}
      </div>
    </div>
  )
}
