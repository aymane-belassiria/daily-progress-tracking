# Daily Progress Tracking

Private single-user progress tracking system with:

- React frontend intended for GitHub Pages
- Go + SQLite backend intended for a VPS
- JWT auth for one private account
- NVIDIA-hosted Qwen integration for roadmap generation and progress hints

## Architecture

- `frontend/`: Vite + React SPA
- `backend/`: compiled Go API with SQLite persistence
- GitHub Pages hosts the frontend build output
- Your VPS runs the backend and keeps the NVIDIA API key private

For a shareable architecture and deployment flow document (including Caddy, Nginx, and `nip.io` routing), see:
- [`docs/ARCHITECTURE_AND_FLOW.md`](docs/ARCHITECTURE_AND_FLOW.md)

## Features

- Weekly and monthly goal tracking
- Daily progress entries with wins, blockers, and notes
- Dashboard with completion summaries
- AI-generated weekly roadmap suggestions
- AI-generated daily hints based on your active goals and recent entries
- Single-user login protected by JWT

## Setup

1. Copy `backend/.env.example` to `backend/.env`.
2. Set your admin email, password, JWT secret, allowed frontend origin, and NVIDIA API key.
3. Install dependencies:

```bash
npm run install:all
```

4. Run the backend:

```bash
npm run dev:backend
```

5. Run the frontend:

```bash
npm run dev:frontend
```

## Deploy

### Backend on VPS

- Install Go on the VPS.
- Copy the `backend/` directory to the server.
- Create `backend/.env`.
- Set `FRONTEND_ORIGIN` to your GitHub Pages site origin, for example `https://your-username.github.io`.
- Build with `go build -o bin/progress-api ./cmd/server`.
- Start with `./bin/progress-api`.
- Put it behind Nginx or Caddy with HTTPS.

### Frontend on GitHub Pages

- In GitHub repository variables, set `VITE_API_BASE_URL` to your VPS backend origin, for example `https://api.yourdomain.com`.
- Optionally set `VITE_BASE_PATH` if you are not using the default project-page path of `/<repo-name>/`.
- Run `npm run build:frontend`.
- Publish `frontend/dist`.

### Deployment Notes

- GitHub Pages can host only the frontend. The Go API must stay on your VPS.
- The backend should be exposed over HTTPS on a stable domain or subdomain, not a raw IP:port.
- If `VITE_API_BASE_URL` includes a trailing slash, the frontend now normalizes it automatically.
- The GitHub Pages workflow reads `VITE_API_BASE_URL` and `VITE_BASE_PATH` from repository variables.

## Important

- Do not commit real secrets.
- The NVIDIA API key belongs only in `backend/.env`.
- The app is designed for a single private user account, not multi-tenant use.
