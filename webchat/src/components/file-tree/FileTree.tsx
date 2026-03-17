'use client'

import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import {
  ChevronRight, ChevronDown, File, Folder, FolderOpen,
  RefreshCw, Search, X, Save, Check,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import MonacoReactEditor, { type OnMount } from '@monaco-editor/react'

// ── 타입 ──────────────────────────────────────────────────────────
interface FileNode {
  name: string
  path: string
  type: 'file' | 'directory'
  size?: number
  modified?: number
  children?: FileNode[]
}

// ── 유틸 ──────────────────────────────────────────────────────────
const EXT_COLOR: Record<string, string> = {
  ts: 'text-blue-400', tsx: 'text-blue-300',
  js: 'text-yellow-400', jsx: 'text-yellow-300',
  py: 'text-green-400', rs: 'text-orange-400',
  go: 'text-cyan-400',  md: 'text-zinc-300',
  json: 'text-yellow-300', css: 'text-pink-400',
  scss: 'text-pink-300', html: 'text-orange-300',
  sh: 'text-green-300', yaml: 'text-red-300',
  yml: 'text-red-300',  toml: 'text-orange-200',
  env: 'text-zinc-400', sql: 'text-blue-200',
}

function extColor(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() ?? ''
  return EXT_COLOR[ext] ?? 'text-zinc-400'
}

const EXT_LANG: Record<string, string> = {
  ts: 'typescript', tsx: 'typescript', js: 'javascript', jsx: 'javascript',
  py: 'python', rs: 'rust', go: 'go', md: 'markdown',
  json: 'json', css: 'css', scss: 'scss', html: 'html',
  sh: 'shell', bash: 'shell', yaml: 'yaml', yml: 'yaml',
  toml: 'toml', sql: 'sql', xml: 'xml', graphql: 'graphql',
  prisma: 'prisma', dockerfile: 'dockerfile',
}

function extLang(name: string): string {
  const lower = name.toLowerCase()
  if (lower === 'dockerfile') return 'dockerfile'
  const ext = lower.split('.').pop() ?? ''
  return EXT_LANG[ext] ?? 'plaintext'
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}K`
  return `${(bytes / (1024 * 1024)).toFixed(1)}M`
}

// 트리에서 모든 파일 노드 평탄화 (검색용)
function flattenFiles(nodes: FileNode[]): FileNode[] {
  const out: FileNode[] = []
  for (const n of nodes) {
    if (n.type === 'file') out.push(n)
    if (n.children) out.push(...flattenFiles(n.children))
  }
  return out
}

// ── 파일 트리 아이템 ──────────────────────────────────────────────
function FileNodeItem({
  node, depth = 0, selectedPath, onSelect,
}: {
  node: FileNode
  depth?: number
  selectedPath: string | null
  onSelect: (node: FileNode) => void
}) {
  const [open, setOpen] = useState(depth < 1)
  const isDir  = node.type === 'directory'
  const indent = depth * 14

  return (
    <div>
      <button
        className={cn(
          'flex items-center gap-1.5 w-full text-left py-0.5 rounded text-xs transition-colors',
          isDir ? 'text-zinc-300 hover:text-zinc-100 hover:bg-zinc-800' : 'hover:bg-zinc-800',
          !isDir && selectedPath === node.path
            ? 'bg-zinc-800 text-zinc-100'
            : !isDir && 'text-zinc-400 hover:text-zinc-200'
        )}
        style={{ paddingLeft: `${indent + 8}px`, paddingRight: '8px' }}
        onClick={() => isDir ? setOpen(v => !v) : onSelect(node)}
      >
        {isDir ? (
          <>
            {open
              ? <ChevronDown className="w-3 h-3 shrink-0 text-zinc-500" />
              : <ChevronRight className="w-3 h-3 shrink-0 text-zinc-500" />}
            {open
              ? <FolderOpen className="w-3.5 h-3.5 shrink-0 text-blue-400" />
              : <Folder     className="w-3.5 h-3.5 shrink-0 text-blue-400" />}
          </>
        ) : (
          <>
            <span className="w-3 h-3 shrink-0" />
            <File className={cn('w-3.5 h-3.5 shrink-0', extColor(node.name))} />
          </>
        )}
        <span className="truncate flex-1">{node.name}</span>
        {!isDir && node.size !== undefined && (
          <span className="text-zinc-600 text-[10px] shrink-0">{formatSize(node.size)}</span>
        )}
      </button>

      {isDir && open && node.children?.map(child => (
        <FileNodeItem
          key={child.path}
          node={child}
          depth={depth + 1}
          selectedPath={selectedPath}
          onSelect={onSelect}
        />
      ))}
    </div>
  )
}

// ── 검색 결과 아이템 ──────────────────────────────────────────────
function SearchResultItem({
  node, selectedPath, onSelect, rootPath,
}: {
  node: FileNode
  selectedPath: string | null
  onSelect: (node: FileNode) => void
  rootPath: string
}) {
  const rel = node.path.replace(rootPath, '').replace(/^\//, '')
  return (
    <button
      onClick={() => onSelect(node)}
      className={cn(
        'flex flex-col w-full text-left px-2 py-1.5 rounded text-xs transition-colors hover:bg-zinc-800',
        selectedPath === node.path ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400'
      )}
    >
      <div className="flex items-center gap-1.5">
        <File className={cn('w-3 h-3 shrink-0', extColor(node.name))} />
        <span className="font-medium text-zinc-200">{node.name}</span>
      </div>
      <span className="text-zinc-600 text-[10px] font-mono truncate pl-4">{rel}</span>
    </button>
  )
}

// ── Monaco 에디터 ─────────────────────────────────────────────────
function MonacoEditor({ path, content, getApiUrl }: { path: string; content: string; getApiUrl: (p: string) => string }) {
  const name = path.split('/').pop() ?? path
  const lang = extLang(name)

  const [value, setValue]   = useState(content)
  const [dirty, setDirty]   = useState(false)
  const [saved, setSaved]   = useState(false)
  const [saving, setSaving] = useState(false)
  // handleSave의 최신 참조를 저장 (addCommand에서 항상 최신 함수 호출)
  const saveRef = useRef<() => void>(() => {})

  // path/content가 바뀌면 내용 초기화
  useEffect(() => {
    setValue(content)
    setDirty(false)
    setSaved(false)
  }, [path, content])

  const handleSave = useCallback(async () => {
    if (!dirty || saving) return
    setSaving(true)
    try {
      const res = await fetch(getApiUrl('/api/ws/file'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, content: value }),
      })
      if (res.ok) {
        setDirty(false)
        setSaved(true)
        setTimeout(() => setSaved(false), 2000)
      }
    } catch { /* */ } finally {
      setSaving(false)
    }
  }, [dirty, saving, path, value])

  // ref 항상 최신 유지
  useEffect(() => { saveRef.current = handleSave }, [handleSave])

  const handleMount: OnMount = (editor, monaco) => {
    editor.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
      () => { saveRef.current() }
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* 헤더 */}
      <div className="flex items-center justify-between px-3 py-1.5 bg-zinc-900 border-b border-zinc-800 shrink-0">
        <div className="flex items-center gap-1.5 min-w-0">
          <File className={cn('w-3.5 h-3.5 shrink-0', extColor(name))} />
          <span className="text-xs font-mono truncate min-w-0" title={path}>
            <span className="text-zinc-500">{path.slice(0, path.lastIndexOf('/') + 1)}</span>
            <span className="text-zinc-200">{name}</span>
          </span>
          <span className="text-zinc-600 text-[10px] shrink-0 ml-1">{lang}</span>
          {dirty && (
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0" title="저장되지 않은 변경사항" />
          )}
        </div>
        <button
          onClick={handleSave}
          disabled={!dirty || saving}
          className={cn(
            'flex items-center gap-1 text-xs transition-colors shrink-0 px-2 py-0.5 rounded',
            dirty && !saving
              ? 'text-zinc-200 hover:bg-zinc-700 cursor-pointer'
              : 'text-zinc-600 cursor-default'
          )}
        >
          {saved
            ? <><Check className="w-3 h-3 text-emerald-400" /><span className="text-emerald-400">저장됨</span></>
            : saving
              ? <><RefreshCw className="w-3 h-3 animate-spin" /><span>저장 중...</span></>
              : <><Save className="w-3 h-3" /><span>저장</span></>
          }
        </button>
      </div>

      {/* Monaco Editor */}
      <div className="flex-1 min-h-0" onKeyDown={(e) => {
        if ((e.metaKey || e.ctrlKey) && e.key === 's') {
          e.preventDefault()
          handleSave()
        }
      }}>
        <MonacoReactEditor
          height="100%"
          language={lang}
          value={value}
          theme="vs-dark"
          onMount={handleMount}
          onChange={(val: string | undefined) => {
            setValue(val ?? '')
            setDirty(true)
            setSaved(false)
          }}
          options={{
            fontSize: 13,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            lineHeight: 1.5,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            wordWrap: 'off',
            automaticLayout: true,
            tabSize: 2,
            renderWhitespace: 'selection',
            smoothScrolling: true,
            cursorBlinking: 'smooth',
            padding: { top: 8, bottom: 8 },
          }}
        />
      </div>
    </div>
  )
}

// ── 메인 컴포넌트 ─────────────────────────────────────────────────
export default function FileTree() {
  const { activeProject } = useProject()
  const { getApiUrl }     = useServer()

  const [tree, setTree]           = useState<FileNode[]>([])
  const [loading, setLoading]     = useState(false)
  const [query, setQuery]         = useState('')
  const [selectedNode, setSelectedNode] = useState<FileNode | null>(null)
  const [fileContent, setFileContent]   = useState<string | null>(null)
  const [fileLoading, setFileLoading]   = useState(false)

  // 트리 로드
  const loadTree = useCallback(async () => {
    if (!activeProject?.path) return
    setLoading(true)
    try {
      const res = await fetch(getApiUrl(`/api/ws/files?path=${encodeURIComponent(activeProject.path)}`))
      if (res.ok) setTree(await res.json())
    } catch { /* */ } finally {
      setLoading(false)
    }
  }, [activeProject?.path, getApiUrl])

  useEffect(() => {
    setSelectedNode(null)
    setFileContent(null)
    setQuery('')
    loadTree()
  }, [loadTree])

  // 파일 선택 → 내용 로드
  const handleSelect = useCallback(async (node: FileNode) => {
    if (node.type !== 'file') return
    setSelectedNode(node)
    setFileLoading(true)
    setFileContent(null)
    try {
      const res = await fetch(getApiUrl(`/api/ws/file?path=${encodeURIComponent(node.path)}`))
      if (res.ok) {
        const data = await res.json()
        setFileContent(data.content as string)
      } else {
        setFileContent('파일을 읽을 수 없습니다.')
      }
    } catch {
      setFileContent('읽기 오류가 발생했습니다.')
    } finally {
      setFileLoading(false)
    }
  }, [getApiUrl])

  // 검색 필터
  const searchResults = useMemo(() => {
    if (!query.trim()) return null
    const q = query.toLowerCase()
    return flattenFiles(tree).filter(f => f.name.toLowerCase().includes(q)).slice(0, 50)
  }, [query, tree])

  const rootPath = activeProject?.path ?? ''

  return (
    <div className="flex h-full bg-zinc-950 rounded-lg overflow-hidden border border-zinc-800">

      {/* ── 좌측: 파일 트리 ──────────────────────── */}
      <div className="flex flex-col w-64 shrink-0 border-r border-zinc-800">
        {/* 헤더 */}
        <div className="flex items-center gap-1.5 px-2 py-2 bg-zinc-900 border-b border-zinc-800 shrink-0">
          <span className="text-xs text-zinc-400 truncate flex-1 font-mono">
            {activeProject ? activeProject.name : '프로젝트 선택'}
          </span>
          <Button variant="ghost" size="icon" className="h-5 w-5 text-zinc-500 hover:text-zinc-200"
            onClick={loadTree} disabled={loading}>
            <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
          </Button>
        </div>

        {/* 검색창 */}
        <div className="px-2 py-1.5 border-b border-zinc-800 shrink-0">
          <div className="flex items-center gap-1.5 bg-zinc-800 rounded px-2 py-1">
            <Search className="w-3 h-3 text-zinc-500 shrink-0" />
            <input
              value={query}
              onChange={e => setQuery(e.target.value)}
              placeholder="파일 검색..."
              className="flex-1 bg-transparent text-xs text-zinc-200 placeholder:text-zinc-600 outline-none"
            />
            {query && (
              <button onClick={() => setQuery('')}>
                <X className="w-3 h-3 text-zinc-500 hover:text-zinc-200" />
              </button>
            )}
          </div>
        </div>

        {/* 트리 / 검색 결과 */}
        <ScrollArea className="flex-1">
          <div className="py-1 px-1">
            {searchResults ? (
              searchResults.length > 0 ? (
                searchResults.map(n => (
                  <SearchResultItem
                    key={n.path} node={n}
                    selectedPath={selectedNode?.path ?? null}
                    onSelect={handleSelect}
                    rootPath={rootPath}
                  />
                ))
              ) : (
                <p className="text-xs text-zinc-600 text-center py-6">검색 결과 없음</p>
              )
            ) : (
              tree.length === 0 && !loading
                ? <p className="text-xs text-zinc-600 text-center py-6">파일 없음</p>
                : tree.map(node => (
                    <FileNodeItem
                      key={node.path} node={node}
                      selectedPath={selectedNode?.path ?? null}
                      onSelect={handleSelect}
                    />
                  ))
            )}
          </div>
        </ScrollArea>
      </div>

      {/* ── 우측: 코드 뷰어 ──────────────────────── */}
      <div className="flex-1 min-w-0 overflow-hidden">
        {fileLoading ? (
          <div className="flex items-center justify-center h-full">
            <RefreshCw className="w-5 h-5 text-zinc-600 animate-spin" />
          </div>
        ) : selectedNode && fileContent !== null ? (
          <MonacoEditor path={selectedNode.path} content={fileContent} getApiUrl={getApiUrl} />
        ) : (
          <div className="flex flex-col items-center justify-center h-full gap-2 text-center">
            <File className="w-10 h-10 text-zinc-700" />
            <p className="text-sm text-zinc-500">파일을 선택하세요</p>
            <p className="text-xs text-zinc-600">왼쪽 트리에서 파일을 클릭하면 내용이 표시됩니다</p>
          </div>
        )}
      </div>
    </div>
  )
}
