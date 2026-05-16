import { useEffect, useState } from "react";
import { api, getStoredToken, setStoredToken } from "./api";

const today = new Date().toISOString().slice(0, 10);

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
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
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
      <button type="submit">{submitLabel}</button>
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
      <div className="inline-grid">
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

export default function App() {
  const [tokenReady, setTokenReady] = useState(Boolean(getStoredToken()));
  const [authBusy, setAuthBusy] = useState(false);
  const [authError, setAuthError] = useState("");
  const [dashboard, setDashboard] = useState(null);
  const [goalForm, setGoalForm] = useState(emptyGoal);
  const [goalEditingId, setGoalEditingId] = useState(null);
  const [entryForm, setEntryForm] = useState(emptyEntry);
  const [aiPrompt, setAiPrompt] = useState("");
  const [roadmap, setRoadmap] = useState("");
  const [hints, setHints] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  async function loadDashboard() {
    const data = await api.dashboard();
    setDashboard(data);

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

  const weekly = dashboard.goals.filter((goal) => goal.period === "weekly");
  const monthly = dashboard.goals.filter((goal) => goal.period === "monthly");
  const weeklyStats = dashboard.stats.find((item) => item.period === "weekly");
  const monthlyStats = dashboard.stats.find((item) => item.period === "monthly");

  return (
    <div className="shell">
      <header className="hero">
        <div>
          <p className="eyebrow">Daily progress tracking</p>
          <h1>Execution dashboard</h1>
          <p className="muted">
            Weekly and monthly goals, daily notes, and Qwen-generated guidance routed through your private backend.
          </p>
        </div>
        <button type="button" className="secondary" onClick={logout}>
          Log out
        </button>
      </header>

      <section className="stats-grid">
        <StatCard
          label="Weekly"
          value={`${weeklyStats?.averageCompletion || 0}%`}
          subvalue={`${weeklyStats?.completed || 0}/${weeklyStats?.total || 0} completed`}
        />
        <StatCard
          label="Monthly"
          value={`${monthlyStats?.averageCompletion || 0}%`}
          subvalue={`${monthlyStats?.completed || 0}/${monthlyStats?.total || 0} completed`}
        />
        <StatCard label="Entries" value={dashboard.entries.length} subvalue="Recent daily logs" />
      </section>

      {message ? <p className="message">{message}</p> : null}
      {error ? <p className="error">{error}</p> : null}

      <section className="layout-grid">
        <div className="stack">
          <GoalForm
            goal={goalForm}
            onChange={setGoalForm}
            onSubmit={handleSaveGoal}
            submitLabel={goalEditingId ? "Update goal" : "Create goal"}
          />

          <div className="split-panel">
            <div>
              <p className="eyebrow">Weekly goals</p>
              <GoalList
                goals={weekly}
                onEdit={(goal) => {
                  setGoalEditingId(goal.id);
                  setGoalForm({
                    ...goal,
                    due_date: goal.due_date || ""
                  });
                }}
                onDelete={handleDeleteGoal}
              />
            </div>
            <div>
              <p className="eyebrow">Monthly goals</p>
              <GoalList
                goals={monthly}
                onEdit={(goal) => {
                  setGoalEditingId(goal.id);
                  setGoalForm({
                    ...goal,
                    due_date: goal.due_date || ""
                  });
                }}
                onDelete={handleDeleteGoal}
              />
            </div>
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
            <p className="eyebrow">Recent entries</p>
            <EntryList entries={dashboard.entries} />
          </div>
        </div>
      </section>
    </div>
  );
}
