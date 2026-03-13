import { NextRequest } from 'next/server'
import { getProject } from '@/lib/store'

export const dynamic = 'force-dynamic'

/**
 * GET /api/issue?projectId=xxx
 * 최근 5개 이슈 목록 반환 (open + closed, 최신순)
 */
export async function GET(request: NextRequest) {
  const projectId = new URL(request.url).searchParams.get('projectId')
  if (!projectId) return Response.json({ error: 'projectId 필요' }, { status: 400 })

  const project = await getProject(projectId)
  if (!project) return Response.json({ error: '없는 프로젝트' }, { status: 404 })

  const [owner, repo] = project.repo.split('/')

  const res = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/issues?state=all&per_page=5&sort=created&direction=desc`,
    { headers: githubHeaders(project.token) }
  )

  if (!res.ok) {
    const err = await res.json().catch(() => ({})) as { message?: string }
    return Response.json({ error: err.message ?? res.status }, { status: res.status })
  }

  const issues = await res.json() as Array<{
    number: number
    title: string
    state: string
    html_url: string
    created_at: string
    closed_at: string | null
    labels: Array<{ name: string }>
  }>

  // PR은 제외 (pull_request 필드가 있으면 PR)
  const filtered = issues
    .filter((i: Record<string, unknown>) => !i.pull_request)
    .map(i => ({
      number: i.number,
      title: i.title,
      state: i.state,           // 'open' | 'closed'
      url: i.html_url,
      createdAt: i.created_at,
      closedAt: i.closed_at,
      labels: i.labels.map(l => l.name),
    }))

  return Response.json(filtered)
}

/**
 * POST /api/issue
 * body: { projectId, title, body? }
 * → GitHub 이슈 생성 (jikime-todo 라벨)
 */
export async function POST(request: NextRequest) {
  const { projectId, title, body } = await request.json() as {
    projectId: string
    title: string
    body?: string
  }

  if (!projectId || !title) {
    return Response.json({ error: 'projectId, title 필수' }, { status: 400 })
  }

  const project = await getProject(projectId)
  if (!project) return Response.json({ error: '없는 프로젝트' }, { status: 404 })

  const [owner, repo] = project.repo.split('/')

  // 1. jikime-todo 라벨 자동 생성 (없으면)
  await ensureLabel(owner, repo, project.token, 'jikime-todo', '0075ca')
  await ensureLabel(owner, repo, project.token, 'jikime-done', '0e8a16')

  // 2. 이슈 생성
  // {{ }} 는 jikime serve 템플릿 변수로 해석되므로 이스케이프 처리
  const safeBody = (body ?? '').replace(/\{\{/g, '{ {').replace(/\}\}/g, '} }')

  const res = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/issues`,
    {
      method: 'POST',
      headers: githubHeaders(project.token),
      body: JSON.stringify({
        title,
        body: safeBody,
        labels: ['jikime-todo'],
      }),
    }
  )

  if (!res.ok) {
    const err = await res.json().catch(() => ({})) as { message?: string }
    return Response.json(
      { error: `GitHub 이슈 생성 실패: ${err.message ?? res.status}` },
      { status: res.status }
    )
  }

  const issue = await res.json() as { number: number; html_url: string; title: string }
  return Response.json({
    number: issue.number,
    url: issue.html_url,
    title: issue.title,
  })
}

function githubHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
    Accept: 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28',
    'Content-Type': 'application/json',
  }
}

async function ensureLabel(
  owner: string, repo: string, token: string,
  name: string, color: string
): Promise<void> {
  const check = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/labels/${encodeURIComponent(name)}`,
    { headers: githubHeaders(token) }
  )
  if (check.ok) return  // 이미 존재

  await fetch(`https://api.github.com/repos/${owner}/${repo}/labels`, {
    method: 'POST',
    headers: githubHeaders(token),
    body: JSON.stringify({ name, color }),
  })
}
