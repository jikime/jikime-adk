# 설치 가이드

## 사전 요구사항

| 항목 | 버전 | 비고 |
|---|---|---|
| Node.js | 22 이상 | |
| pnpm | 최신 | `corepack enable && corepack prepare pnpm@latest --activate` |
| Claude CLI | 최신 | `npm install -g @anthropic-ai/claude-code` |
| Git | 2.x 이상 | 터미널 Git 패널 사용 시 |

---

## 방법 1 — 로컬 직접 실행

### 1. 의존성 설치

```bash
pnpm install
```

macOS에서는 설치 후 `spawn-helper` 실행 권한이 자동으로 부여됩니다.
Linux(Rocky / RHEL 계열)에서 node-pty 컴파일이 필요한 경우 아래를 참고하세요 → [트러블슈팅 — node-pty](./troubleshooting.md#node-pty-컴파일-실패-linux)

### 2. 개발 서버 실행

```bash
pnpm dev
```

브라우저에서 `http://localhost:4000` 으로 접속합니다.

### 3. 프로덕션 빌드 후 실행

```bash
pnpm build
NODE_ENV=production pnpm dev
```

### 환경변수

`.env.example`을 복사해 `.env`를 만들고 필요한 값을 설정합니다.

```bash
cp .env.example .env
```

| 변수 | 기본값 | 설명 |
|---|---|---|
| `PORT` | `4000` | 서버 포트 |
| `HOSTNAME` | `localhost` | 바인드 주소 (`0.0.0.0`으로 설정하면 외부 접근 허용) |
| `CLAUDE_PATH` | 자동 탐색 | Claude CLI 네이티브 바이너리 경로 (자동 탐색 실패 시 직접 지정) |

---

## 방법 2 — Docker (권장: 서버 배포)

### 사전 요구사항

- Docker Engine 24 이상
- Docker Compose v2

### 1. 이미지 빌드 및 컨테이너 실행

```bash
docker compose up -d --build
```

빌드 중 수행되는 작업:
- Node.js 22 + 빌드 도구(python3, gcc, g++) 설치
- `pnpm install` — node-pty 네이티브 컴파일 포함
- `next build` — Next.js 프로덕션 빌드
- Claude CLI 설치(`npm install -g @anthropic-ai/claude-code`)

### 2. Claude CLI 인증

컨테이너가 실행된 후 컨테이너 내부에서 직접 인증합니다.

```bash
# 컨테이너 쉘 접속
docker exec -it webchat bash

# Claude CLI 인증
claude auth login

# 인증 확인
claude --version

# 쉘 종료
exit
```

인증 정보는 `claude_data` 볼륨(`/root/.claude`)에 저장되므로 **컨테이너 재시작 후에도 유지**됩니다.

### 3. 접속

```
http://<서버 IP>:4000
```

### 주요 명령

```bash
# 로그 확인
docker compose logs -f

# 컨테이너 재시작
docker compose restart

# 중지 (볼륨 유지)
docker compose down

# 중지 + 볼륨 삭제 (인증 정보 포함 전체 초기화)
docker compose down -v

# 이미지 재빌드
docker compose up -d --build --force-recreate
```

### 볼륨

| 볼륨 이름 | 컨테이너 경로 | 내용 |
|---|---|---|
| `claude_data` | `/root/.claude` | Claude 인증 정보 · 세션 히스토리 · 프로젝트 목록 |

### 프로젝트 디렉터리 마운트

컨테이너 내부에서 호스트의 프로젝트 디렉터리에 접근하려면 `docker-compose.yml`의 volumes 항목을 수정합니다.

```yaml
volumes:
  - claude_data:/root/.claude
  - /home:/home          # 호스트 /home 을 컨테이너에 동일 경로로 마운트
  - /root:/root/projects # 호스트 /root 를 컨테이너 /root/projects 로 마운트
```

> **주의**: 마운트 경로를 변경하면 Claude가 프로젝트를 인식하는 경로도 달라집니다. 가능하면 호스트와 동일한 절대 경로로 마운트하는 것을 권장합니다.
