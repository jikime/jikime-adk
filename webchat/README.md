# Webchat

Claude Code를 웹 브라우저에서 사용할 수 있는 웹 인터페이스입니다.
로컬 및 원격 서버 모두 지원하며, 채팅 · 파일 편집 · 터미널 · Git 작업을 한 화면에서 처리할 수 있습니다.

## 빠른 시작

### 로컬 실행

```bash
pnpm install
pnpm dev
```

`http://localhost:4000` 접속

### Docker 실행

```bash
docker compose up -d --build

# Claude CLI 인증 (최초 1회)
docker exec -it webchat bash
claude auth login
```

`http://localhost:4000` 접속

---

## 주요 기능

- **Claude 채팅** — 스트리밍 응답, 세션 히스토리, 도구 사용 승인
- **파일 편집기** — 파일 트리 탐색 + Monaco 에디터
- **웹 터미널** — node-pty 기반 쉘 접속 (xterm.js)
- **Git 패널** — 상태 조회, 스테이징, 커밋, 브랜치 관리
- **원격 서버** — 여러 서버를 등록하고 브라우저에서 전환
- **권한 모드** — bypassPermissions / acceptEdits / default

---

## 문서

자세한 내용은 `docs/` 디렉터리를 참고하세요.

| 문서 | 내용 |
|---|---|
| [설치 가이드](./docs/installation.md) | 로컬 / Docker 설치 방법 |
| [사용법](./docs/usage.md) | 화면 구성, 채팅, 터미널, Git 사용법 |
| [원격 서버 연결](./docs/remote-server.md) | 원격 서버 등록 및 접속 방법 |
| [아키텍처](./docs/architecture.md) | 프로젝트 구조 및 기술 스택 |
| [트러블슈팅](./docs/troubleshooting.md) | 자주 발생하는 오류 해결법 |
| [업데이트 가이드](./docs/update.md) | 배포된 컨테이너 업데이트 방법 |

---

## 스크립트

| 명령 | 설명 |
|---|---|
| `pnpm dev` | 개발 서버 실행 (tsx server.ts) |
| `pnpm build` | Next.js 프로덕션 빌드 |
| `pnpm fix-pty` | Linux node-pty 수동 재컴파일 |
