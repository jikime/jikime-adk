'use client'

import { useState, useEffect, useCallback } from 'react'
import { Plus, FolderGit2, Trash2, Bot, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import ProjectForm from '@/components/ProjectForm'
import ChatView from '@/components/ChatView'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
} from '@/components/ui/dialog'

interface Project {
  id: string
  name: string
  repo: string
  cwd: string
  port: number
  pid: number | null
  status: 'running' | 'stopped'
}

export default function Home() {
  const [projects, setProjects] = useState<Project[]>([])
  const [selected, setSelected] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [loading, setLoading] = useState(true)

  const loadProjects = useCallback(async () => {
    try {
      const res = await fetch('/api/projects')
      const data = await res.json() as Project[]
      setProjects(data)
      if (data.length > 0 && !selected) setSelected(data[0].id)
    } finally {
      setLoading(false)
    }
  }, [selected])

  useEffect(() => { loadProjects() }, [])  // eslint-disable-line

  const handleCreated = () => { loadProjects() }

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm('프로젝트를 삭제할까요?')) return
    await fetch(`/api/projects?id=${id}`, { method: 'DELETE' })
    if (selected === id) setSelected(null)
    loadProjects()
  }

  const activeProject = projects.find(p => p.id === selected) ?? null

  return (
    <div className="flex h-screen bg-zinc-950 text-zinc-100">
      {/* 사이드바 */}
      <aside className="w-64 shrink-0 flex flex-col bg-zinc-900 border-r border-zinc-800">
        <div className="px-4 py-4 border-b border-zinc-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="w-5 h-5 text-blue-400" />
            <span className="font-bold text-blue-400 text-sm">webchat-git</span>
          </div>
          <button
            onClick={() => setShowForm(true)}
            className="w-7 h-7 rounded-lg bg-blue-600 hover:bg-blue-500 flex items-center justify-center transition-colors"
            title="새 프로젝트"
          >
            <Plus className="w-3.5 h-3.5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto py-2">
          {loading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="w-4 h-4 animate-spin text-zinc-600" />
            </div>
          )}

          {!loading && projects.length === 0 && (
            <div className="px-4 py-8 text-center">
              <FolderGit2 className="w-8 h-8 text-zinc-700 mx-auto mb-2" />
              <p className="text-xs text-zinc-600">프로젝트가 없어요</p>
              <button onClick={() => setShowForm(true)} className="mt-3 text-xs text-blue-500 hover:text-blue-400">
                + 새 프로젝트 만들기
              </button>
            </div>
          )}

          {projects.map(p => (
            <div
              key={p.id}
              onClick={() => setSelected(p.id)}
              className={cn(
                'w-full px-3 py-2.5 flex items-start gap-2.5 hover:bg-zinc-800 transition-colors group cursor-pointer',
                selected === p.id && 'bg-zinc-800'
              )}
            >
              <div className={cn(
                'w-1.5 h-1.5 rounded-full mt-1.5 shrink-0',
                p.status === 'running' ? 'bg-emerald-400' : 'bg-zinc-600'
              )} />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-zinc-200 truncate font-medium">{p.name}</p>
                <p className="text-[11px] text-zinc-500 font-mono truncate">{p.repo}</p>
              </div>
              <button
                onClick={(e) => handleDelete(p.id, e)}
                className="opacity-0 group-hover:opacity-100 text-zinc-600 hover:text-red-400 transition-all shrink-0"
              >
                <Trash2 className="w-3 h-3" />
              </button>
            </div>
          ))}
        </div>

        <div className="px-4 py-3 border-t border-zinc-800">
          <p className="text-[10px] text-zinc-700 font-mono">jikime serve harness</p>
        </div>
      </aside>

      {/* 메인 */}
      <main className="flex-1 min-w-0">
        {activeProject ? (
          <ChatView
            key={activeProject.id}
            project={activeProject}
            onProjectUpdate={(updated) =>
              setProjects(prev => prev.map(p => p.id === updated.id ? updated : p))
            }
          />
        ) : (
          <div className="flex flex-col items-center justify-center h-full gap-4 text-center">
            <FolderGit2 className="w-16 h-16 text-zinc-800" />
            <div>
              <p className="text-zinc-400 font-medium">프로젝트를 선택하세요</p>
              <p className="text-sm text-zinc-600 mt-1">사이드바에서 프로젝트를 선택하거나 새로 만들어보세요</p>
            </div>
            <button
              onClick={() => setShowForm(true)}
              className="flex items-center gap-2 px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-sm text-white transition-colors"
            >
              <Plus className="w-4 h-4" /> 새 프로젝트
            </button>
          </div>
        )}
      </main>

      {/* 새 프로젝트 Dialog */}
      <Dialog open={showForm} onOpenChange={setShowForm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>새 프로젝트</DialogTitle>
            <DialogDescription>
              GitHub 저장소와 jikime serve 연동 설정을 입력해 주세요
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <ProjectForm
              onCreated={handleCreated}
              onClose={() => setShowForm(false)}
            />
          </DialogBody>
        </DialogContent>
      </Dialog>
    </div>
  )
}
