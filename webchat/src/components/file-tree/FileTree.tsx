'use client'

import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import {
  ChevronRight, ChevronDown, File, Folder, FolderOpen,
  RefreshCw, Search, X, Save, Check,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useProject } from '@/contexts/ProjectContext'
import { useServer } from '@/contexts/ServerContext'
import { useLocale } from '@/contexts/LocaleContext'
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
  return EXT_COLOR[ext] ?? 'text-muted-foreground'
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
          isDir ? 'text-foreground/80 hover:text-foreground hover:bg-muted' : 'hover:bg-muted',
          !isDir && selectedPath === node.path
            ? 'bg-muted text-foreground'
            : !isDir && 'text-muted-foreground hover:text-foreground'
        )}
        style={{ paddingLeft: `${indent + 8}px`, paddingRight: '8px' }}
        onClick={() => isDir ? setOpen(v => !v) : onSelect(node)}
      >
        {isDir ? (
          <>
            {open
              ? <ChevronDown className="w-3 h-3 shrink-0 text-muted-foreground" />
              : <ChevronRight className="w-3 h-3 shrink-0 text-muted-foreground" />}
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
          <span className="text-muted-foreground/50 text-[10px] shrink-0">{formatSize(node.size)}</span>
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
        'flex flex-col w-full text-left px-2 py-1.5 rounded text-xs transition-colors hover:bg-muted',
        selectedPath === node.path ? 'bg-muted text-foreground' : 'text-muted-foreground'
      )}
    >
      <div className="flex items-center gap-1.5">
        <File className={cn('w-3 h-3 shrink-0', extColor(node.name))} />
        <span className="font-medium text-foreground">{node.name}</span>
      </div>
      <span className="text-muted-foreground/50 text-[10px] font-mono truncate pl-4">{rel}</span>
    </button>
  )
}

