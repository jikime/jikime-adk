'use client'

import { useState, useRef, useEffect } from 'react'
import { FolderGit2, KeyRound, Server, FolderOpen, Link2, Loader2, Plus, Check } from 'lucide-react'
import { cn } from '@/lib/utils'

interface GithubRepo {
  full_name: string   // "owner/repo"
  name: string
  private: boolean
  description: string | null
}

interface Props {
  onCreated: () => void
  onClose: () => void
}

export default function ProjectForm({ onCreated, onClose }: Props) {
  const [form, setForm] = useState({ name: '', repo: '', token: '', cwd: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // 저장소 연결
  const [fetching, setFetching] = useState(false)
  const [fetchError, setFetchError] = useState('')
  const [repos, setRepos] = useState<GithubRepo[]>([])
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [repoQuery, setRepoQuery] = useState('')
  const dropdownRef = useRef<HTMLDivElement>(null)

  // 드롭다운 외부 클릭 시 닫기
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const field = (key: keyof typeof form) => ({
    value: form[key],
    onChange: (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm(f => ({ ...f, [key]: e.target.value })),
  })

  // GitHub 저장소 목록 가져오기
  const fetchRepos = async () => {
    if (!form.token) {
      setFetchError('GitHub 토큰을 먼저 입력해 주세요')
      return
    }
    setFetching(true)
    setFetchError('')
    try {
      const res = await fetch('https://api.github.com/user/repos?per_page=100&sort=updated&affiliation=owner,collaborator', {
        headers: {
          Authorization: `Bearer ${form.token}`,
          Accept: 'application/vnd.github+json',
        },
      })
      if (!res.ok) {
        setFetchError('토큰이 유효하지 않거나 권한이 없어요')
        return
      }
      const data = await res.json() as GithubRepo[]
      setRepos(data)
      setDropdownOpen(true)
      setRepoQuery('')
    } catch {
      setFetchError('GitHub 연결에 실패했어요')
    } finally {
      setFetching(false)
    }
  }

  const selectRepo = (fullName: string) => {
    setForm(f => ({ ...f, repo: fullName }))
    setDropdownOpen(false)
  }

  const filteredRepos = repos.filter(r =>
    r.full_name.toLowerCase().includes(repoQuery.toLowerCase())
  )

  const submit = async (e: React.SyntheticEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await fetch('/api/projects', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      if (!res.ok) {
        const data = await res.json() as { error: string }
        setError(data.error)
        return
      }
      onCreated()
      onClose()
    } catch {
      setError('네트워크 오류')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      {/* 프로젝트 이름 */}
      <Field icon={<Server className="w-4 h-4" />} label="프로젝트 이름">
        <input {...field('name')} placeholder="My Project" required className={inputCls} />
      </Field>

      {/* GitHub 토큰 */}
      <Field icon={<KeyRound className="w-4 h-4" />} label="GitHub 토큰">
        <input {...field('token')} type="password" placeholder="ghp_..." required className={inputCls} />
      </Field>

      {/* GitHub 저장소 */}
      <Field icon={<FolderGit2 className="w-4 h-4" />} label="GitHub 저장소">
        <div className="flex gap-2">
          <input
            {...field('repo')}
            placeholder="owner/repo"
            required
            className={cn(inputCls, 'flex-1')}
          />
          <button
            type="button"
            onClick={fetchRepos}
            disabled={fetching}
            title="토큰으로 저장소 목록 불러오기"
            className={cn(
              'shrink-0 flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium transition-colors',
              'bg-zinc-700 hover:bg-zinc-600 text-zinc-200 border border-zinc-600',
              'disabled:opacity-50'
            )}
          >
            {fetching
              ? <Loader2 className="w-3.5 h-3.5 animate-spin" />
              : <Link2 className="w-3.5 h-3.5" />
            }
            연결
          </button>
        </div>

        {/* 에러 */}
        {fetchError && <p className="text-[11px] text-red-400 mt-1">{fetchError}</p>}

        {/* 저장소 드롭다운 */}
        {dropdownOpen && repos.length > 0 && (
          <div
            ref={dropdownRef}
            className="mt-1.5 bg-zinc-800 border border-zinc-700 rounded-xl shadow-xl overflow-hidden z-10"
          >
            {/* 검색 */}
            <div className="px-3 py-2 border-b border-zinc-700">
              <input
                autoFocus
                value={repoQuery}
                onChange={e => setRepoQuery(e.target.value)}
                placeholder="저장소 검색..."
                className="w-full bg-transparent text-sm text-zinc-100 placeholder:text-zinc-500 outline-none"
              />
            </div>

            {/* 목록 */}
            <ul className="max-h-52 overflow-y-auto">
              {filteredRepos.length === 0 && (
                <li className="px-3 py-3 text-xs text-zinc-500 text-center">검색 결과 없음</li>
              )}
              {filteredRepos.map(r => (
                <li key={r.full_name}>
                  <button
                    type="button"
                    onClick={() => selectRepo(r.full_name)}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-3 py-2.5 text-left hover:bg-zinc-700 transition-colors',
                      form.repo === r.full_name && 'bg-zinc-700'
                    )}
                  >
                    <FolderGit2 className="w-3.5 h-3.5 text-zinc-500 shrink-0" />
                    <span className="flex-1 min-w-0">
                      <span className="text-sm text-zinc-200 truncate block">{r.full_name}</span>
                      {r.description && (
                        <span className="text-[11px] text-zinc-500 truncate block">{r.description}</span>
                      )}
                    </span>
                    <span className="text-[10px] text-zinc-600 shrink-0">{r.private ? '🔒' : '🌐'}</span>
                    {form.repo === r.full_name && <Check className="w-3.5 h-3.5 text-blue-400 shrink-0" />}
                  </button>
                </li>
              ))}
              {/* 직접 입력 옵션 */}
              <li className="border-t border-zinc-700">
                <button
                  type="button"
                  onClick={() => setDropdownOpen(false)}
                  className="w-full flex items-center gap-2 px-3 py-2.5 text-left hover:bg-zinc-700 transition-colors text-zinc-400"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span className="text-xs">직접 입력</span>
                </button>
              </li>
            </ul>
          </div>
        )}
      </Field>

      {/* 로컬 작업 경로 */}
      <Field icon={<FolderOpen className="w-4 h-4" />} label="로컬 작업 경로">
        <input {...field('cwd')} placeholder="/Users/me/projects/my-project" required className={inputCls} />
        <p className="text-[11px] text-zinc-500 mt-1">WORKFLOW.md가 이 경로에 생성돼요</p>
      </Field>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <div className="flex gap-3 pt-2">
        <button
          type="button" onClick={onClose}
          className="flex-1 py-2 rounded-xl border border-zinc-700 text-zinc-400 text-sm hover:bg-zinc-800 transition-colors"
        >
          취소
        </button>
        <button
          type="submit" disabled={loading}
          className="flex-1 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium disabled:opacity-50 transition-colors"
        >
          {loading ? '생성 중...' : '프로젝트 생성'}
        </button>
      </div>
    </form>
  )
}

function Field({ icon, label, children }: { icon: React.ReactNode; label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="flex items-center gap-1.5 text-xs text-zinc-400 mb-1.5">
        <span className="text-zinc-500">{icon}</span>
        {label}
      </label>
      {children}
    </div>
  )
}

const inputCls = cn(
  'w-full px-3 py-2 rounded-lg bg-zinc-800 border border-zinc-700',
  'text-sm text-zinc-100 placeholder:text-zinc-600',
  'outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 transition-colors'
)
