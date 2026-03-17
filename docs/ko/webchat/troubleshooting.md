# 트러블슈팅

---

## node-pty 컴파일 실패 (Linux)

### 증상

서버 시작 시 아래 메시지가 출력되고 터미널 탭이 비활성화됩니다.

```
[server] node-pty 로드 실패 — 터미널 기능 비활성화됨
```

### 원인

Linux에서 `node-pty`는 네이티브 C++ 모듈로 컴파일이 필요합니다. 빌드 도구가 없거나 컴파일이 완료되지 않은 경우 발생합니다.

### 해결

**1. 빌드 도구 설치 (Rocky Linux / RHEL)**

```bash
sudo dnf install -y python3 make gcc gcc-c++
```

**Ubuntu / Debian**

```bash
sudo apt-get install -y python3 make gcc g++ build-essential
```

**2. node-pty 재컴파일**

```bash
# 자동 스크립트 실행
pnpm fix-pty

# 또는 직접 실행
bash scripts/fix-pty-linux.sh
```

**3. 서버 재시작**

```bash
pnpm dev
```

Docker를 사용하는 경우 이미지 빌드 시 자동으로 처리됩니다.

```bash
docker compose up -d --build
```

---

## Claude CLI를 찾지 못하는 경우

### 증상

```
[server] claude 네이티브 바이너리를 찾지 못했습니다.
```

채팅 입력 시 `Claude Code executable not found` 오류가 발생합니다.

### 원인

- Claude CLI가 설치되지 않았거나
- 설치된 경로가 탐색 대상에 없는 경우입니다.

### 해결

**1. Claude CLI 설치 확인**

```bash
which claude
claude --version
```

**2. CLAUDE_PATH 환경변수로 경로 직접 지정**

```bash
# 실행 시 직접 지정
CLAUDE_PATH=/usr/local/bin/claude pnpm dev

# 또는 .env 파일에 추가
echo "CLAUDE_PATH=/usr/local/bin/claude" >> .env
```

**3. Docker 환경**

`docker-compose.yml`의 environment에 이미 `CLAUDE_PATH=/usr/local/bin/claude`가 설정되어 있습니다. 컨테이너 내부에서 Claude CLI가 다른 경로에 설치된 경우 해당 경로로 수정하세요.

```bash
# 컨테이너 내부에서 경로 확인
docker exec -it webchat which claude
```

---

## 채팅 오류: exit code 1

### 증상

```
claude 종료 코드 1 (경로: /usr/local/bin/claude)
```

### 원인 및 해결

**1. Claude CLI 인증 미완료**

```bash
# 로컬 실행 시
claude auth login

# Docker 컨테이너 실행 시
docker exec -it webchat bash
claude auth login
```

**2. root 환경에서 dangerously-skip-permissions 차단**

root로 실행된 서버에서 `bypassPermissions` 모드를 선택하면 Claude CLI가 차단합니다.
서버가 자동으로 `acceptEdits` 모드로 전환하지만, 그래도 오류가 발생하면 채팅 권한 모드를 `default` 또는 `acceptEdits`로 변경해 보세요.

**3. API 키 / 인증 만료**

```bash
claude auth status
```

---

## 원격 서버 연결 실패

### 증상

사이드바에 `연결 중...` 상태가 계속 표시됩니다.

### 확인 사항

**1. 원격 서버가 실행 중인지 확인**

```bash
# 원격 서버에서
curl http://localhost:4000/api/ws/health
# 응답: {"ok":true}
```

**2. 방화벽 포트 개방 확인**

```bash
# Rocky Linux / RHEL
firewall-cmd --list-ports

# Ubuntu
ufw status
```

**3. 호스트 입력 형식 확인**

서버 추가 시 호스트는 `IP:포트` 형식으로 입력합니다. `ws://` 또는 `http://` 접두사를 포함하면 자동으로 제거되지만, 혼동을 피하기 위해 처음부터 접두사 없이 입력하는 것을 권장합니다.

```
올바른 예:  221.143.48.77:4000
잘못된 예:  ws://221.143.48.77:4000  (자동 제거되긴 하지만 권장하지 않음)
잘못된 예:  http://221.143.48.77:4000
```

---

## 프로젝트 경로가 잘못 표시되는 경우

### 증상

`jikime-adk` 같이 `-`가 포함된 디렉터리명이 `/`로 치환되어 `jikime/adk`로 표시됩니다.

### 원인

Claude는 프로젝트 경로의 `/`를 `-`로 치환하여 저장합니다. 경로에 `-`가 포함된 디렉터리가 있으면 단순 치환으로 복원이 불가능합니다.

### 현재 동작

서버에서 **파일시스템 탐색 기반 최장 일치 알고리즘**으로 실제 경로를 복원합니다. 실제 파일시스템에 해당 경로가 존재하면 자동으로 올바르게 표시됩니다. 경로가 여전히 잘못 표시된다면 해당 디렉터리가 서버에 실제로 존재하는지 확인하세요.

---

## pnpm 설치 중 node-pty 빌드 오류

### 증상

```
 ERR_PNPM_RECURSIVE_EXEC_FIRST_FAIL  Command "build" failed
```

### 원인

pnpm이 node-pty의 빌드 스크립트 실행을 허용하지 않아서 발생합니다.

### 해결

`package.json`에 `pnpm.allowedBuilds` 설정이 포함되어 있는지 확인합니다.

```json
"pnpm": {
  "allowedBuilds": ["node-pty"]
}
```

설정 후 다시 설치합니다.

```bash
pnpm install
```
