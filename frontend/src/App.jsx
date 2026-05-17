import { useEffect, useState } from "react";
import { api, getStoredToken, setStoredToken } from "./api";
import { buildGraphLayout, buildScoreTrend, scoreLabel, taskTotals } from "./roadmap";

const today = new Date().toISOString().slice(0, 10);

const PERIOD_LABELS = {
  daily: "Daily",
  weekly: "Weekly",
  "2-weeks": "2 Weeks",
  monthly: "Monthly",
  quarterly: "Quarterly",
  "6-months": "6 Months",
  yearly: "Yearly",
};

const emptyGoal = {
  period: "weekly",
  title: "",
  description: "",
  target_value: 1,
  current_value: 0,
  due_date: "",
  status: "active"
};

const emptyEntry = {
  entry_date: today,
  summary: "",
  wins: "",
  blockers: "",
  notes: "",
  mood: ""
};

function normalizeDashboard(data) {
  return {
    ...data,
    goals: Array.isArray(data?.goals) ? data.goals : [],
    stats: Array.isArray(data?.stats) ? data.stats : [],
    entries: Array.isArray(data?.entries) ? data.entries : []
  };
}

function normalizeRoadmaps(data) {
  return Array.isArray(data?.roadmaps) ? data.roadmaps : [];
}

function LoginScreen({ onLogin, error, busy }) {
  const [form, setForm] = useState({ email: "", password: "" });

  return (
    <div className="shell auth-shell">
      <div className="auth-card">
        <p className="eyebrow">Private console</p>
        <h1>Track execution, not intentions.</h1>
        <p className="muted">
          Single-user dashboard for weekly goals, monthly goals, daily notes, and AI-generated roadmap support.
        </p>
        <form
          className="stack"
          onSubmit={(event) => {
            event.preventDefault();
            onLogin(form);
          }}
        >
          <label>
            <span>Email</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) => setForm({ ...form, email: event.target.value })}
              required
            />
          </label>
          <label>
            <span>Password</span>
            <input
              type="password"
              value={form.password}
              onChange={(event) => setForm({ ...form, password: event.target.value })}
              required
            />
          </label>
          {error ? <p className="error">{error}</p> : null}
          <button type="submit" disabled={busy}>
            {busy ? "Signing in..." : "Sign in"}
          </button>
        </form>
      </div>
    </div>
  );
}

function StatCard({ label, value, subvalue }) {
  return (
    <div className="card stat-card">
      <p className="eyebrow">{label}</p>
      <h3>{value}</h3>
      <p className="muted">{subvalue}</p>
    </div>
  );
}

function GoalForm({ goal, onChange, onSubmit, submitLabel }) {
  return (
    <form
      className="card stack"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit();
      }}
    >
      <div className="inline-grid">
        <label>
          <span>Period</span>
          <select value={goal.period} onChange={(event) => onChange({ ...goal, period: event.target.value })}>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="2-weeks">2 Weeks</option>
            <option value="monthly">Monthly</option>
            <option value="quarterly">Quarterly (3 months)</option>
            <option value="6-months">6 Months</option>
            <option value="yearly">Yearly</option>
          </select>
        </label>
        <label>
          <span>Status</span>
          <select value={goal.status} onChange={(event) => onChange({ ...goal, status: event.target.value })}>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="paused">Paused</option>
          </select>
        </label>
      </div>
      <label>
        <span>Title</span>
        <input value={goal.title} onChange={(event) => onChange({ ...goal, title: event.target.value })} required />
      </label>
      <label>
        <span>Description</span>
        <textarea
          rows="3"
          value={goal.description}
          onChange={(event) => onChange({ ...goal, description: event.target.value })}
        />
      </label>
      <div className="inline-grid">
        <label>
          <span>Target</span>
          <input
            type="number"
            min="1"
            value={goal.target_value}
            onChange={(event) => onChange({ ...goal, target_value: event.target.value })}
            required
          />
        </label>
        <label>
          <span>Current</span>
          <input
            type="number"
            min="0"
            value={goal.current_value}
            onChange={(event) => onChange({ ...goal, current_value: event.target.value })}
            required
          />
        </label>
        <label>
          <span>Due date</span>
          <input
            type="date"
            value={goal.due_date || ""}
            onChange={(event) => onChange({ ...goal, due_date: event.target.value })}
          />
        </label>
      </div>
      <div className="form-submit-row">
        <button type="submit">{submitLabel}</button>
      </div>
    </form>
  );
}

