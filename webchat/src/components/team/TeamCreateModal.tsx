'use client'

import { useState, useEffect } from 'react'
import { useTeam } from '@/contexts/TeamContext'
import { useServer } from '@/contexts/ServerContext'
import { useProject } from '@/contexts/ProjectContext'
import { useLocale } from '@/contexts/LocaleContext'
import { Button } from '@/components/ui/button'
import { Input }  from '@/components/ui/input'
import { Label }  from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Users, LayoutTemplate, Settings2, Coins, AlertCircle, Loader2 } from 'lucide-react'
import TemplateManagerModal from './TemplateManagerModal'

interface Props {
  open:    boolean
  onClose: () => void
}

export default function TeamCreateModal({ open, onClose }: Props) {
  const { refreshTeams, setActiveTeam } = useTeam()
  const { getApiUrl }     = useServer()
  const { activeProject } = useProject()
  const { t } = useLocale()
  const [name,     setName]     = useState('')
  const [template, setTemplate] = useState<string | null>('leader-worker')
  const [workers,  setWorkers]  = useState('2')
  const [budget,   setBudget]   = useState('')
  const [busy,     setBusy]     = useState(false)
  const [error,    setError]    = useState('')
  const [tmplOpen, setTmplOpen] = useState(false)
  const [customTemplates, setCustomTemplates] = useState<string[]>([])

  const BUILTIN = ['leader-worker', 'leader-worker-reviewer', 'parallel-workers']

  async function fetchTemplates() {
    try {
      const res = await fetch(getApiUrl('/api/template/list'))
      if (!res.ok) return
      const data = await res.json() as { templates: Array<{ name: string }> }
      const custom = data.templates
        .map((tmpl) => tmpl.name)
        .filter((n) => !BUILTIN.includes(n))
      setCustomTemplates(custom)
    } catch { /* ignore */ }
  }

  useEffect(() => {
    if (open) fetchTemplates()
  }, [open])

  useEffect(() => {
    if (!tmplOpen && open) fetchTemplates()
  }, [tmplOpen])

  async function handleCreate() {
    if (!name.trim()) return
    setBusy(true)
    setError('')
    try {
      const res = await fetch(getApiUrl('/api/team/create'), {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({
          name:        name.trim(),
          template:    template || undefined,
          workers,
          budget:      budget || undefined,
          projectPath: activeProject?.path,
        }),
      })
      if (!res.ok) {
        const err = await res.json() as { error?: string }
        setError(err.error || t.team.createFailed)
        return
      }
      await refreshTeams()
      setActiveTeam(name.trim())
      setName('')
      onClose()
    } catch (e) {
      setError(String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <TemplateManagerModal open={tmplOpen} onClose={() => setTmplOpen(false)} />
      <Dialog open={open} onOpenChange={(v) => { if (!v) onClose() }}>
        <DialogContent className="w-[420px] max-w-[95vw] p-0 overflow-hidden">

          {/* ── 헤더 ── */}
          <DialogHeader className="px-5 pt-5 pb-4 border-b border-border/60 bg-muted/30">
            <DialogTitle className="flex items-center gap-2 text-sm font-semibold">
              <span className="flex items-center justify-center w-6 h-6 rounded-md bg-primary/10 text-primary">
                <Users className="w-3.5 h-3.5" />
              </span>
              {t.team.createTitle}
            </DialogTitle>
          </DialogHeader>

          {/* ── 폼 ── */}
          <div className="flex flex-col gap-4 px-5 py-4">

            {/* 팀 이름 */}
            <div className="flex flex-col gap-1.5">
              <Label className="text-xs font-medium text-muted-foreground">
                {t.team.createNameLabel}
              </Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="my-team"
                className="h-8 text-xs placeholder:text-[11px]"
                onKeyDown={(e) => { if (e.key === 'Enter' && name.trim()) handleCreate() }}
              />
            </div>

            {/* 템플릿 */}
            <div className="flex flex-col gap-1.5">
              <div className="flex items-center justify-between">
                <Label className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                  <LayoutTemplate className="w-3 h-3" />
                  {t.team.createTemplateLabel}
                </Label>
                <button
                  type="button"
                  onClick={() => setTmplOpen(true)}
                  className="flex items-center gap-0.5 text-[11px] text-muted-foreground hover:text-foreground transition-colors"
                >
                  <Settings2 className="w-3 h-3" />
                  {t.team.createTemplateManage}
                </button>
              </div>
              <Select value={template} onValueChange={(v) => setTemplate(v)}>
                <SelectTrigger className="w-full h-8 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel className="text-[11px]">기본 템플릿</SelectLabel>
                    <SelectItem value="leader-worker" className="text-xs">leader-worker</SelectItem>
                    <SelectItem value="leader-worker-reviewer" className="text-xs">leader-worker-reviewer</SelectItem>
                    <SelectItem value="parallel-workers" className="text-xs">parallel-workers</SelectItem>
                  </SelectGroup>
                  {customTemplates.length > 0 && (
                    <>
                      <SelectSeparator />
                      <SelectGroup>
                        <SelectLabel className="text-[11px]">커스텀 템플릿</SelectLabel>
                        {customTemplates.map((tmplName) => (
                          <SelectItem key={tmplName} value={tmplName} className="text-xs">{tmplName}</SelectItem>
                        ))}
                      </SelectGroup>
                    </>
                  )}
                </SelectContent>
              </Select>
            </div>

            {/* 워커 수 + 토큰 예산 — 2열 */}
            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs font-medium text-muted-foreground">
                  {t.team.createWorkersLabel}
                </Label>
                <Input
                  type="number"
                  value={workers}
                  onChange={(e) => setWorkers(e.target.value)}
                  min={1}
                  max={10}
                  className="h-8 text-[11px] placeholder:text-[11px]"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                  <Coins className="w-3 h-3" />
                  {t.team.createBudgetLabel}
                </Label>
                <Input
                  type="number"
                  value={budget}
                  onChange={(e) => setBudget(e.target.value)}
                  placeholder={t.team.createBudgetHint}
                  className="h-8 text-xs placeholder:text-[11px]"
                />
              </div>
            </div>

            {/* 오류 메시지 */}
            {error && (
              <div className="flex items-center gap-1.5 text-xs text-destructive bg-destructive/10 border border-destructive/20 rounded-md px-3 py-2">
                <AlertCircle className="w-3.5 h-3.5 shrink-0" />
                {error}
              </div>
            )}
          </div>

          {/* ── 푸터 ── */}
          <div className="flex items-center justify-end gap-2 px-5 py-3 border-t border-border/60 bg-muted/20">
            <Button variant="ghost" size="sm" onClick={onClose} disabled={busy} className="h-7 text-xs px-3">
              {t.team.cancel}
            </Button>
            <Button
              size="sm"
              onClick={handleCreate}
              disabled={busy || !name.trim()}
              className="h-7 text-xs px-4 gap-1.5"
            >
              {busy && <Loader2 className="w-3 h-3 animate-spin" />}
              {busy ? t.team.creating : t.team.createBtn}
            </Button>
          </div>

        </DialogContent>
      </Dialog>
    </>
  )
}
