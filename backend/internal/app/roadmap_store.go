package app

import (
	"database/sql"
	"encoding/json"
	"math"
	"strings"
)

func (s *Store) SaveRoadmap(input RoadmapInput) (*Roadmap, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM roadmaps WHERE goal_id = ? AND period = ?`, input.GoalID, input.Period); err != nil {
		return nil, err
	}
	result, err := tx.Exec(`
		INSERT INTO roadmaps (goal_id, period, start_date, end_date, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, input.GoalID, input.Period, input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}
	roadmapID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for index, node := range input.Nodes {
		dependsOn, err := json.Marshal(node.DependsOn)
		if err != nil {
			return nil, err
		}
		nodeResult, err := tx.Exec(`
			INSERT INTO roadmap_nodes (roadmap_id, sequence, title, description, day_index, depends_on, success_criteria)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, roadmapID, index+1, node.Title, node.Description, node.DayIndex, string(dependsOn), node.SuccessCriteria)
		if err != nil {
			return nil, err
		}
		nodeID, err := nodeResult.LastInsertId()
		if err != nil {
			return nil, err
		}
		for _, task := range node.Tasks {
			if _, err := tx.Exec(`
				INSERT INTO roadmap_tasks (node_id, task_date, title, description)
				VALUES (?, ?, ?, ?)
			`, nodeID, task.TaskDate, task.Title, task.Description); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetRoadmap(roadmapID)
}

func (s *Store) ListRoadmaps() ([]Roadmap, error) {
	rows, err := s.db.Query(`
		SELECT id
		FROM roadmaps
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roadmaps := []Roadmap{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		roadmap, err := s.GetRoadmap(id)
		if err != nil {
			return nil, err
		}
		if roadmap != nil {
			roadmaps = append(roadmaps, *roadmap)
		}
	}
	return roadmaps, rows.Err()
}

func (s *Store) GetRoadmap(id int64) (*Roadmap, error) {
	row := s.db.QueryRow(`
		SELECT id, goal_id, period, start_date, end_date, created_at, updated_at
		FROM roadmaps
		WHERE id = ?
	`, id)

	var roadmap Roadmap
	if err := row.Scan(
		&roadmap.ID, &roadmap.GoalID, &roadmap.Period, &roadmap.StartDate,
		&roadmap.EndDate, &roadmap.CreatedAt, &roadmap.UpdatedAt,
	); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	nodes, err := s.listRoadmapNodes(roadmap.ID)
	if err != nil {
		return nil, err
	}
	roadmap.Nodes = nodes
	roadmap.Score = s.scoreRoadmap(roadmap)
	return &roadmap, nil
}

func (s *Store) SetRoadmapTaskDone(taskID int64, done bool) (*Roadmap, error) {
	value := 0
	if done {
		value = 1
	}
	var roadmapID int64
	row := s.db.QueryRow(`
		SELECT rn.roadmap_id
		FROM roadmap_tasks rt
		JOIN roadmap_nodes rn ON rn.id = rt.node_id
		WHERE rt.id = ?
	`, taskID)
	if err := row.Scan(&roadmapID); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(`UPDATE roadmap_tasks SET done = ? WHERE id = ?`, value, taskID); err != nil {
		return nil, err
	}
	return s.GetRoadmap(roadmapID)
}

func (s *Store) listRoadmapNodes(roadmapID int64) ([]RoadmapNode, error) {
	rows, err := s.db.Query(`
		SELECT id, roadmap_id, sequence, title, description, day_index, depends_on, success_criteria
		FROM roadmap_nodes
		WHERE roadmap_id = ?
		ORDER BY sequence ASC
	`, roadmapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := []RoadmapNode{}
	for rows.Next() {
		var node RoadmapNode
		var dependsOn string
		if err := rows.Scan(
			&node.ID, &node.RoadmapID, &node.Sequence, &node.Title, &node.Description,
			&node.DayIndex, &dependsOn, &node.SuccessCriteria,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(dependsOn), &node.DependsOn); err != nil {
			node.DependsOn = []int{}
		}
		tasks, err := s.listRoadmapTasks(node.ID)
		if err != nil {
			return nil, err
		}
		node.Tasks = tasks
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *Store) listRoadmapTasks(nodeID int64) ([]RoadmapTask, error) {
	rows, err := s.db.Query(`
		SELECT id, node_id, task_date, title, description, done
		FROM roadmap_tasks
		WHERE node_id = ?
		ORDER BY task_date ASC, id ASC
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []RoadmapTask{}
	for rows.Next() {
		var task RoadmapTask
		var done int
		if err := rows.Scan(&task.ID, &task.NodeID, &task.TaskDate, &task.Title, &task.Description, &done); err != nil {
			return nil, err
		}
		task.Done = done == 1
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *Store) scoreRoadmap(roadmap Roadmap) RoadmapScore {
	totalTasks := 0
	doneTasks := 0
	for _, node := range roadmap.Nodes {
		for _, task := range node.Tasks {
			totalTasks++
			if task.Done {
				doneTasks++
			}
		}
	}

	taskCompletion := 0
	if totalTasks > 0 {
		taskCompletion = int(math.Round((float64(doneTasks) / float64(totalTasks)) * 100))
	}

	goalProgress := 0
	if goal, err := s.GetGoal(roadmap.GoalID); err == nil && goal != nil {
		target := goal.TargetValue
		if target < 1 {
			target = 1
		}
		current := goal.CurrentValue
		if current > target {
			current = target
		}
		goalProgress = int(math.Round((float64(current) / float64(target)) * 100))
	}

	entryConsistency := s.entryConsistency(roadmap.StartDate, roadmap.EndDate)
	overall := int(math.Round(float64(taskCompletion)*0.65 + float64(goalProgress)*0.25 + float64(entryConsistency)*0.10))
	return RoadmapScore{
		Overall:          overall,
		TaskCompletion:   taskCompletion,
		GoalProgress:     goalProgress,
		EntryConsistency: entryConsistency,
		Diagnosis:        scoreDiagnosis(overall),
		NextAction:       nextRoadmapAction(roadmap),
	}
}

func (s *Store) entryConsistency(startDate, endDate string) int {
	rows, err := s.db.Query(`
		SELECT COUNT(*)
		FROM entries
		WHERE entry_date >= ? AND entry_date <= ?
	`, startDate, endDate)
	if err != nil {
		return 0
	}
	defer rows.Close()

	count := 0
	if rows.Next() {
		_ = rows.Scan(&count)
	}
	if count <= 0 {
		return 0
	}
	return 100
}

func scoreDiagnosis(score int) string {
	switch {
	case score >= 80:
		return "Strong execution rhythm."
	case score >= 50:
		return "Progress is moving, but consistency needs attention."
	case score > 0:
		return "Momentum is starting; protect the next small task."
	default:
		return "No measurable progress yet."
	}
}

func nextRoadmapAction(roadmap Roadmap) string {
	for _, node := range roadmap.Nodes {
		for _, task := range node.Tasks {
			if !task.Done {
				return strings.TrimSpace(task.Title)
			}
		}
	}
	return "Review the roadmap and raise the goal target if this is complete."
}
