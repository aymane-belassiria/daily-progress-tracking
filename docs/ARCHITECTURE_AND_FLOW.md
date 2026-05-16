# Daily Progress Tracking: Architecture and Request Flow

This document explains how the system is structured and how traffic flows through frontend, reverse proxy, and backend.

## 1) High-Level Architecture

- `frontend/` is a Vite + React single-page app (SPA).
- `backend/` is a Go API with SQLite storage.
- GitHub Pages hosts the built frontend assets.
- A VPS hosts the Go backend, SQLite DB, and reverse proxy (Caddy or Nginx).
- JWT auth protects all private API routes.
- NVIDIA-hosted Qwen is called by the backend for roadmap/hints generation.

## 2) Runtime Components

- Browser (user): loads static frontend from GitHub Pages.
- GitHub Pages: serves `frontend/dist`.
- Reverse proxy on VPS:
- Option A: Caddy (`deploy/Caddyfile`)
- Option B: Nginx (`deploy/nginx-progress-tracker.conf`)
- Go API process: `backend/bin/progress-api`, typically managed by systemd (`deploy/progress-tracker.service`).
- SQLite database: local file on VPS used by backend only.

## 3) Domain and Proxy Setup (nip.io Example)

This repo includes proxy examples using:

- `209.126.9.94.nip.io`

`nip.io` maps hostnames containing an IP back to that IP via wildcard DNS.  
So `209.126.9.94.nip.io` resolves to `209.126.9.94`, which is useful for quick setup/testing.

### Caddy flow

Current config:

```caddyfile
209.126.9.94.nip.io {
  reverse_proxy 127.0.0.1:4000
}
```

### Nginx flow

Current config:

```nginx
server {
    listen 80;
    server_name 209.126.9.94.nip.io;

    client_max_body_size 2m;

    location / {
        proxy_pass http://127.0.0.1:4000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Both proxy options expose public HTTP(S) and forward requests to local backend port `4000`.

## 4) End-to-End Request Flow

1. User opens GitHub Pages URL.
2. Browser downloads static files (`index.html`, JS, CSS).
3. Frontend calls API base URL (`VITE_API_BASE_URL`) for login and data.
4. Request reaches VPS reverse proxy (Caddy/Nginx).
5. Proxy forwards request to `127.0.0.1:4000`.
6. Go backend validates auth/token, reads/writes SQLite, optionally calls NVIDIA API.
7. Backend returns JSON response.
8. Frontend updates dashboard UI.

## 5) Authentication Flow

1. Frontend `POST /api/auth/login` with email/password.
2. Backend validates credentials from `backend/.env`.
3. Backend returns JWT token.
4. Frontend stores token and sends `Authorization: Bearer <token>` on protected routes.
5. Backend middleware validates token for routes like:
- `/api/dashboard`
- `/api/goals`
- `/api/entries`
- `/api/ai/roadmap`
- `/api/ai/hints`

## 6) AI Flow

1. Frontend sends prompt context to backend AI endpoints.
2. Backend builds model prompt using current goals/entries.
3. Backend calls NVIDIA-hosted Qwen using `NVIDIA_API_KEY`.
4. Backend returns generated content to frontend.

Security note: API key never goes to browser; it stays server-side in `backend/.env`.

## 7) Deployment Flow

### Frontend (GitHub Pages)

1. Push to `master`.
2. GitHub Actions workflow `.github/workflows/deploy-pages.yml` runs.
3. Workflow builds frontend with:
- `VITE_API_BASE_URL`
- optional `VITE_BASE_PATH`
4. Workflow uploads artifact and deploys to GitHub Pages.

### Backend (VPS)

1. Pull latest code on VPS.
2. Build backend binary: `go build -o bin/progress-api ./cmd/server`
3. Restart service: `systemctl restart progress-tracker.service`
4. Verify health: `curl http://127.0.0.1:4000/api/health`

## 8) Operational Notes

- `FRONTEND_ORIGIN` in backend `.env` must match your frontend origin for CORS.
- Keep backend on private loopback (`127.0.0.1:4000`) and expose only proxy ports externally.
- Use HTTPS in production for both frontend and backend origins.
- Do not commit secrets; keep credentials and API keys in environment files.
