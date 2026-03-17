# 업데이트 가이드

소스가 변경된 후 이미 배포된 Docker 컨테이너를 업데이트하는 방법입니다.

> `claude_data` 볼륨은 재빌드/재시작과 무관하게 유지됩니다.
> Claude 인증 정보와 세션 히스토리는 업데이트 후에도 그대로 남아있습니다.

---

## 방법 1 — 서버에서 직접 재빌드 (권장)

서버에 git과 소스가 있는 경우 가장 단순한 방법입니다.

```bash
# 1. 최신 소스 받기
git pull

# 2. 이미지 재빌드 + 컨테이너 교체
docker compose up -d --build --force-recreate
```

---

## 방법 2 — 로컬에서 빌드 후 원격에 전달

원격 서버에 git이나 소스가 없는 경우입니다.

```bash
# 로컬: 이미지 빌드 후 tar 파일로 저장
docker build -t webchat:latest .
docker save webchat:latest | gzip > webchat.tar.gz

# 원격 서버로 전송
scp webchat.tar.gz user@서버IP:/opt/webchat/

# 원격 서버: 이미지 로드 후 재시작
ssh user@서버IP '
  docker load < /opt/webchat/webchat.tar.gz
  cd /opt/webchat
  docker compose up -d --force-recreate
'
```

---

## 방법 3 — Docker Registry 활용 (다수 서버)

Docker Hub 또는 사설 레지스트리를 이용하면 여러 서버를 일괄 업데이트할 수 있습니다.

**로컬에서 빌드 후 push**

```bash
docker build -t myregistry/webchat:latest .
docker push myregistry/webchat:latest
```

**원격 서버에서 pull 후 재시작**

```bash
ssh user@서버IP '
  docker pull myregistry/webchat:latest
  cd /opt/webchat
  docker compose up -d --force-recreate
'
```

이 방법을 사용할 경우 `docker-compose.yml`에서 `build:` 블록을 제거하고 `image:`만 지정합니다.

```yaml
services:
  webchat:
    image: myregistry/webchat:latest   # build: 블록 제거
    container_name: webchat
    ...
```

---

## 방법 선택 기준

| 상황 | 권장 방법 |
|---|---|
| 서버에 git + 소스 있음 | 방법 1 |
| 서버에 소스 없음 | 방법 2 |
| 서버가 여러 대 | 방법 3 |

---

## 업데이트 후 확인

```bash
# 컨테이너 상태 확인
docker compose ps

# 로그 확인
docker compose logs -f

# 헬스체크 확인
curl http://localhost:4000/api/ws/health
```
