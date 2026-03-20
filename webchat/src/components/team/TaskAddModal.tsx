'use client'

import { useState } from 'react'
import { useTeam } from '@/contexts/TeamContext'
import { useLocale } from '@/contexts/LocaleContext'
import { Button } from '@/components/ui/button'
import { Input }  from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface Props {
  open:    boolean
  onClose: () => void
}

export default function TaskAddModal({ open, onClose }: Props) {
  const { activeTeam, createTask } = useTeam()
  const { t } = useLocale()
  const [title, setTitle] = useState('')
  const [desc,  setDesc]  = useState('')
  const [busy,  setBusy]  = useState(false)

  async function handleSubmit() {
    if (!title.trim() || !activeTeam) return
    setBusy(true)
    try {
      await createTask(activeTeam, title.trim(), desc.trim() || undefined)
      setTitle('')
      setDesc('')
      onClose()
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open={open && !!activeTeam} onOpenChange={(v) => { if (!v) onClose() }}>
      <DialogContent className="w-96 max-w-[95vw] flex flex-col gap-4">
        <DialogHeader>
          <DialogTitle>{t.team.addTaskTitle}</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-3">
          <div>
            <label className="text-base font-medium text-muted-foreground mb-1 block">
              {t.team.taskTitleLabel}
            </label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t.team.taskTitlePlaceholder}
              onKeyDown={(e) => { if (e.key === 'Enter') handleSubmit() }}
            />
          </div>
          <div>
            <label className="text-base font-medium text-muted-foreground mb-1 block">
              {t.team.taskDescLabel}
            </label>
            <textarea
              value={desc}
              onChange={(e) => setDesc(e.target.value)}
              placeholder={t.team.taskDescPlaceholder}
              rows={3}
              className="w-full border border-border rounded px-3 py-2 text-lg bg-muted text-foreground resize-none placeholder:text-muted-foreground outline-none"
            />
          </div>
        </div>

        <div className="flex gap-2 justify-end">
          <Button variant="outline" onClick={onClose} disabled={busy}>{t.team.cancel}</Button>
          <Button onClick={handleSubmit} disabled={busy || !title.trim()}>
            {busy ? t.team.adding : t.team.addBtn}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
