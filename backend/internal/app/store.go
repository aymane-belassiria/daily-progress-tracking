package app

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db              *sql.DB
	listGoalsAll    *sql.Stmt
	listGoalsPeriod *sql.Stmt
	getGoalByID     *sql.Stmt
	insertGoal      *sql.Stmt
	updateGoal      *sql.Stmt
	deleteGoalStmt  *sql.Stmt
	listEntriesStmt *sql.Stmt
	getEntryByDate  *sql.Stmt
	insertEntry     *sql.Stmt
	updateEntry     *sql.Stmt
}

func NewStore(databasePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0755); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&cache=shared", databasePath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	store := &Store{db: db}
	if err := store.prepare(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS goals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			period TEXT NOT NULL,
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

		CREATE TABLE IF NOT EXISTS roadmaps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			goal_id INTEGER NOT NULL,
			period TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(goal_id) REFERENCES goals(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS roadmap_nodes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			roadmap_id INTEGER NOT NULL,
			sequence INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			day_index INTEGER NOT NULL,
			depends_on TEXT NOT NULL DEFAULT '[]',
			success_criteria TEXT DEFAULT '',
			FOREIGN KEY(roadmap_id) REFERENCES roadmaps(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS roadmap_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			node_id INTEGER NOT NULL,
			task_date TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			done INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY(node_id) REFERENCES roadmap_nodes(id) ON DELETE CASCADE
		);
	`)
	return err
}

func (s *Store) prepare() error {
	var err error
	if s.listGoalsAll, err = s.db.Prepare(`
		SELECT id, period, title, description, target_value, current_value, due_date, status, created_at, updated_at
		FROM goals
		ORDER BY
			CASE status WHEN 'active' THEN 0 WHEN 'paused' THEN 1 ELSE 2 END,
			due_date IS NULL,
			due_date ASC,
			created_at DESC
	`); err != nil {
		return err
	}
	if s.listGoalsPeriod, err = s.db.Prepare(`
		SELECT id, period, title, description, target_value, current_value, due_date, status, created_at, updated_at
		FROM goals
		WHERE period = ?
		ORDER BY due_date IS NULL, due_date ASC, created_at DESC
	`); err != nil {
		return err
	}
	if s.getGoalByID, err = s.db.Prepare(`
		SELECT id, period, title, description, target_value, current_value, due_date, status, created_at, updated_at
		FROM goals WHERE id = ?
	`); err != nil {
		return err
	}
	if s.insertGoal, err = s.db.Prepare(`
		INSERT INTO goals (period, title, description, target_value, current_value, due_date, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`); err != nil {
		return err
	}
	if s.updateGoal, err = s.db.Prepare(`
		UPDATE goals
		SET period = ?, title = ?, description = ?, target_value = ?, current_value = ?, due_date = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`); err != nil {
		return err
	}
	if s.deleteGoalStmt, err = s.db.Prepare(`DELETE FROM goals WHERE id = ?`); err != nil {
		return err
	}
	if s.listEntriesStmt, err = s.db.Prepare(`
		SELECT id, entry_date, summary, wins, blockers, notes, mood, created_at, updated_at
		FROM entries
		ORDER BY entry_date DESC
		LIMIT ?
	`); err != nil {
		return err
	}
	if s.getEntryByDate, err = s.db.Prepare(`
		SELECT id, entry_date, summary, wins, blockers, notes, mood, created_at, updated_at
		FROM entries WHERE entry_date = ?
	`); err != nil {
		return err
	}
	if s.insertEntry, err = s.db.Prepare(`
		INSERT INTO entries (entry_date, summary, wins, blockers, notes, mood, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`); err != nil {
		return err
	}
	if s.updateEntry, err = s.db.Prepare(`
		UPDATE entries
		SET summary = ?, wins = ?, blockers = ?, notes = ?, mood = ?, updated_at = CURRENT_TIMESTAMP
		WHERE entry_date = ?
	`); err != nil {
		return err
	}
	return nil
}

func (s *Store) Close() error {
	statements := []*sql.Stmt{
		s.listGoalsAll, s.listGoalsPeriod, s.getGoalByID, s.insertGoal, s.updateGoal,
		s.deleteGoalStmt, s.listEntriesStmt, s.getEntryByDate, s.insertEntry, s.updateEntry,
	}
	for _, stmt := range statements {
		if stmt != nil {
			_ = stmt.Close()
		}
	}
	return s.db.Close()
}

func (s *Store) ListGoals(period string) ([]Goal, error) {
	var rows *sql.Rows
	var err error
	if period == "" {
		rows, err = s.listGoalsAll.Query()
	} else {
		rows, err = s.listGoalsPeriod.Query(period)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []Goal
	for rows.Next() {
		goal, scanErr := scanGoal(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		goals = append(goals, goal)
	}
	if goals == nil {
		goals = []Goal{}
	}
	return goals, rows.Err()
}

func (s *Store) GetGoal(id int64) (*Goal, error) {
	row := s.getGoalByID.QueryRow(id)
	goal, err := scanGoal(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (s *Store) CreateGoal(input GoalInput) (*Goal, error) {
	result, err := s.insertGoal.Exec(
		input.Period, input.Title, input.Description, input.TargetValue,
		input.CurrentValue, nullableStringValue(input.DueDate), input.Status,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetGoal(id)
}

func (s *Store) UpdateGoal(id int64, input GoalInput) (*Goal, error) {
	result, err := s.updateGoal.Exec(
		input.Period, input.Title, input.Description, input.TargetValue,
		input.CurrentValue, nullableStringValue(input.DueDate), input.Status, id,
	)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, nil
	}
	return s.GetGoal(id)
}

func (s *Store) DeleteGoal(id int64) error {
	_, err := s.deleteGoalStmt.Exec(id)
	return err
}

func (s *Store) ListEntries(limit int) ([]Entry, error) {
	rows, err := s.listEntriesStmt.Query(limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		entry, scanErr := scanEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries = append(entries, entry)
	}
	if entries == nil {
		entries = []Entry{}
	}
	return entries, rows.Err()
}

func (s *Store) GetEntryByDate(entryDate string) (*Entry, error) {
	row := s.getEntryByDate.QueryRow(entryDate)
	entry, err := scanEntry(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *Store) UpsertEntry(input EntryInput) (*Entry, error) {
	current, err := s.GetEntryByDate(input.EntryDate)
	if err != nil {
		return nil, err
	}

	if current == nil {
		if _, err := s.insertEntry.Exec(input.EntryDate, input.Summary, input.Wins, input.Blockers, input.Notes, input.Mood); err != nil {
			return nil, err
		}
		return s.GetEntryByDate(input.EntryDate)
	}

	if _, err := s.updateEntry.Exec(input.Summary, input.Wins, input.Blockers, input.Notes, input.Mood, input.EntryDate); err != nil {
		return nil, err
	}
	return s.GetEntryByDate(input.EntryDate)
}

func (s *Store) Dashboard() (Dashboard, error) {
	goals, err := s.ListGoals("")
	if err != nil {
		return Dashboard{}, err
	}
	entries, err := s.ListEntries(14)
	if err != nil {
		return Dashboard{}, err
	}

	// Collect unique periods from goals
	periodSet := make(map[string]struct{})
	for _, goal := range goals {
		periodSet[goal.Period] = struct{}{}
	}
	// Always show weekly and monthly even if no goals exist for them
	for _, p := range []string{"weekly", "monthly"} {
		periodSet[p] = struct{}{}
	}
	periods := make([]string, 0, len(periodSet))
	for p := range periodSet {
		periods = append(periods, p)
	}
	sort.Strings(periods)

	stats := make([]Stat, 0, len(periods))
	for _, period := range periods {
		total := 0
		completed := 0
		progress := 0.0

		for _, goal := range goals {
			if goal.Period != period {
				continue
			}
			total++
			if goal.Status == "completed" {
				completed++
			}
			target := goal.TargetValue
			if target < 1 {
				target = 1
			}
			current := goal.CurrentValue
			if current > target {
				current = target
			}
			progress += float64(current) / float64(target)
		}

		average := 0
		if total > 0 {
			average = int(math.Round((progress / float64(total)) * 100))
		}
		stats = append(stats, Stat{
			Period:            period,
			Total:             total,
			Completed:         completed,
			AverageCompletion: average,
		})
	}

	return Dashboard{
		Stats:   stats,
		Goals:   goals,
		Entries: entries,
	}, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanGoal(s scanner) (Goal, error) {
	var goal Goal
	var dueDate sql.NullString
	err := s.Scan(
		&goal.ID, &goal.Period, &goal.Title, &goal.Description, &goal.TargetValue,
		&goal.CurrentValue, &dueDate, &goal.Status, &goal.CreatedAt, &goal.UpdatedAt,
	)
	if err != nil {
		return Goal{}, err
	}
	if dueDate.Valid {
		goal.DueDate = &dueDate.String
	}
	return goal, nil
}

func scanEntry(s scanner) (Entry, error) {
	var entry Entry
	err := s.Scan(
		&entry.ID, &entry.EntryDate, &entry.Summary, &entry.Wins, &entry.Blockers,
		&entry.Notes, &entry.Mood, &entry.CreatedAt, &entry.UpdatedAt,
	)
	if err != nil {
		return Entry{}, err
	}
	return entry, nil
}

func nullableStringValue(value *string) interface{} {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (id TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	if !migrationDone(db, "001_flexible_period") {
		if err := migration001FlexiblePeriod(db); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (id) VALUES (?)`, "001_flexible_period"); err != nil {
			return err
		}
	}
	return nil
}

func migrationDone(db *sql.DB, id string) bool {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE id = ?`, id).Scan(&n)
	return n > 0
}

func migration001FlexiblePeriod(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	steps := []string{
		`PRAGMA foreign_keys = OFF`,
		`CREATE TABLE goals_v2 (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			period TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			target_value INTEGER DEFAULT 1,
			current_value INTEGER DEFAULT 0,
			due_date TEXT,
			status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'completed', 'paused')),
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO goals_v2 SELECT id, period, title, description, target_value, current_value, due_date, status, created_at, updated_at FROM goals`,
		`DROP TABLE goals`,
		`ALTER TABLE goals_v2 RENAME TO goals`,
		`CREATE TABLE roadmaps_v2 (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			goal_id INTEGER NOT NULL,
			period TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(goal_id) REFERENCES goals(id) ON DELETE CASCADE
		)`,
		`INSERT INTO roadmaps_v2 SELECT id, goal_id, period, start_date, end_date, created_at, updated_at FROM roadmaps`,
		`DROP TABLE roadmaps`,
		`ALTER TABLE roadmaps_v2 RENAME TO roadmaps`,
		`PRAGMA foreign_keys = ON`,
	}
	for _, stmt := range steps {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration001 failed at %q: %w", stmt[:min(40, len(stmt))], err)
		}
	}
	return tx.Commit()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
