import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import Database from "better-sqlite3";

const currentDir = path.dirname(fileURLToPath(import.meta.url));
const dataDir = path.resolve(currentDir, "..", "data");
fs.mkdirSync(dataDir, { recursive: true });

const db = new Database(path.join(dataDir, "progress.db"));
db.pragma("journal_mode = WAL");

db.exec(`
  CREATE TABLE IF NOT EXISTS goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    period TEXT NOT NULL CHECK(period IN ('weekly', 'monthly')),
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    target_value INTEGER DEFAULT 1,
    current_value INTEGER DEFAULT 0,
    due_date TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'completed', 'paused')),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_date TEXT NOT NULL UNIQUE,
    summary TEXT DEFAULT '',
    wins TEXT DEFAULT '',
    blockers TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    mood TEXT DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
`);

const listGoalsStmt = db.prepare(`
  SELECT *
  FROM goals
  ORDER BY
    CASE status
      WHEN 'active' THEN 0
      WHEN 'paused' THEN 1
      ELSE 2
    END,
    due_date IS NULL,
    due_date ASC,
    created_at DESC
`);

const listEntriesStmt = db.prepare(`
  SELECT *
  FROM entries
  ORDER BY entry_date DESC
  LIMIT ?
`);

export function listGoals(period) {
  if (!period) {
    return listGoalsStmt.all();
  }

  return db
    .prepare("SELECT * FROM goals WHERE period = ? ORDER BY due_date IS NULL, due_date ASC, created_at DESC")
    .all(period);
}

export function getGoal(id) {
  return db.prepare("SELECT * FROM goals WHERE id = ?").get(id);
}

export function createGoal(goal) {
  const result = db
    .prepare(`
      INSERT INTO goals (period, title, description, target_value, current_value, due_date, status, updated_at)
      VALUES (@period, @title, @description, @target_value, @current_value, @due_date, @status, CURRENT_TIMESTAMP)
    `)
    .run(goal);

  return getGoal(result.lastInsertRowid);
}

export function updateGoal(id, goal) {
  db.prepare(`
    UPDATE goals
    SET
      period = @period,
      title = @title,
      description = @description,
      target_value = @target_value,
      current_value = @current_value,
      due_date = @due_date,
      status = @status,
      updated_at = CURRENT_TIMESTAMP
    WHERE id = @id
  `).run({ id, ...goal });

  return getGoal(id);
}

export function deleteGoal(id) {
  db.prepare("DELETE FROM goals WHERE id = ?").run(id);
}

export function listEntries(limit = 30) {
  return listEntriesStmt.all(limit);
}

export function getEntryByDate(entryDate) {
  return db.prepare("SELECT * FROM entries WHERE entry_date = ?").get(entryDate);
}

export function upsertEntry(entry) {
  const existing = getEntryByDate(entry.entry_date);

  if (existing) {
    db.prepare(`
      UPDATE entries
      SET
        summary = @summary,
        wins = @wins,
        blockers = @blockers,
        notes = @notes,
        mood = @mood,
        updated_at = CURRENT_TIMESTAMP
      WHERE entry_date = @entry_date
    `).run(entry);

    return getEntryByDate(entry.entry_date);
  }

  db.prepare(`
    INSERT INTO entries (entry_date, summary, wins, blockers, notes, mood, updated_at)
    VALUES (@entry_date, @summary, @wins, @blockers, @notes, @mood, CURRENT_TIMESTAMP)
  `).run(entry);

  return getEntryByDate(entry.entry_date);
}

export function getDashboard() {
  const goals = listGoals();
  const entries = listEntries(14);

  const byPeriod = ["weekly", "monthly"].map((period) => {
    const periodGoals = goals.filter((goal) => goal.period === period);
    const total = periodGoals.length;
    const completed = periodGoals.filter((goal) => goal.status === "completed").length;
    const progress = periodGoals.reduce((sum, goal) => {
      const target = Math.max(goal.target_value || 1, 1);
      const current = Math.min(goal.current_value || 0, target);
      return sum + current / target;
    }, 0);

    return {
      period,
      total,
      completed,
      averageCompletion: total ? Math.round((progress / total) * 100) : 0
    };
  });

  return {
    stats: byPeriod,
    goals,
    entries
  };
}
