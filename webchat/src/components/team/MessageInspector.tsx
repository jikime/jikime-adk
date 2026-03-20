'use client'

import { useState } from 'react'
import { useTeam } from '@/contexts/TeamContext'
import { useServer } from '@/contexts/ServerContext'
import { Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export default function MessageInspector() {
  const { activeTeam, agents, sendMessage } = useTeam()
  const { getApiUrl } = useServer()
  const [to, setTo]       = useState<string | null>('')
  const [body, setBody]   = useState('')
  const [msgs, setMsgs]   = useState<unknown[]>([])
  const [loading, setLoading] = useState(false)

  if (!activeTeam) return null

  async function peekInbox(agentID: string) {
    setLoading(true)
    try {
      const res = await fetch(getApiUrl(`/api/team/${activeTeam}/inbox/peek?agent=${agentID}`))
      if (res.ok) {
        const data = await res.json() as { messages?: unknown[] }
        setMsgs(data.messages || [])
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleSend() {
    if (!to || !body || !activeTeam) return
    await sendMessage(activeTeam, to, body)
    setBody('')
  }

  return (
    <div className="flex flex-col gap-2 p-3 bg-slate-50 dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800">
      <div className="text-base font-semibold text-slate-500 dark:text-slate-400">메시지 인스펙터</div>

      {/* Agent inbox peek */}
      <div className="flex gap-1 flex-wrap">
        {agents.map((a) => (
          <button
            key={a.id}
            onClick={() => peekInbox(a.id)}
            className="text-[10px] px-2 py-0.5 bg-slate-200 dark:bg-slate-700 rounded hover:bg-slate-300 dark:hover:bg-slate-600"
          >
            {a.id}
          </button>
        ))}
      </div>

      {loading && <div className="text-base text-slate-400">로딩 중…</div>}

      {msgs.length > 0 && (
        <div className="max-h-32 overflow-y-auto text-base space-y-1">
          {msgs.map((m, i) => {
            const msg = m as Record<string, string>
            return (
              <div key={i} className="bg-white dark:bg-slate-800 rounded p-1.5 border border-slate-100 dark:border-slate-700">
                <div className="text-slate-400">{msg['sent_at']} from: <span className="text-blue-500">{msg['from']}</span></div>
                <div className="text-slate-700 dark:text-slate-300">{msg['body']}</div>
              </div>
            )
          })}
        </div>
      )}

      {/* Send message */}
      <div className="flex gap-2 items-center">
        <Select value={to} onValueChange={(v) => setTo(v)}>
          <SelectTrigger className="h-7 text-xs w-36">
            <SelectValue placeholder="수신자" />
          </SelectTrigger>
          <SelectContent>
            {agents.map((a) => <SelectItem key={a.id} value={a.id}>{a.id}</SelectItem>)}
            <SelectItem value="leader">leader</SelectItem>
            <SelectItem value="broadcast">전체 브로드캐스트</SelectItem>
          </SelectContent>
        </Select>
        <Input
          className="text-base h-7 flex-1"
          placeholder="메시지 입력…"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') handleSend() }}
        />
        <Button size="icon" variant="ghost" className="h-7 w-7" onClick={handleSend}>
          <Send className="w-3 h-3" />
        </Button>
      </div>
    </div>
  )
}