function EntryForm({ entry, onChange, onSubmit }) {
  return (
    <form
      className="card stack"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit();
      }}
    >
      <div className="two-col-grid">
        <label>
          <span>Date</span>
          <input
            type="date"
            value={entry.entry_date}
            onChange={(event) => onChange({ ...entry, entry_date: event.target.value })}
          />
        </label>
        <label>
          <span>Mood</span>
          <input value={entry.mood} onChange={(event) => onChange({ ...entry, mood: event.target.value })} />
        </label>
      </div>
      <label>
        <span>Summary</span>
        <textarea rows="3" value={entry.summary} onChange={(event) => onChange({ ...entry, summary: event.target.value })} />
      </label>
      <label>
        <span>Wins</span>
        <textarea rows="3" value={entry.wins} onChange={(event) => onChange({ ...entry, wins: event.target.value })} />
      </label>
      <label>
        <span>Blockers</span>
        <textarea
          rows="3"
          value={entry.blockers}
          onChange={(event) => onChange({ ...entry, blockers: event.target.value })}
        />
      </label>
      <label>
        <span>Notes</span>
        <textarea rows="4" value={entry.notes} onChange={(event) => onChange({ ...entry, notes: event.target.value })} />
      </label>
      <button type="submit">Save daily entry</button>
    </form>
  );
}

function GoalList({ goals, onEdit, onDelete }) {
  return (
    <div className="stack">
      {goals.map((goal) => {
        const percent = Math.min(100, Math.round((goal.current_value / Math.max(goal.target_value, 1)) * 100));

        return (
          <article className="card goal-card" key={goal.id}>
            <div className="goal-topline">
              <div>
                <p className="eyebrow">{goal.period}</p>
                <h3>{goal.title}</h3>
              </div>
              <span className={`status-pill ${goal.status}`}>{goal.status}</span>
            </div>
            <p className="muted">{goal.description || "No description."}</p>
            <div className="goal-meter">
              <div className="goal-meter-fill" style={{ width: `${percent}%` }} />
            </div>
            <p className="muted">
              {goal.current_value}/{goal.target_value} complete {goal.due_date ? `by ${goal.due_date}` : ""}
            </p>
            <div className="button-row">
              <button type="button" className="secondary" onClick={() => onEdit(goal)}>
                Edit
              </button>
              <button type="button" className="danger" onClick={() => onDelete(goal.id)}>
                Delete
              </button>
            </div>
          </article>
        );
      })}
    </div>
  );
}

function EntryList({ entries }) {
  return (
    <div className="stack">
      {entries.map((entry) => (
        <article className="card" key={entry.id}>
          <div className="goal-topline">
            <div>
              <p className="eyebrow">{entry.entry_date}</p>
              <h3>{entry.summary || "No summary"}</h3>
            </div>
            <span className="status-pill neutral">{entry.mood || "No mood"}</span>
          </div>
          <p><strong>Wins:</strong> {entry.wins || "None"}</p>
          <p><strong>Blockers:</strong> {entry.blockers || "None"}</p>
          <p><strong>Notes:</strong> {entry.notes || "None"}</p>
        </article>
      ))}
    </div>
  );
}

