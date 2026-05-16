package app

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir, err := ioutil.TempDir("", "progress-roadmap-test-*")
	if err != nil {
		t.Fatalf("TempDir() error = %v", err)
	}

	store, err := NewStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return store
}

func TestSaveRoadmapReturnsStoredGraphWithTasks(t *testing.T) {
	store := newTestStore(t)
	goal, err := store.CreateGoal(GoalInput{
		Period:      "weekly",
		Title:       "Launch study routine",
		Description: "Study math every day and review weak areas.",
		TargetValue: 7,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("CreateGoal() error = %v", err)
	}

	roadmap, err := store.SaveRoadmap(RoadmapInput{
		GoalID:    goal.ID,
		Period:    "weekly",
		StartDate: "2026-05-16",
		EndDate:   "2026-05-22",
		Nodes: []RoadmapNodeInput{
			{
				Title:           "Foundation",
				Description:     "Collect material and set baseline.",
				DayIndex:        1,
				DependsOn:       []int{},
				SuccessCriteria: "Baseline is written down.",
				Tasks: []RoadmapTaskInput{
					{TaskDate: "2026-05-16", Title: "List weak topics", Description: "Write the top three gaps."},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveRoadmap() error = %v", err)
	}

	if roadmap.ID == 0 {
		t.Fatal("SaveRoadmap() returned zero ID")
	}
	if len(roadmap.Nodes) != 1 {
		t.Fatalf("len(roadmap.Nodes) = %d, want 1", len(roadmap.Nodes))
	}
	if len(roadmap.Nodes[0].Tasks) != 1 {
		t.Fatalf("len(roadmap.Nodes[0].Tasks) = %d, want 1", len(roadmap.Nodes[0].Tasks))
	}
	if roadmap.Score.TaskCompletion != 0 {
		t.Fatalf("TaskCompletion = %d, want 0", roadmap.Score.TaskCompletion)
	}
}

func TestSetRoadmapTaskDoneUpdatesScore(t *testing.T) {
	store := newTestStore(t)
	goal, err := store.CreateGoal(GoalInput{
		Period:      "weekly",
		Title:       "Write every day",
		TargetValue: 2,
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("CreateGoal() error = %v", err)
	}
	roadmap, err := store.SaveRoadmap(RoadmapInput{
		GoalID:    goal.ID,
		Period:    "weekly",
		StartDate: "2026-05-16",
		EndDate:   "2026-05-17",
		Nodes: []RoadmapNodeInput{
			{
				Title:    "Draft",
				DayIndex: 1,
				Tasks: []RoadmapTaskInput{
					{TaskDate: "2026-05-16", Title: "Draft paragraph"},
					{TaskDate: "2026-05-17", Title: "Revise paragraph"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveRoadmap() error = %v", err)
	}

	updated, err := store.SetRoadmapTaskDone(roadmap.Nodes[0].Tasks[0].ID, true)
	if err != nil {
		t.Fatalf("SetRoadmapTaskDone() error = %v", err)
	}

	if !updated.Nodes[0].Tasks[0].Done {
		t.Fatal("task Done = false, want true")
	}
	if updated.Score.TaskCompletion != 50 {
		t.Fatalf("TaskCompletion = %d, want 50", updated.Score.TaskCompletion)
	}
	if updated.Score.Overall <= 0 {
		t.Fatalf("Overall score = %d, want positive", updated.Score.Overall)
	}
}

func TestFallbackRoadmapDecomposesDescriptionByPeriod(t *testing.T) {
	goal := Goal{
		ID:          9,
		Period:      "monthly",
		Title:       "Get fit",
		Description: "Build a consistent training habit.",
		TargetValue: 30,
	}

	weekly := fallbackRoadmapInput(goal, "weekly", "2026-05-16")
	monthly := fallbackRoadmapInput(goal, "monthly", "2026-05-16")

	if len(weekly.Nodes) != 7 {
		t.Fatalf("weekly nodes = %d, want 7", len(weekly.Nodes))
	}
	if len(monthly.Nodes) != 30 {
		t.Fatalf("monthly nodes = %d, want 30", len(monthly.Nodes))
	}
	if weekly.Nodes[0].Tasks[0].TaskDate != "2026-05-16" {
		t.Fatalf("first task date = %q, want 2026-05-16", weekly.Nodes[0].Tasks[0].TaskDate)
	}
}
