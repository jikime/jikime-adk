# Remote Server Connection

Connect to a webchat instance running on a remote server (Rocky Linux, Ubuntu, etc.) from your local browser.

---

## Overview

```
[Browser (Local PC)]
       |  HTTP / WebSocket
       v
[Remote Server :4000]
  webchat server.ts
       |
       +-- Claude Agent SDK  -->  Claude API
       +-- node-pty (terminal)
       +-- filesystem / Git
```

- **Chat**: Claude runs on the remote server and edits files directly on that server.
- **Terminal**: Connects to the remote server shell.
- **Files/Git**: Operates relative to project paths on the remote server.

---

## Remote Server Setup

### 1. Start Webchat Server (Remote)

Set `HOSTNAME=0.0.0.0` to allow external access.

```bash
# Direct execution
HOSTNAME=0.0.0.0 pnpm dev

# Or via JikiME CLI
HOSTNAME=0.0.0.0 jikime webchat start

# Or add to .env
echo "HOSTNAME=0.0.0.0" >> ~/.jikime/webchat/.env
jikime webchat start
```

For Docker, `HOSTNAME=0.0.0.0` is already configured in `docker-compose.yml`.

### 2. Open Firewall Port

```bash
# Rocky Linux / RHEL
firewall-cmd --permanent --add-port=4000/tcp
firewall-cmd --reload

# Ubuntu
ufw allow 4000/tcp
```

---

## Register Remote Server in Browser

### 1. Add Server

Click **Add Server** in the sidebar server dropdown.

| Field | Example | Description |
|---|---|---|
| Name | `Dev Server` | Display name |
| Host | `221.143.48.77:4000` | IP:port format (no protocol prefix) |
| Secure | OFF | Enable for HTTPS/WSS |

### 2. Switch Server

Select a registered server from the dropdown to connect immediately. WebSocket auto-reconnects and project list updates.

---

## HTTPS / WSS (Optional)

Use Nginx reverse proxy with SSL certificates for secure connections.

```nginx
server {
    listen 443 ssl;
    server_name webchat.example.com;

    ssl_certificate     /etc/letsencrypt/live/webchat.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/webchat.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:4000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 3600s;
    }
}
```

Enable **Secure Connection** when adding the server in the browser.