function RoadmapGraph({ roadmap, selectedNodeId, onSelectNode }) {
  const [zoom, setZoom] = useState(1);
  const layout = buildGraphLayout(roadmap?.nodes || []);
  const byDayIndex = new Map(layout.nodes.map((node) => [node.day_index, node]));

  return (
    <div>
      <div className="zoom-controls">
        <button
          type="button"
          className="secondary zoom-btn"
          onClick={() => setZoom((z) => Math.max(0.5, parseFloat((z - 0.25).toFixed(2))))}
          disabled={zoom <= 0.5}
          aria-label="Zoom out"
        >
          −
        </button>
        <span className="zoom-label">{Math.round(zoom * 100)}%</span>
        <button
          type="button"
          className="secondary zoom-btn"
          onClick={() => setZoom((z) => Math.min(3, parseFloat((z + 0.25).toFixed(2))))}
          disabled={zoom >= 3}
          aria-label="Zoom in"
        >
          +
        </button>
      </div>
      <div className="roadmap-graph-wrap">
        <svg
          className="roadmap-graph"
          viewBox={`0 0 ${layout.width} ${layout.height}`}
          style={{ width: `${layout.width * zoom}px`, height: `${layout.height * zoom}px` }}
          role="img"
          aria-label="Roadmap graph"
        >
          {layout.edges.map((edge) => {
            const from = byDayIndex.get(edge.from);
            const to = byDayIndex.get(edge.to);
            if (!from || !to) {
              return null;
            }
            return <line key={`${edge.from}-${edge.to}`} x1={from.x} y1={from.y} x2={to.x} y2={to.y} />;
          })}
          {layout.nodes.map((node) => {
            const doneTasks = (node.tasks || []).filter((task) => task.done).length;
            const totalTasks = (node.tasks || []).length || 1;
            const isSelected = node.id === selectedNodeId;
            return (
              <g
                key={node.id}
                className={`roadmap-node ${isSelected ? "selected" : ""}`}
                transform={`translate(${node.x}, ${node.y})`}
                onClick={() => onSelectNode(node.id)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    onSelectNode(node.id);
                  }
                }}
                role="button"
                tabIndex="0"
              >
                <circle r="30" />
                <text y="-3">{node.day_index}</text>
                <text y="15" className="node-progress">
                  {doneTasks}/{totalTasks}
                </text>
              </g>
            );
          })}
        </svg>
      </div>
    </div>
  );
}

