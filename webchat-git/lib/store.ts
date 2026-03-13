/**
 * 프로젝트 저장소 — ~/.webchat-git/projects.json
 * 서버사이드 전용 (Node.js fs)
 */
import { readFile, writeFile, mkdir } from 'fs/promises'
import { join } from 'path'
import { randomUUID } from 'crypto'

const STORE_DIR  = join(process.env.HOME ?? '~', '.webchat-git')
const STORE_FILE = join(STORE_DIR, 'projects.json')

export interface Project {
  id: string
  name: string
  repo: string          // "owner/repo"
  token: string         // GitHub personal access token
  cwd: string           // 로컬 작업 경로
  port: number          // jikime serve 포트
  pid: number | null    // 프로세스 ID (실행 중일 때)
  status: 'stopped' | 'running'
  createdAt: string
}

// ─── 내부 헬퍼 ───────────────────────────────────────────────────────────────

async function readStore(): Promise<Project[]> {
  try {
    const raw = await readFile(STORE_FILE, 'utf8')
    return JSON.parse(raw) as Project[]
  } catch {
    return []
  }
}

async function writeStore(projects: Project[]): Promise<void> {
  await mkdir(STORE_DIR, { recursive: true })
  await writeFile(STORE_FILE, JSON.stringify(projects, null, 2), 'utf8')
}

// ─── 퍼블릭 API ──────────────────────────────────────────────────────────────

export async function listProjects(): Promise<Project[]> {
  return readStore()
}

export async function getProject(id: string): Promise<Project | null> {
  const all = await readStore()
  return all.find(p => p.id === id) ?? null
}

export async function createProject(
  data: Omit<Project, 'id' | 'pid' | 'status' | 'createdAt' | 'port'>
): Promise<Project> {
  const all = await readStore()
  const usedPorts = new Set(all.map(p => p.port))
  // 8001부터 사용 가능한 포트 탐색
  let port = 8001
  while (usedPorts.has(port)) port++

  const project: Project = {
    ...data,
    id: randomUUID(),
    port,
    pid: null,
    status: 'stopped',
    createdAt: new Date().toISOString(),
  }
  await writeStore([...all, project])
  return project
}

export async function updateProject(id: string, patch: Partial<Project>): Promise<void> {
  const all = await readStore()
  const updated = all.map(p => p.id === id ? { ...p, ...patch } : p)
  await writeStore(updated)
}

export async function deleteProject(id: string): Promise<void> {
  const all = await readStore()
  await writeStore(all.filter(p => p.id !== id))
}
