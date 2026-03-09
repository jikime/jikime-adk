# 하네스 엔지니어링 테스트 플로우

새 프로젝트에 Harness Engineering을 처음 적용할 때의 전체 테스트 절차입니다.

---

## 사전 준비 사항

- `gh` CLI 설치 및 로그인 (`gh auth status`)
- Go 1.21+ 설치 (소스 빌드 시)
- GitHub 계정 (public repo 생성 권한)

---

## Step 0: 최신 jikime 설치

```bash
cd /path/to/jikime-adk

# 최신 소스로 빌드 후 설치
go install .

# 설치 확인 — serve init 서브커맨드가 보여야 함
jikime serve --help
```

**확인 포인트**: `Available Commands: init` 가 출력되어야 합니다.

---

## Step 1: GitHub 원격 저장소 생성

```bash
cd /path/to/harness-test

# GitHub에 public repo 생성
gh repo create <owner>/harness-test --public --description "Harness Engineering Test"

# 로컬 git 초기화 + 첫 커밋 + 원격 연결
git init
echo "# harness-test" > README.md
git add .
git commit -m "chore: initial commit"
git branch -M main
git remote add origin https://github.com/<owner>/harness-test.git
git push -u origin main
```

---

## Step 2: WORKFLOW.md 생성 (`jikime serve init`)

```bash
cd /path/to/harness-test

jikime serve init
```

대화형 마법사 입력값:

| 질문 | 입력값 | 비고 |
|------|--------|------|
| GitHub repo slug | `<owner>/harness-test` | git remote에서 자동 감지됨 |
| Active label | `jikime-todo` | Enter (기본값) |
| Workspace root | `/tmp/jikime-harness-test` | Enter (기본값) |
| HTTP status API port | `8888` | Enter (기본값) |
| Max concurrent agents | `1` | Enter (기본값) |

완료 후 `WORKFLOW.md`가 현재 디렉토리에 생성됩니다.

> **JiKiME-ADK 모드**: 프로젝트에 `.claude/` 디렉토리가 있으면 자동으로
> JiKiME-ADK 모드로 생성됩니다 (jarvis 에이전트 활용).

---

## Step 3: GitHub 라벨 생성

`serve init` 완료 후 출력된 명령어를 그대로 실행합니다:

```bash
# 작업 대상 라벨 (AI 에이전트가 이 라벨이 붙은 이슈를 처리)
gh label create "jikime-todo" \
  --repo <owner>/harness-test \
  --description "Ready for AI agent" \
  --color "0e8a16"

# 완료 라벨 (처리 완료 시 에이전트가 자동 부여)
gh label create "jikime-done" \
  --repo <owner>/harness-test \
  --description "Completed by AI agent" \
  --color "6f42c1"
```

---

## Step 4: 테스트용 GitHub 이슈 생성

```bash
gh issue create \
  --repo <owner>/harness-test \
  --title "Build a simple app with Next.js 16, Tailwind CSS 4, and shadcn/ui" \
  --body "## 요구사항

Next.js 16 App Router, Tailwind CSS 4, shadcn/ui를 사용해 간단한 앱을 구현해주세요.

### 기능
- 메인 페이지: 히어로 섹션 + 카드 목록
- 카드 컴포넌트: 제목, 설명, 배지, 버튼 포함
- 다크모드 토글 (shadcn/ui ThemeProvider)
- 반응형 레이아웃 (모바일/데스크톱)

### 기술 스택
- Next.js 16 (App Router)
- TypeScript
- Tailwind CSS 4
- shadcn/ui (Card, Button, Badge, Switch 컴포넌트)

### 설치 순서
\`\`\`bash
npx create-next-app@latest . --typescript --tailwind --app --yes
npx shadcn@latest init -y
npx shadcn@latest add card button badge switch
\`\`\`

### 파일 구조
\`\`\`
app/
  page.tsx              # 메인 페이지 (히어로 + 카드 목록)
  layout.tsx            # 루트 레이아웃 (ThemeProvider 포함)
  globals.css           # Tailwind CSS 4 설정
components/
  theme-provider.tsx    # 다크모드 Provider
  theme-toggle.tsx      # 다크모드 토글 버튼
  feature-card.tsx      # 재사용 카드 컴포넌트
\`\`\`

### 완료 조건
- \`npm run build\` 성공
- shadcn/ui 컴포넌트 정상 렌더링
- 다크모드 토글 동작
- TypeScript 오류 없음" \
  --label "jikime-todo"
```

이슈에 `jikime-todo` 라벨이 붙으면 `jikime serve`가 자동으로 감지합니다.

---

## Step 5: jikime serve 실행

```bash
cd /path/to/harness-test

jikime serve WORKFLOW.md
```

정상 동작 시 터미널에 다음 순서로 로그가 출력됩니다:

```
[poller]    found issue #1 "Build a simple app with Next.js 16, Tailwind CSS 4, and shadcn/ui" [jikime-todo]
[workspace] creating /tmp/jikime-harness-test/issue-1
[hook]      after_create: git clone https://github.com/<owner>/harness-test.git .
[hook]      before_run: syncing to latest main...
[agent]     starting claude on issue #1 (attempt 1)
...
[agent]     done — created PR #1
[poller]    issue #1 state changed → jikime-done
```

---

## Step 6: 상태 모니터링 (별도 터미널)

서비스가 실행 중인 상태에서 다른 터미널에서 확인합니다:

```bash
# 텍스트 대시보드 (사람이 읽기 편한 형태)
curl http://localhost:8888/

# JSON 상태 API
curl -s http://localhost:8888/api/v1/state | jq .

# 실행 중인 에이전트만 보기
curl -s http://localhost:8888/api/v1/state | jq '.running'

# 즉시 폴링 트리거 (이슈 빠르게 감지)
curl -s -X POST http://localhost:8888/api/v1/refresh | jq .
```

---

## Step 7: 결과 확인

```bash
# 생성된 PR 목록
gh pr list --repo <owner>/harness-test

# PR 상세 내용
gh pr view 1 --repo <owner>/harness-test

# 변경된 파일 diff
gh pr diff 1 --repo <owner>/harness-test
```

---

## 전체 플로우 요약

```
go install .
  ↓
cd harness-test
gh repo create + git push
  ↓
jikime serve init
  → WORKFLOW.md 생성
  ↓
gh label create jikime-todo
gh label create jikime-done
  ↓
gh issue create --label jikime-todo
  ↓
jikime serve WORKFLOW.md
  ↓
[다른 터미널] curl localhost:8888/status
  ↓
gh pr list → 결과 확인
```

---

## 자주 겪는 문제

| 증상 | 원인 | 해결 |
|------|------|------|
| `serve init` 커맨드 없음 | 구버전 바이너리 | `go install .` 재실행 |
| 이슈 감지 안 됨 | 라벨 이름 불일치 | WORKFLOW.md `active_states`와 라벨명 일치 확인 |
| clone 실패 | GitHub 인증 문제 | `gh auth status` 확인 |
| 에이전트 stall | `stall_timeout_ms` 초과 | WORKFLOW.md `claude.stall_timeout_ms` 값 늘리기 |
| port 충돌 | 8888 이미 사용 중 | `--port 9999` 또는 `server.port: 0` (비활성화) |

---

## 관련 문서

- [하네스 엔지니어링 개요](./harness-engineering.md)
- [WORKFLOW.md 레퍼런스](./harness-engineering.md#workflowmd-설정-레퍼런스)
- [SPEC.md](../../symphony/SPEC.md) (Symphony 원본 스펙)
