# 원격 서버 연결

로컬 PC의 브라우저에서 원격 서버(Rocky Linux, Ubuntu 등)에 있는 Webchat 인스턴스에 접속하는 방법입니다.

---

## 개요

```
[브라우저 (로컬 PC)]
       │  HTTP / WebSocket
       ▼
[원격 서버 :4000]
  webchat server.ts
       │
       ├── Claude Agent SDK  ──▶  Claude API
       ├── node-pty (터미널)
       └── 파일시스템 / Git
```

- **채팅**: 원격 서버에서 Claude가 실행되고, 원격 서버의 파일을 직접 편집합니다.
- **터미널**: 원격 서버의 쉘에 접속합니다.
- **파일/Git**: 원격 서버의 프로젝트 경로를 기준으로 동작합니다.

---

## 원격 서버 설정

### 1. Webchat 서버 실행 (원격)

원격 서버에서 외부 접근을 허용하려면 `HOSTNAME=0.0.0.0`으로 설정합니다.

```bash
# 직접 실행
HOSTNAME=0.0.0.0 pnpm dev

# 또는 .env 파일에 설정
echo "HOSTNAME=0.0.0.0" >> .env
pnpm dev
```

Docker를 사용하는 경우 `docker-compose.yml`에 이미 `HOSTNAME=0.0.0.0`이 설정되어 있습니다.

```bash
docker compose up -d --build
```

### 2. 방화벽 포트 허용

```bash
# Rocky Linux / RHEL
firewall-cmd --permanent --add-port=4000/tcp
firewall-cmd --reload

# Ubuntu
ufw allow 4000/tcp
```

---

## 브라우저에서 원격 서버 등록

### 1. 서버 추가

사이드바 상단의 **서버 선택** 드롭다운 → **서버 추가** 버튼을 클릭합니다.

| 필드 | 입력 예시 | 설명 |
|---|---|---|
| 이름 | `개발 서버` | 식별용 표시 이름 |
| 호스트 | `221.143.48.77:4000` | IP:포트 형식 (프로토콜 제외) |
| 보안 연결 | OFF | HTTPS/WSS 사용 시 ON |

> **주의**: 호스트 입력 시 `ws://` 또는 `http://` 등 프로토콜 접두사를 입력하면 자동으로 제거됩니다. `IP:포트` 형식만 입력하세요.

### 2. 서버 전환

사이드바의 서버 드롭다운에서 등록된 서버를 선택하면 즉시 해당 서버로 연결됩니다.

- WebSocket이 자동으로 재연결됩니다.
- 프로젝트 목록이 선택된 서버 기준으로 갱신됩니다.
- 서버 목록은 `localStorage`에 저장되어 브라우저를 닫아도 유지됩니다.

### 3. 연결 상태 확인

사이드바 서버 아이콘 옆의 점 색상으로 연결 상태를 확인합니다.

- 🟢 초록 — 연결됨
- ⚫ 회색 — 연결 중 / 끊김 (3초마다 자동 재시도)

---

## HTTPS / WSS 사용 (선택)

도메인과 SSL 인증서가 있는 경우 Nginx 등 리버스 프록시를 앞에 두고 WSS를 사용할 수 있습니다.

### Nginx 설정 예시

```nginx
server {
    listen 443 ssl;
    server_name webchat.example.com;

    ssl_certificate     /etc/letsencrypt/live/webchat.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/webchat.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:4000;
        proxy_http_version 1.1;

        # WebSocket 업그레이드 헤더
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;  # 긴 스트리밍 응답을 위해 타임아웃 연장
    }
}
```

브라우저에서 서버 추가 시 **보안 연결** 옵션을 ON으로 설정하면 `wss://` / `https://` 로 통신합니다.

---

## 서버 목록 관리

| 동작 | 방법 |
|---|---|
| 서버 추가 | 사이드바 → 서버 드롭다운 → 서버 추가 |
| 서버 편집 | 서버 항목 옆 편집 아이콘 클릭 |
| 서버 삭제 | 서버 항목 옆 삭제 아이콘 클릭 |
| 로컬 서버 | 삭제 불가 (기본 서버로 항상 유지) |
