/**
 * GitHub 원격 저장소 확인/생성 + 로컬 git 초기화
 */
import { mkdir } from 'fs/promises'
import { existsSync } from 'fs'
import { join } from 'path'
import { execFile } from 'child_process'
import { promisify } from 'util'

const exec = promisify(execFile)

async function git(args: string[], cwd: string) {
  return exec('git', args, { cwd })
}

/** GitHub API 공통 헤더 */
function ghHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
    Accept: 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28',
    'Content-Type': 'application/json',
  }
}

/**
 * GitHub 원격 저장소 확인 → 없으면 생성
 */
async function ensureGithubRepo(token: string, repo: string): Promise<'existed' | 'created'> {
  const checkRes = await fetch(`https://api.github.com/repos/${repo}`, {
    headers: ghHeaders(token),
  })

  if (checkRes.ok) return 'existed'

  if (checkRes.status !== 404) {
    const err = await checkRes.json() as { message?: string }
    throw new Error(`GitHub API 오류: ${err.message ?? checkRes.status}`)
  }

  const repoName = repo.split('/')[1]
  const owner = repo.split('/')[0]

  const userRes = await fetch('https://api.github.com/user', { headers: ghHeaders(token) })
  if (!userRes.ok) throw new Error('토큰으로 GitHub 사용자 정보를 가져올 수 없어요')
  const user = await userRes.json() as { login: string }

  let createRes: Response
  if (user.login === owner) {
    createRes = await fetch('https://api.github.com/user/repos', {
      method: 'POST',
      headers: ghHeaders(token),
      body: JSON.stringify({ name: repoName, private: false, auto_init: false }),
    })
  } else {
    createRes = await fetch(`https://api.github.com/orgs/${owner}/repos`, {
      method: 'POST',
      headers: ghHeaders(token),
      body: JSON.stringify({ name: repoName, private: false, auto_init: false }),
    })
  }

  if (!createRes.ok) {
    const err = await createRes.json() as { message?: string }
    throw new Error(`저장소 생성 실패: ${err.message ?? createRes.status}`)
  }

  return 'created'
}

/**
 * 로컬 git 초기화 + remote origin + branch tracking 설정
 *
 * .git/config 결과:
 *   [remote "origin"]
 *     url = https://github.com/owner/repo.git
 *     fetch = +refs/heads/*:refs/remotes/origin/*
 *   [branch "main"]
 *     remote = origin
 *     merge = refs/heads/main
 */
export async function initGitRepo(
  cwd: string,
  repo: string,
  token: string,
): Promise<{ repoStatus: 'existed' | 'created'; localStatus: 'existed' | 'initialized' }> {
  // 1. GitHub 원격 저장소 확인/생성
  const repoStatus = await ensureGithubRepo(token, repo)

  // 2. 로컬 디렉터리 생성
  await mkdir(cwd, { recursive: true })

  const gitDir = join(cwd, '.git')
  const alreadyGit = existsSync(gitDir)
  const localStatus = alreadyGit ? 'existed' : 'initialized'

  if (!alreadyGit) {
    // 3. git init + 기본 브랜치를 main으로
    await git(['init', '-b', 'main'], cwd).catch(async () => {
      // 구버전 git은 -b 미지원 → init 후 branch rename
      await git(['init'], cwd)
      await git(['checkout', '-b', 'main'], cwd).catch(() => {
        // 이미 main이면 무시
      })
    })

    // 4. 빈 초기 커밋 (branch가 실제로 생성되려면 커밋이 필요)
    await git(['commit', '--allow-empty', '-m', 'chore: init project'], cwd)
  }

  // 5. remote origin 설정
  const remoteUrl = `https://github.com/${repo}.git`
  try {
    const { stdout } = await git(['remote', 'get-url', 'origin'], cwd)
    if (stdout.trim() !== remoteUrl) {
      await git(['remote', 'set-url', 'origin', remoteUrl], cwd)
    }
  } catch {
    await git(['remote', 'add', 'origin', remoteUrl], cwd)
  }

  // 6. [branch "main"] tracking 설정
  //    remote에 main이 있으면 fetch 후 --set-upstream-to
  //    없으면 git config로 직접 기록 (push 할 때 자동 인식)
  try {
    await git(['fetch', 'origin', 'main'], cwd)
    await git(['branch', '--set-upstream-to=origin/main', 'main'], cwd)
  } catch {
    // 원격에 아직 main이 없는 경우 → config에 직접 기록
    await git(['config', 'branch.main.remote', 'origin'], cwd)
    await git(['config', 'branch.main.merge', 'refs/heads/main'], cwd)
  }

  return { repoStatus, localStatus }
}