function ScoreChart({ points }) {
  const width = 360;
  const height = 120;
  const max = 100;
  const step = points.length > 1 ? width / (points.length - 1) : width;
  const polyline = points
    .map((point, index) => {
      const x = points.length > 1 ? index * step : width / 2;
      const y = height - (Math.max(0, Math.min(max, point.value)) / max) * (height - 20) - 10;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <div className="score-chart">
      <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Daily progress score chart">
        <polyline points={polyline} />
        {points.map((point, index) => {
          const x = points.length > 1 ? index * step : width / 2;
          const y = height - (Math.max(0, Math.min(max, point.value)) / max) * (height - 20) - 10;
          return <circle key={`${point.label}-${index}`} cx={x} cy={y} r="4" />;
        })}
      </svg>
      <div className="chart-labels">
        {points.map((point, index) => (
          <span key={`${point.label}-${index}`}>{point.label}</span>
        ))}
      </div>
    </div>
  );
}

function RoadmapPanel({
  goals,
  entries,
  roadmaps,
  trackingPeriod,
  selectedGoalId,
  selectedNodeId,
  busy,
  onSelectPeriod,
  onSelectGoal,
  onSelectNode,
  onGenerate,
  onToggleTask
}) {
  const selectedGoal = goals.find((goal) => goal.id === selectedGoalId) || goals[0];
  const roadmap = roadmaps.find((item) => item.goal_id === selectedGoal?.id && item.period === trackingPeriod);
  const selectedNode = roadmap?.nodes.find((node) => node.id === selectedNodeId) || roadmap?.nodes[0];
  const totals = taskTotals(roadmap);
  const trend = buildScoreTrend(entries, roadmap);

  return (
    <section className="roadmap-lab">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Roadmap lab</p>
          <h2>Graph plan and daily score</h2>
        </div>
        <div className="button-row">
          <button
            type="button"
            className={trackingPeriod === "weekly" ? "" : "secondary"}
            onClick={() => onSelectPeriod("weekly")}
          >
            Weekly
          </button>
          <button
            type="button"
            className={trackingPeriod === "monthly" ? "" : "secondary"}
            onClick={() => onSelectPeriod("monthly")}
          >
            Monthly
          </button>
        </div>
      </div>

      <div className="roadmap-controls">
        <label>
          <span>Goal</span>
          <select
            value={selectedGoal?.id || ""}
            onChange={(event) => onSelectGoal(Number(event.target.value))}
            disabled={goals.length === 0}
          >
            {goals.length === 0 ? <option value="">Create a goal first</option> : null}
            {goals.map((goal) => (
              <option key={goal.id} value={goal.id}>
                {goal.title}
              </option>
            ))}
          </select>
        </label>
        <button type="button" onClick={() => selectedGoal && onGenerate(selectedGoal.id)} disabled={!selectedGoal || busy}>
          {busy ? "Generating..." : roadmap ? "Regenerate graph" : "Generate graph"}
        </button>
      </div>

      {roadmap ? (
        <div className="roadmap-grid">
          <div className="roadmap-main">
            <RoadmapGraph roadmap={roadmap} selectedNodeId={selectedNode?.id} onSelectNode={onSelectNode} />
            <div className="task-strip">
              <span>{totals.done}/{totals.total} daily tasks done</span>
              <span>{roadmap.start_date} to {roadmap.end_date}</span>
            </div>
          </div>

          <aside className="roadmap-detail">
            <p className="eyebrow">Selected node</p>
            <h3>{selectedNode?.title || "No node selected"}</h3>
            <p className="muted">{selectedNode?.description || "Generate a roadmap to see the daily plan."}</p>
            <p><strong>Success:</strong> {selectedNode?.success_criteria || "Complete the listed tasks."}</p>
            <div className="stack">
              {(selectedNode?.tasks || []).map((task) => (
                <label className="task-check" key={task.id}>
                  <input
                    type="checkbox"
                    checked={task.done}
                    onChange={(event) => onToggleTask(task.id, event.target.checked)}
                  />
                  <span>
                    <strong>{task.task_date}: {task.title}</strong>
                    <small>{task.description || "No extra description."}</small>
                  </span>
                </label>
              ))}
            </div>
          </aside>

          <aside className="score-panel">
            <p className="eyebrow">Daily coach score</p>
            <div className="score-number">
              <strong>{roadmap.score.overall}</strong>
              <span>{scoreLabel(roadmap.score.overall)}</span>
            </div>
            <p className="muted">{roadmap.score.diagnosis}</p>
            <p><strong>Next action:</strong> {roadmap.score.next_action}</p>
            <div className="score-metrics">
              <span>Tasks {roadmap.score.task_completion}%</span>
              <span>Goal {roadmap.score.goal_progress}%</span>
              <span>Entries {roadmap.score.entry_consistency}%</span>
            </div>
            <ScoreChart points={trend} />
          </aside>
        </div>
      ) : (
        <div className="empty-roadmap">
          <p className="muted">Select a goal and generate a roadmap graph to decompose the description into daily tasks.</p>
        </div>
      )}
    </section>
  );
}

function HelpModal({ onClose }) {
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-box" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>How to use this app</h2>
          <button type="button" className="modal-close secondary" onClick={onClose}>
            Close
          </button>
        </div>
        <div className="modal-body stack">
          <div>
            <p className="eyebrow">Goals</p>
            <p>Create <strong>weekly</strong> or <strong>monthly</strong> goals with a target value and optional due date. Track progress by updating the current value as you advance.</p>
          </div>
          <div>
            <p className="eyebrow">Daily entries</p>
            <p>Log each day with a <strong>summary</strong>, <strong>wins</strong>, <strong>blockers</strong>, and notes. If you enter blockers, the AI will automatically analyze them and suggest adaptations to your roadmap.</p>
          </div>
          <div>
            <p className="eyebrow">Roadmap lab</p>
            <p>Select a goal and click <strong>Generate graph</strong> to get an AI-decomposed daily task plan. Click any node to see tasks for that day. Check off tasks as you complete them.</p>
          </div>
          <div>
            <p className="eyebrow">AI guidance</p>
            <p>Use the <strong>Generate roadmap</strong> and <strong>Generate hints</strong> buttons for free-form AI coaching based on your current goals and past entries.</p>
          </div>
          <div>
            <p className="eyebrow">Adaptive reasoning</p>
            <p>When you save a daily entry with blockers, the AI automatically detects obstacles and proposes enhancements. You can <strong>apply the enhancement</strong> (regenerate your roadmap with adapted context) or <strong>keep existing tasks</strong>.</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function AdaptationPanel({ adaptation, adaptBusy, selectedGoalId, trackingPeriod, onApply, onDismiss }) {
  if (adaptBusy) {
    return (
      <div className="adapt-panel adapt-busy">
        <p className="eyebrow">Reasoning</p>
        <p className="muted">Analyzing your blockers and generating an adaptive enhancement...</p>
      </div>
    );
  }
  if (!adaptation) return null;

  return (
    <div className="adapt-panel">
      <div className="adapt-header">
        <div>
          <p className="eyebrow">Adaptive reasoning</p>
          <h3>Enhancement detected</h3>
        </div>
        <button type="button" className="secondary" onClick={onDismiss}>
          Keep existing tasks
        </button>
      </div>
      <p className="adapt-analysis">{adaptation.analysis}</p>
      {adaptation.suggestions.length > 0 && (
        <ul className="adapt-suggestions">
          {adaptation.suggestions.map((s, i) => (
            <li key={i}>{s}</li>
          ))}
        </ul>
      )}
      <div className="adapt-enhancement">
        <p className="eyebrow">Proposed adaptation</p>
        <p>{adaptation.enhancement}</p>
      </div>
      <button type="button" onClick={onApply} disabled={!selectedGoalId}>
        Apply enhancement (regenerate roadmap)
      </button>
    </div>
  );
}

export default function App() {
  const [tokenReady, setTokenReady] = useState(Boolean(getStoredToken()));
  const [authBusy, setAuthBusy] = useState(false);
  const [authError, setAuthError] = useState("");
  const [dashboard, setDashboard] = useState(null);
  const [roadmaps, setRoadmaps] = useState([]);
  const [trackingPeriod, setTrackingPeriod] = useState("weekly");
  const [selectedGoalId, setSelectedGoalId] = useState(null);
  const [selectedNodeId, setSelectedNodeId] = useState(null);
  const [roadmapBusy, setRoadmapBusy] = useState(false);
  const [goalForm, setGoalForm] = useState(emptyGoal);
  const [goalEditingId, setGoalEditingId] = useState(null);
  const [entryForm, setEntryForm] = useState(emptyEntry);
  const [aiPrompt, setAiPrompt] = useState("");
  const [roadmap, setRoadmap] = useState("");
  const [hints, setHints] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [showHelp, setShowHelp] = useState(false);
  const [adaptation, setAdaptation] = useState(null);
  const [adaptBusy, setAdaptBusy] = useState(false);

  async function loadDashboard() {
    const [dashboardData, roadmapData] = await Promise.all([api.dashboard(), api.roadmaps()]);
    const data = normalizeDashboard(dashboardData);
    setDashboard(data);
    const nextRoadmaps = normalizeRoadmaps(roadmapData);
    setRoadmaps(nextRoadmaps);

    if (!selectedGoalId && data.goals.length > 0) {
      setSelectedGoalId(data.goals[0].id);
    }
    if (!selectedNodeId && nextRoadmaps[0]?.nodes?.length > 0) {
      setSelectedNodeId(nextRoadmaps[0].nodes[0].id);
    }

    const todayEntry = data.entries.find((entry) => entry.entry_date === today);
    if (todayEntry) {
      setEntryForm(todayEntry);
    }
  }

  useEffect(() => {
    if (!tokenReady) {
      return;
    }

    api
      .me()
      .then(loadDashboard)
      .catch(() => {
        setStoredToken("");
        setTokenReady(false);
      });
  }, [tokenReady]);

  async function handleLogin(payload) {
    setAuthBusy(true);
    setAuthError("");

    try {
      const data = await api.login(payload);
      setStoredToken(data.token);
      setTokenReady(true);
    } catch (error) {
      setAuthError(error.message);
    } finally {
      setAuthBusy(false);
    }
  }

  async function handleSaveGoal() {
    setMessage("");
    setError("");

    try {
      const payload = {
        ...goalForm,
        target_value: Number(goalForm.target_value),
        current_value: Number(goalForm.current_value),
        due_date: goalForm.due_date || null
      };

      await api.saveGoal(payload, goalEditingId);
      setGoalForm(emptyGoal);
      setGoalEditingId(null);
      setMessage("Goal saved.");
      await loadDashboard();
    } catch (nextError) {
      setError(nextError.message);
    }
  }

  async function handleDeleteGoal(id) {
    setMessage("");
    setError("");

    try {
      await api.deleteGoal(id);
      setMessage("Goal deleted.");
      await loadDashboard();
    } catch (nextError) {
      setError(nextError.message);
    }
  }

  async function handleSaveEntry() {
    setMessage("");
    setError("");

    try {
      await api.saveEntry(entryForm);
      setMessage("Daily entry saved.");
      await loadDashboard();

      if (entryForm.blockers?.trim()) {
        setAdaptation(null);
        setAdaptBusy(true);
        try {
          const result = await api.adapt({
            entry_date: entryForm.entry_date,
            blockers: entryForm.blockers,
            summary: entryForm.summary
          });
          setAdaptation(result);
        } catch {
          // adaptation is best-effort
        } finally {
          setAdaptBusy(false);
        }
      }
    } catch (nextError) {
      setError(nextError.message);
    }
  }

  async function handleGenerateRoadmap() {
    setMessage("");
    setError("");

    try {
      const data = await api.generateRoadmap(aiPrompt);
      setRoadmap(data.content);
    } catch (nextError) {
      setError(nextError.message);
    }
  }

  async function handleGenerateHints() {
    setMessage("");
    setError("");

    try {
      const data = await api.generateHints(aiPrompt);
      setHints(data.content);
    } catch (nextError) {
      setError(nextError.message);
    }
  }

  async function handleGenerateRoadmapGraph(goalId) {
    setMessage("");
    setError("");
    setRoadmapBusy(true);

    try {
      const data = await api.generateRoadmapGraph({
        goal_id: goalId,
        period: trackingPeriod,
        start_date: today
      });
      const roadmap = data.roadmap;
      setRoadmaps((current) => [
        roadmap,
        ...current.filter((item) => !(item.goal_id === roadmap.goal_id && item.period === roadmap.period))
      ]);
      setSelectedNodeId(roadmap.nodes?.[0]?.id || null);
      setMessage("Roadmap graph generated.");
    } catch (nextError) {
      setError(nextError.message);
    } finally {
      setRoadmapBusy(false);
    }
  }

  async function handleToggleRoadmapTask(taskId, done) {
    setMessage("");
    setError("");

    try {
      const data = await api.updateRoadmapTask(taskId, done);
      const roadmap = data.roadmap;
      setRoadmaps((current) => current.map((item) => (item.id === roadmap.id ? roadmap : item)));
    } catch (nextError) {
      setError(nextError.message);
    }
  }

  async function handleApplyAdaptation() {
    if (!selectedGoalId) return;
    setAdaptation(null);
    setMessage("Applying enhancement — regenerating roadmap...");
    setRoadmapBusy(true);
    setError("");
    try {
      const data = await api.generateRoadmapGraph({
        goal_id: selectedGoalId,
        period: trackingPeriod,
        start_date: today
      });
      const newRoadmap = data.roadmap;
      setRoadmaps((current) => [
        newRoadmap,
        ...current.filter((item) => !(item.goal_id === newRoadmap.goal_id && item.period === newRoadmap.period))
      ]);
      setSelectedNodeId(newRoadmap.nodes?.[0]?.id || null);
      setMessage("Enhancement applied — roadmap regenerated.");
    } catch (nextError) {
      setError(nextError.message);
    } finally {
      setRoadmapBusy(false);
    }
  }

  function logout() {
    setStoredToken("");
    setTokenReady(false);
    setDashboard(null);
  }

  if (!tokenReady) {
    return <LoginScreen onLogin={handleLogin} error={authError} busy={authBusy} />;
  }

  if (!dashboard) {
    return (
      <div className="shell">
        <p>Loading dashboard...</p>
      </div>
    );
  }

  const entryCount = dashboard.entries.length;
  const nonEmptyStats = dashboard.stats.filter((s) => s.total > 0);
  const displayStats = nonEmptyStats.length > 0
    ? nonEmptyStats.slice(0, 2)
    : dashboard.stats.filter((s) => s.period === "weekly" || s.period === "monthly");

  return (
    <div className="shell">
      {showHelp && <HelpModal onClose={() => setShowHelp(false)} />}
      <header className="hero">
        <div>
          <p className="eyebrow">Daily progress tracking</p>
          <h1>Execution dashboard</h1>
          <p className="muted">
            Weekly and monthly goals, daily notes, and Qwen-powered adaptive guidance routed through your private backend.
          </p>
        </div>
        <div className="button-row">
          <button type="button" className="help-btn secondary" onClick={() => setShowHelp(true)} title="How to use this app">
            ?
          </button>
          <button type="button" className="secondary" onClick={logout}>
            Log out
          </button>
        </div>
      </header>

      <section className="stats-grid">
        {displayStats.map((stat) => (
          <StatCard
            key={stat.period}
            label={PERIOD_LABELS[stat.period] ?? (stat.period.charAt(0).toUpperCase() + stat.period.slice(1))}
            value={`${stat.averageCompletion}%`}
            subvalue={`${stat.completed}/${stat.total} completed`}
          />
        ))}
        <StatCard label="Entries" value={entryCount} subvalue="Recent daily logs" />
      </section>

      {message ? <p className="message">{message}</p> : null}
      {error ? <p className="error">{error}</p> : null}

      <AdaptationPanel
        adaptation={adaptation}
        adaptBusy={adaptBusy}
        selectedGoalId={selectedGoalId}
        trackingPeriod={trackingPeriod}
        onApply={handleApplyAdaptation}
        onDismiss={() => setAdaptation(null)}
      />

      <RoadmapPanel
        goals={dashboard.goals}
        entries={dashboard.entries}
        roadmaps={roadmaps}
        trackingPeriod={trackingPeriod}
        selectedGoalId={selectedGoalId}
        selectedNodeId={selectedNodeId}
        busy={roadmapBusy}
        onSelectPeriod={(period) => {
          setTrackingPeriod(period);
          setSelectedNodeId(null);
        }}
        onSelectGoal={(goalId) => {
          setSelectedGoalId(goalId);
          setSelectedNodeId(null);
        }}
        onSelectNode={setSelectedNodeId}
        onGenerate={handleGenerateRoadmapGraph}
        onToggleTask={handleToggleRoadmapTask}
      />

      <section className="layout-grid">
        <div className="stack">
          <GoalForm
            goal={goalForm}
            onChange={setGoalForm}
            onSubmit={handleSaveGoal}
            submitLabel={goalEditingId ? "Update goal" : "Create goal"}
          />

          <div className="split-panel">
            {(() => {
              const goalPeriods = ["weekly", "monthly", ...dashboard.goals
                .map((g) => g.period)
                .filter((p, i, arr) => p !== "weekly" && p !== "monthly" && arr.indexOf(p) === i)
              ];
              return goalPeriods.map((period) => {
                const periodGoals = dashboard.goals.filter((g) => g.period === period);
                const label = PERIOD_LABELS[period] ?? (period.charAt(0).toUpperCase() + period.slice(1));
                return (
                  <div key={period}>
                    <p className="eyebrow section-label">{label} goals</p>
                    {periodGoals.length === 0 ? (
                      <p className="muted">No {label.toLowerCase()} goals yet.</p>
                    ) : (
                      <GoalList
                        goals={periodGoals}
                        onEdit={(goal) => {
                          setGoalEditingId(goal.id);
                          setGoalForm({
                            ...goal,
                            due_date: goal.due_date || ""
                          });
                        }}
                        onDelete={handleDeleteGoal}
                      />
                    )}
                  </div>
                );
              });
            })()}
          </div>
        </div>

        <div className="stack">
          <EntryForm entry={entryForm} onChange={setEntryForm} onSubmit={handleSaveEntry} />

          <div className="card stack">
            <p className="eyebrow">AI guidance</p>
            <label>
              <span>Extra context for the model</span>
              <textarea rows="4" value={aiPrompt} onChange={(event) => setAiPrompt(event.target.value)} />
            </label>
            <div className="button-row">
              <button type="button" onClick={handleGenerateRoadmap}>
                Generate roadmap
              </button>
              <button type="button" className="secondary" onClick={handleGenerateHints}>
                Generate hints
              </button>
            </div>
            {roadmap ? <pre className="ai-output">{roadmap}</pre> : null}
            {hints ? <pre className="ai-output">{hints}</pre> : null}
          </div>

          <div>
            <p className="eyebrow section-label">Recent entries</p>
            <EntryList entries={dashboard.entries} />
          </div>
        </div>
      </section>
    </div>
  );
}