// ── Monaco 에디터 ─────────────────────────────────────────────────
function MonacoEditor({ path, content, getApiUrl, onClose }: { path: string; content: string; getApiUrl: (p: string) => string; onClose: () => void }) {
  const { t } = useLocale()
  const name = path.split('/').pop() ?? path
  const lang = extLang(name)

  const [value, setValue]   = useState(content)
  const [dirty, setDirty]   = useState(false)
  const [saved, setSaved]   = useState(false)
  const [saving, setSaving] = useState(false)
  const saveRef = useRef<() => void>(() => {})

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
      <div className="flex items-center gap-2 px-3 py-1.5 bg-white dark:bg-accent border-b border-border shrink-0">
        <div className="flex items-center gap-1.5 min-w-0 flex-1">
          <File className={cn('w-3.5 h-3.5 shrink-0', extColor(name))} />
          <span className="text-xs font-mono truncate min-w-0" title={path}>
            <span className="text-muted-foreground">{path.slice(0, path.lastIndexOf('/') + 1)}</span>
            <span className="text-foreground">{name}</span>
          </span>
          <span className="text-muted-foreground/50 text-[10px] shrink-0 ml-1">{lang}</span>
          {dirty && (
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0" title={t.files.unsavedChanges} />
          )}
        </div>
        <Button
          variant="ghost" size="sm"
          onClick={handleSave}
          disabled={!dirty || saving}
          className={cn(
            'flex items-center gap-1 h-6 px-2 text-xs shrink-0',
            dirty && !saving ? 'text-foreground' : 'text-muted-foreground/50'
          )}
        >
          {saved
            ? <><Check className="w-3 h-3 text-emerald-400" /><span className="text-emerald-400">{t.files.saved}</span></>
            : saving
              ? <><RefreshCw className="w-3 h-3 animate-spin" /><span>{t.files.saving}</span></>
              : <><Save className="w-3 h-3" /><span>{t.files.save}</span></>
          }
        </Button>
        <Button
          variant="ghost" size="icon-sm"
          onClick={onClose}
          className="h-6 w-6 text-muted-foreground hover:text-foreground shrink-0"
          title="닫기"
        >
          <X className="w-3.5 h-3.5" />
        </Button>
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
  const { t } = useLocale()
  const { activeProject } = useProject()
  const { getApiUrl }     = useServer()

  const [tree, setTree]           = useState<FileNode[]>([])
  const [resolvedPath, setResolvedPath] = useState<string>('')
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
      if (res.ok) {
        const data = await res.json()
        // API가 { path, tree } 형태 또는 이전 호환 배열 형태 모두 처리
        if (Array.isArray(data)) {
          setTree(data)
          setResolvedPath(activeProject.path)
        } else {
          setTree(data.tree ?? [])
          setResolvedPath(data.path ?? activeProject.path)
        }
      }
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
        setFileContent(t.files.readError)
      }
    } catch {
      setFileContent(t.files.readErrorGeneric)
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

  const rootPath = resolvedPath || activeProject?.path || ''

  return (
    <div className="flex h-full bg-muted dark:bg-background rounded-lg overflow-hidden border border-border">

      {/* ── 좌측: 파일 트리 ──────────────────────── */}
      <div className="flex flex-col w-64 shrink-0 border-r border-border">
        {/* 헤더 */}
        <div className="flex items-center gap-1.5 px-2 py-2 bg-white dark:bg-accent border-b border-border shrink-0">
          <span className="text-xs text-muted-foreground truncate flex-1 font-mono" title={rootPath}>
            {activeProject ? (rootPath || activeProject.name) : t.files.selectProject}
          </span>
          <Button variant="ghost" size="icon" className="h-5 w-5 text-muted-foreground hover:text-foreground"
            onClick={loadTree} disabled={loading}>
            <RefreshCw className={cn('w-3 h-3', loading && 'animate-spin')} />
          </Button>
        </div>

        {/* 검색창 */}
        <div className="px-2 py-1.5 border-b border-border shrink-0">
          <div className="flex items-center gap-1.5 bg-muted rounded px-2 py-1">
            <Search className="w-3 h-3 text-muted-foreground shrink-0" />
            <input
              value={query}
              onChange={e => setQuery(e.target.value)}
              placeholder={t.files.search}
              className="flex-1 bg-transparent text-xs text-foreground placeholder:text-muted-foreground/40 outline-none"
            />
            {query && (
              <button onClick={() => setQuery('')}>
                <X className="w-3 h-3 text-muted-foreground hover:text-foreground" />
              </button>
            )}
          </div>
        </div>

        {/* 트리 / 검색 결과 */}
        <div className="flex-1 overflow-y-auto overflow-x-hidden
          [&::-webkit-scrollbar]:w-[3px]
          [&::-webkit-scrollbar-track]:bg-transparent
          [&::-webkit-scrollbar-thumb]:bg-transparent
          [&::-webkit-scrollbar-thumb]:rounded-full
          hover:[&::-webkit-scrollbar-thumb]:bg-muted-foreground/25
          [&::-webkit-scrollbar-thumb]:transition-colors">
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
                <p className="text-xs text-muted-foreground/50 text-center py-6">{t.files.noResults}</p>
              )
            ) : (
              tree.length === 0 && !loading
                ? <p className="text-xs text-muted-foreground/50 text-center py-6">{t.files.noFiles}</p>
                : tree.map(node => (
                    <FileNodeItem
                      key={node.path} node={node}
                      selectedPath={selectedNode?.path ?? null}
                      onSelect={handleSelect}
                    />
                  ))
            )}
          </div>
        </div>
      </div>

      {/* ── 우측: 코드 뷰어 ──────────────────────── */}
      <div className="flex-1 min-w-0 overflow-hidden">
        {fileLoading ? (
          <div className="flex items-center justify-center h-full">
            <RefreshCw className="w-5 h-5 text-muted-foreground/50 animate-spin" />
          </div>
        ) : selectedNode && fileContent !== null ? (
          <MonacoEditor
            path={selectedNode.path}
            content={fileContent}
            getApiUrl={getApiUrl}
            onClose={() => { setSelectedNode(null); setFileContent(null) }}
          />
        ) : (
          <div className="flex flex-col items-center justify-center h-full gap-2 text-center">
            <File className="w-10 h-10 text-muted-foreground/30" />
            <p className="text-sm text-muted-foreground">{t.files.selectFile}</p>
            <p className="text-xs text-muted-foreground/50">{t.files.selectFileHint}</p>
          </div>
        )}
      </div>
    </div>
  )
}
