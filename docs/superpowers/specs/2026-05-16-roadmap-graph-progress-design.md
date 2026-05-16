# Roadmap Graph Progress Design

## Goal

Add goal-attached roadmaps that decompose a goal description into daily tasks, show those tasks as a clickable graph, and summarize progress with a daily score and chart.

## Product Shape

Existing goals remain the main records. A user can choose a weekly or monthly tracking mode, select a goal, and generate a structured roadmap for that goal. The roadmap contains graph nodes, daily tasks, dependencies, success criteria, and coach-style guidance.

## Data Flow

1. User creates or selects a goal.
2. User chooses weekly or monthly tracking.
3. Frontend asks the backend to generate a roadmap for that goal.
4. Backend calls Qwen and requests strict JSON.
5. Backend stores the roadmap and task completion state in SQLite.
6. Frontend renders nodes as a graph and opens details for the selected node.
7. User marks daily tasks complete.
8. Backend recomputes progress score and dashboard data.

## UX

- Add a Roadmap lab panel to the dashboard.
- Weekly/monthly tracking mode controls the generation horizon.
- Graph nodes are clickable and show details in a side panel.
- Daily tasks are grouped by node and can be checked off.
- Progress score combines task completion, goal progress, entry consistency, and blockers.
- A simple chart shows recent daily progress.
- Mobile stacks graph, details, score, and task list vertically with tighter padding.

## Implementation Boundaries

- Use plain SVG for the first graph version.
- Keep AI key server-side.
- Keep existing goal CRUD intact.
- Store AI output as structured JSON, with server-side fallback if AI returns invalid JSON.
- Avoid adding graph/chart dependencies for the first version.
