'use client'

import { useState } from 'react'
import { useTeam } from '@/contexts/TeamContext'
import { useServer } from '@/contexts/ServerContext'
import { useProject } from '@/contexts/ProjectContext'
import { useLocale } from '@/contexts/LocaleContext'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Play, GitBranch } from 'lucide-react'

interface Props {
  open:    boolean
  onClose: () => void
}

export default function TeamServeModal({ open, onClose }: Props) {
  const { activeTeam, refreshTeam } = useTeam()
  const { getApiUrl }   = useServer()
  const { activeProject } = useProject()
  const { t } = useLocale()

  const [goal,     setGoal]     = useState('')
  const [worktree, setWorktree] = useState(false)
  const [busy,     setBusy]     = useState(false)
  const [output,   setOutput]   = useState('')
  const [error,    setError]    = useState('')

  function handleClose() {
    if (busy) return
    onClose()
    setOutput('')
    setGoal('')
    setError('')
  }

  async function handleRun() {
    if (!activeTeam || !goal.trim()) return
    setBusy(true)
    setError('')
    setOutput('')
    try {
      const res = await fetch(getApiUrl(`/api/team/${activeTeam}/run`), {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({
          goal:        goal.trim(),
          worktree:    worktree ? '1' : undefined,
          projectPath: activeProject?.path,
        }),
      })
      const data = await res.json() as { ok?: boolean; output?: string; error?: string }
      if (!res.ok) {
        setError(data.error || t.team.runFailed)
        setOutput(data.output || '')
        return
      }
      setOutput(data.output || '')
      await refreshTeam()
      setTimeout(() => {
        onClose()
        setOutput('')
        setGoal('')
      }, 2000)
    } catch (e) {
      setError(String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleClose() }}>
      <DialogContent className="w-[460px] max-w-[95vw] flex flex-col gap-4">

        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-sm font-semibold">
            <Play className="w-3.5 h-3.5 text-green-400" />
            {t.team.runTitle}
            {activeTeam && (
              <span className="text-xs font-normal text-muted-foreground">
                — {activeTeam}
              </span>
            )}
          </DialogTitle>
        </DialogHeader>

        {/* Goal */}
        <div>
          <label className="text-[11px] font-medium text-muted-foreground mb-1.5 block uppercase tracking-wider">
            {t.team.goalLabel} <span className="text-red-400">*</span>
          </label>
          <textarea
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            placeholder={t.team.goalPlaceholder}
            rows={5}
            className="w-full bg-muted text-foreground border border-border rounded-lg px-3 py-2.5 text-sm outline-none resize-none placeholder:text-xs placeholder:text-muted-foreground leading-relaxed"
          />
        </div>

        {/* Worktree option */}
        <label className="flex items-center gap-2.5 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={worktree}
            onChange={(e) => setWorktree(e.target.checked)}
            className="rounded accent-green-500"
          />
          <GitBranch className="w-3.5 h-3.5 text-muted-foreground" />
          <span className="text-xs text-muted-foreground">{t.team.worktreeLabel}</span>
        </label>

        {/* Output log */}
        {output && (
          <pre className="bg-black rounded-lg px-3 py-2.5 text-[11px] text-green-400 max-h-36 overflow-y-auto font-mono leading-relaxed whitespace-pre-wrap">
            {output}
          </pre>
        )}

        {/* Error */}
        {error && (
          <div className="text-[11px] text-red-400 bg-red-950/30 border border-red-900/40 rounded-lg px-3 py-2 leading-relaxed">
            {error}
          </div>
        )}

        <div className="flex gap-2 justify-end">
          <Button variant="outline" size="sm" onClick={handleClose} disabled={busy} className="h-7 text-xs px-3">
            {t.team.cancel}
          </Button>
          <Button
            size="sm"
            onClick={handleRun}
            disabled={busy || !goal.trim()}
            className="h-7 text-xs px-4 bg-green-600 hover:bg-green-500 text-white font-medium"
          >
            {busy ? t.team.running : t.team.runBtn}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
