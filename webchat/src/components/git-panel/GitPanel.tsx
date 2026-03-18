'use client'

import { useState, useEffect, useCallback } from 'react'
import { GitBranch, RefreshCw, GitCommit, Check, Plus, Minus, ArrowUp, ArrowDown, CircleDot } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogAction,
} from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'
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
    <div className="font-mono text-xs leading-5">
      {content.split('\n').map((line, i) => {
        const isAdd    = line.startsWith('+') && !line.startsWith('+++')
        const isRemove = line.startsWith('-') && !line.startsWith('---')
        const isHunk   = line.startsWith('@@')
        const isMeta   = line.startsWith('diff') || line.startsWith('index') || line.startsWith('---') || line.startsWith('+++')
        return (
          <div key={i} className={cn(
            'px-2 whitespace-pre-wrap break-all',
            isAdd    && 'bg-emerald-950/40 text-emerald-300',
            isRemove && 'bg-red-950/40 text-red-300',
            isHunk   && 'bg-blue-950/30 text-blue-300 font-semibold',
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
  const [errorMsg, setErrorMsg]     = useState<string | null>(null)

  // 변경사항 탭
  const [files, setFiles]           = useState<GitFile[]>([])
  const [checked, setChecked]       = useState<Set<string>>(new Set())
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [diff, setDiff]             = useState('')
  const [diffLoading, setDiffLoading] = useState(false)
  const [commitMsg, setCommitMsg]   = useState('')
  const [committing, setCommitting] = useState(false)

  // 로그 탭
  const [log, setLog]               = useState('')

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
      <div className="flex flex-col h-full bg-muted dark:bg-background rounded-lg overflow-hidden border border-border">
        <GitHeader currentBranch="" loading={loading} onRefresh={refresh} pushing={pushing} pulling={pulling} onPush={handlePush} onPull={handlePull} />
        <div className="flex flex-col items-center justify-center flex-1 gap-2 text-center px-4">
          <GitBranch className="w-8 h-8 text-muted-foreground/30" />
          <p className="text-sm text-muted-foreground">{t.git.notGitRepo}</p>
          <p className="text-xs text-muted-foreground/50 font-mono break-all">{cwd}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full bg-muted dark:bg-background rounded-lg overflow-hidden border border-border">
      <GitHeader
        currentBranch={currentBranch}
        loading={loading}
        onRefresh={refresh}
        pushing={pushing}
        pulling={pulling}
        onPush={handlePush}
        onPull={handlePull}
      />

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

      {/* 변경사항 탭 */}
      {tab === 'changes' && (
        <div className="flex flex-col flex-1 min-h-0">
          {/* 파일 목록 */}
          <div className="shrink-0 border-b border-border" style={{ maxHeight: '40%' }}>
            <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border/50">
              <Checkbox
                checked={files.length > 0 && checked.size === files.length}
                onCheckedChange={() => toggleAll()}
                className="w-3 h-3"
              />
              <span className="text-xs text-muted-foreground">{t.git.selectAll}</span>
            </div>
            <ScrollArea style={{ maxHeight: 'calc(40vh - 60px)' }}>
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
          </div>

          {/* Diff 뷰어 */}
          <div className="flex-1 min-h-0 overflow-hidden">
            <ScrollArea className="h-full">
              {diffLoading
                ? <p className="text-xs text-muted-foreground text-center py-4">{t.git.loading}</p>
                : selectedFile
                ? <DiffViewer content={diff} noChangesLabel={t.git.noChanges} />
                : <p className="text-xs text-muted-foreground/50 text-center py-8">{t.git.selectFileDiff}</p>
              }
            </ScrollArea>
          </div>

          {/* 커밋 영역 */}
          <div className="shrink-0 border-t border-border p-2 space-y-1.5">
            <div className="flex gap-1.5">
              <Input
                value={commitMsg}
                onChange={e => setCommitMsg(e.target.value)}
                onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) handleCommit() }}
                placeholder={checked.size > 0 ? t.git.commitPlaceholder : t.git.selectFilePlaceholder}
                disabled={checked.size === 0 || committing}
                className="flex-1 h-7 text-xs disabled:opacity-40"
              />
              <Button
                size="sm"
                onClick={handleCommit}
                disabled={!commitMsg.trim() || checked.size === 0 || committing}
                className="shrink-0 text-xs h-7 px-2.5 bg-emerald-700 hover:bg-emerald-600 disabled:opacity-40"
              >
                <GitCommit className="w-3 h-3 mr-1" />
                {t.git.commit}
              </Button>
            </div>
            {checked.size > 0 && (
              <p className="text-xs text-muted-foreground">{checked.size}개 파일 선택됨</p>
            )}
          </div>
        </div>
      )}

      {/* 로그 탭 */}
      {tab === 'log' && (
        <ScrollArea className="flex-1 min-h-0 p-2">
          {log.split('\n').filter(Boolean).map((line, i) => {
            const [hash, ...rest] = line.split(' ')
            return (
              <div key={i} className="flex items-start gap-2 py-1.5 border-b border-border/50 last:border-0">
                <span className="text-xs font-mono text-blue-400 shrink-0">{hash}</span>
                <span className="text-xs text-foreground/80">{rest.join(' ')}</span>
              </div>
            )
          })}
          {!log && <p className="text-xs text-muted-foreground/50 text-center py-8">{t.git.noCommits}</p>}
        </ScrollArea>
      )}

      {/* 브랜치 탭 */}
      {tab === 'branches' && (
        <ScrollArea className="flex-1 min-h-0 p-2">
          {branches.filter(Boolean).map((b, i) => {
            const isCurrent = b.startsWith('* ')
            const name = b.replace('* ', '').trim()
            const isRemote = name.startsWith('remotes/')
            return (
              <button key={i} onClick={() => handleCheckout(b)}
                disabled={isCurrent || isRemote}
                className={cn(
                  'flex items-center gap-2 w-full px-2 py-1.5 rounded text-xs text-left transition-colors',
                  isCurrent  ? 'bg-purple-900/30 text-purple-300 cursor-default' : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                  isRemote   && 'opacity-50 cursor-default',
                )}>
                {isCurrent && <Check className="w-3 h-3 shrink-0 text-purple-400" />}
                {!isCurrent && <GitBranch className="w-3 h-3 shrink-0 text-muted-foreground/50" />}
                <span className="font-mono truncate">{name}</span>
                {isRemote && <span className="ml-auto text-muted-foreground/50 text-xs">remote</span>}
              </button>
            )
          })}
          {branches.length === 0 && <p className="text-xs text-muted-foreground/50 text-center py-8">{t.git.noBranches}</p>}
        </ScrollArea>
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
function GitHeader({ currentBranch, loading, onRefresh, pushing, pulling, onPush, onPull }: {
  currentBranch: string
  loading: boolean
  onRefresh: () => void
  pushing: boolean
  pulling: boolean
  onPush: () => void
  onPull: () => void
}) {
  return (
    <div className="flex items-center justify-between px-3 py-2.5 bg-white dark:bg-accent border-b border-border shrink-0">
      <div className="flex items-center gap-2">
        <GitBranch className="w-4 h-4 text-purple-400" />
        <span className="text-sm font-medium text-foreground">Git</span>
        {currentBranch && (
          <span className="text-xs text-muted-foreground font-mono bg-muted px-1.5 py-0.5 rounded">
            {currentBranch}
          </span>
        )}
      </div>
      <div className="flex items-center gap-1.5">
        {/* Pull 버튼 */}
        <Button
          variant="outline" size="sm"
          className="h-6 px-2.5 text-xs font-medium gap-1 border-sky-500/60 dark:bg-sky-500/10 bg-sky-50 dark:text-sky-300 text-sky-700 hover:dark:bg-sky-500/20 hover:bg-sky-100 disabled:opacity-40"
          onClick={onPull} disabled={pulling || loading}
          title="git pull"
        >
          <ArrowDown className={cn('w-3 h-3', pulling && 'animate-bounce')} />
          Pull
        </Button>
        {/* Push 버튼 */}
        <Button
          variant="outline" size="sm"
          className="h-6 px-2.5 text-xs font-medium gap-1 border-emerald-500/60 dark:bg-emerald-500/10 bg-emerald-50 dark:text-emerald-300 text-emerald-700 hover:dark:bg-emerald-500/20 hover:bg-emerald-100 disabled:opacity-40"
          onClick={onPush} disabled={pushing || loading}
          title="git push"
        >
          <ArrowUp className={cn('w-3 h-3', pushing && 'animate-bounce')} />
          Push
        </Button>
        {/* 새로고침 */}
        <Button variant="ghost" size="icon" className="h-6 w-6 text-muted-foreground hover:text-foreground"
          onClick={onRefresh} disabled={loading}>
          <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
        </Button>
      </div>
    </div>
  )
}
