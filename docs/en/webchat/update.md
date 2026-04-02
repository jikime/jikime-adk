# Update Guide

## Method 1 — JikiME CLI (Recommended)

```bash
jikime webchat install
```

Downloads the latest webchat source, runs `pnpm install` + `pnpm build`. Existing `node_modules` and `.next` are preserved.

---

## Method 2 — Docker: Rebuild on Server

When git and source are available on the server.

```bash
git pull
docker compose up -d --build --force-recreate
```

---

## Method 3 — Docker: Build Locally, Transfer to Remote

When the remote server has no source.

```bash
# Local: build and save image
docker build -t webchat:latest .
docker save webchat:latest | gzip > webchat.tar.gz

# Transfer to remote
scp webchat.tar.gz user@server-ip:/opt/webchat/

# Remote: load and restart
ssh user@server-ip '
  docker load < /opt/webchat/webchat.tar.gz
  cd /opt/webchat
  docker compose up -d --force-recreate
'
```

---

## Method 4 — Docker Registry (Multiple Servers)

```bash
# Local
docker build -t myregistry/webchat:latest .
docker push myregistry/webchat:latest

# Remote
docker pull myregistry/webchat:latest
docker compose up -d --force-recreate
```

---

## Selection Guide

| Scenario | Recommended |
|---|---|
| JikiME-ADK installed | Method 1 |
| Server has git + source | Method 2 |
| Server has no source | Method 3 |
| Multiple servers | Method 4 |

---

## Post-Update Verification

```bash
docker compose ps
docker compose logs -f
curl http://localhost:4000/api/ws/health
```

> The `claude_data` volume persists across rebuilds/restarts. Claude auth and session history are preserved after updates.
