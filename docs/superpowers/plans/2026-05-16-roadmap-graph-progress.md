# Roadmap Graph Progress Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build goal-attached roadmap graphs with clickable details, daily task decomposition, weekly/monthly tracking, and progress scoring.

**Architecture:** Add roadmap persistence and endpoints to the Go backend, using structured JSON for AI output and SQLite for task completion. Add focused frontend utilities and React components inside the existing single-page app, with plain SVG for graph/chart rendering.

**Tech Stack:** Go, SQLite, React, Vite, Node test runner, plain SVG/CSS.

---

### Task 1: Backend Roadmap Model and Store

**Files:**
- Modify: `backend/internal/app/models.go`
- Modify: `backend/internal/app/store.go`
- Test: `backend/internal/app/roadmap_test.go`

- [ ] Add roadmap structs with nodes and daily tasks.
- [ ] Add SQLite tables for `roadmaps` and `roadmap_tasks`.
- [ ] Add store methods to save, fetch, complete tasks, and compute score snapshots.
- [ ] Test that empty stores return empty slices and task completion updates scores.

### Task 2: Backend API and AI JSON Generation

**Files:**
- Modify: `backend/internal/app/server.go`
- Modify: `backend/internal/app/prompts.go`
- Test: `backend/internal/app/roadmap_test.go`

- [ ] Add `GET /api/roadmaps`, `POST /api/roadmaps/generate`, and `PUT /api/roadmap-tasks/{id}`.
- [ ] Generate strict JSON for roadmap nodes/tasks.
- [ ] Fall back to deterministic task decomposition when AI is unavailable or returns invalid JSON.
- [ ] Validate weekly/monthly periods and goal IDs.

### Task 3: Frontend Data Utilities

**Files:**
- Modify: `frontend/src/api.js`
- Create: `frontend/src/roadmap.js`
- Test: `frontend/src/api.test.js`

- [ ] Add roadmap API client methods.
- [ ] Add pure helpers for graph layout, score labels, and chart points.
- [ ] Test helper behavior with empty and populated roadmaps.

### Task 4: Frontend Roadmap UI

**Files:**
- Modify: `frontend/src/App.jsx`
- Modify: `frontend/src/styles.css`

- [ ] Add tracking period selector and goal selector.
- [ ] Add clickable SVG roadmap graph.
- [ ] Add node details and daily task checkboxes.
- [ ] Add daily score card and simple chart.
- [ ] Improve mobile padding, grid stacking, and text wrapping.

### Task 5: Verification and Deploy

**Files:**
- No new files.

- [ ] Run `go test ./...`.
- [ ] Run `npm -C frontend test -- --run`.
- [ ] Run `npm -C . run build:frontend`.
- [ ] Rebuild backend binary and restart service when requested.
