import { NextRequest } from 'next/server'
import { listProjects, createProject, deleteProject, getProject, updateProject } from '@/lib/store'
import { startServe, stopServe, isRunning, isPortRunning, getPidFromPort } from '@/lib/serve'
import { initGitRepo } from '@/lib/git'

export const dynamic = 'force-dynamic'

/** GET /api/projects — 프로젝트 목록 */
export async function GET() {
  const projects = await listProjects()

  // pid 체크 우선, 없으면 포트 응답으로 fallback
  const checked = await Promise.all(projects.map(async p => {
    const byPid = !!(p.pid && isRunning(p.pid))
    const running = byPid || await isPortRunning(p.port)

    // 포트는 살아있는데 pid 정보가 없으면 lsof로 복구
    if (running && !byPid) {
      const pid = getPidFromPort(p.port)
      if (pid) await updateProject(p.id, { pid, status: 'running' })
    }

    return { ...p, token: '***', status: running ? 'running' : 'stopped' }
  }))

  return Response.json(checked)
}

/** POST /api/projects — 프로젝트 생성 */
export async function POST(request: NextRequest) {
  const body = await request.json()
  const { name, repo, token, cwd } = body as Record<string, string>

  if (!name || !repo || !token || !cwd) {
    return Response.json({ error: '필수 항목 누락' }, { status: 400 })
  }
  if (!/^[^/]+\/[^/]+$/.test(repo)) {
    return Response.json({ error: 'repo 형식은 owner/repo 이어야 합니다' }, { status: 400 })
  }

  // GitHub 저장소 확인/생성 + 로컬 git 초기화
  try {
    await initGitRepo(cwd, repo, token)
  } catch (err) {
    return Response.json({ error: (err as Error).message }, { status: 500 })
  }

  const project = await createProject({ name, repo, token, cwd })
  return Response.json({ ...project, token: '***' })
}

/** DELETE /api/projects?id=xxx — 프로젝트 삭제 */
export async function DELETE(request: NextRequest) {
  const id = new URL(request.url).searchParams.get('id')
  if (!id) return Response.json({ error: 'id 필요' }, { status: 400 })

  const project = await getProject(id)
  if (!project) return Response.json({ error: '없는 프로젝트' }, { status: 404 })

  if (project.pid && isRunning(project.pid)) stopServe(id, project.pid)
  await deleteProject(id)
  return Response.json({ ok: true })
}

/** PATCH /api/projects — serve 시작/중지 */
export async function PATCH(request: NextRequest) {
  const { id, action } = await request.json() as { id: string; action: 'start' | 'stop' }
  const project = await getProject(id)
  if (!project) return Response.json({ error: '없는 프로젝트' }, { status: 404 })

  if (action === 'start') {
    const alreadyRunning = !!(project.pid && isRunning(project.pid)) || await isPortRunning(project.port)
    if (alreadyRunning) {
      // pid 복구 후 running 상태로 응답
      const pid = project.pid ?? getPidFromPort(project.port)
      if (pid) await updateProject(id, { pid, status: 'running' })
      return Response.json({ error: '이미 실행 중', alreadyRunning: true }, { status: 409 })
    }
    try {
      const pid = await startServe(project)
      await updateProject(id, { pid, status: 'running' })
      return Response.json({ ok: true, pid })
    } catch (err) {
      return Response.json({ error: (err as Error).message }, { status: 500 })
    }
  }

  if (action === 'stop') {
    if (project.pid) stopServe(id, project.pid)
    await updateProject(id, { pid: null, status: 'stopped' })
    return Response.json({ ok: true })
  }

  return Response.json({ error: 'action은 start|stop' }, { status: 400 })
}
