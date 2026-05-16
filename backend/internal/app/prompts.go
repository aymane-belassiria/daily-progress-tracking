package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func roadmapPrompt(goals []Goal, entries []Entry, userPrompt string) string {
	return fmt.Sprintf(`You are a personal execution coach.

Build a practical roadmap for the next 7 to 30 days based on the user's current goals and recent progress.

Requirements:
- Prioritize the most important goals.
- Break work into small milestones.
- Point out risks, blockers, and dependencies.
- End with a short "next 3 actions" list.
- Use concise markdown.

Current goals:
%s

Recent entries:
%s

Extra user instruction:
%s`, mustJSON(goals), mustJSON(entries), fallbackText(userPrompt))
}

func hintsPrompt(goals []Goal, entries []Entry, userPrompt string) string {
	return fmt.Sprintf(`You are a strict but practical accountability assistant.

Give the user actionable hints for today's work.

Requirements:
- Keep the answer short and direct.
- Recommend what to focus on today.
- Mention one thing to avoid.
- Mention one quick win.
- Use concise markdown bullets.

Current goals:
%s

Recent entries:
%s

Extra user instruction:
%s`, mustJSON(goals), mustJSON(entries), fallbackText(userPrompt))
}

func structuredRoadmapPrompt(goal Goal, period, startDate string) string {
	days := roadmapDays(period)
	return fmt.Sprintf(`You are an execution scientist and practical coach.

Return strict JSON only. Do not wrap it in markdown.

Schema:
{
  "nodes": [
    {
      "title": "short node title",
      "description": "why this node matters",
      "day_index": 1,
      "depends_on": [],
      "success_criteria": "observable completion signal",
      "tasks": [
        {
          "task_date": "YYYY-MM-DD",
          "title": "daily action",
          "description": "specific task instructions"
        }
      ]
    }
  ]
}

Rules:
- Create exactly %d nodes for a %s plan.
- Each node represents one day.
- Each node has one to three concrete daily tasks.
- Use dates starting at %s.
- Decompose the goal description into behavior-sized actions.
- Keep task titles short and measurable.

Goal:
%s`, days, period, startDate, mustJSON(goal))
}

func parseRoadmapInput(goal Goal, period, startDate, raw string) RoadmapInput {
	var parsed struct {
		Nodes []RoadmapNodeInput `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil || len(parsed.Nodes) == 0 {
		return fallbackRoadmapInput(goal, period, startDate)
	}

	input := fallbackRoadmapInput(goal, period, startDate)
	input.Nodes = normalizeRoadmapNodes(parsed.Nodes, startDate, roadmapDays(period))
	input.EndDate = roadmapEndDate(startDate, len(input.Nodes))
	return input
}

func fallbackRoadmapInput(goal Goal, period, startDate string) RoadmapInput {
	days := roadmapDays(period)
	nodes := make([]RoadmapNodeInput, 0, days)
	for index := 0; index < days; index++ {
		taskDate := addDays(startDate, index)
		dependsOn := []int{}
		if index > 0 {
			dependsOn = []int{index}
		}
		nodes = append(nodes, RoadmapNodeInput{
			Title:           fmt.Sprintf("Day %d: %s", index+1, compactTitle(goal.Title)),
			Description:     fallbackNodeDescription(goal),
			DayIndex:        index + 1,
			DependsOn:       dependsOn,
			SuccessCriteria: "The daily task is completed and reflected in the progress entry.",
			Tasks: []RoadmapTaskInput{
				{
					TaskDate:    taskDate,
					Title:       fmt.Sprintf("Complete one focused action for %s", compactTitle(goal.Title)),
					Description: fallbackTaskDescription(goal),
				},
			},
		})
	}
	return RoadmapInput{
		GoalID:    goal.ID,
		Period:    period,
		StartDate: startDate,
		EndDate:   roadmapEndDate(startDate, days),
		Nodes:     nodes,
	}
}

func normalizeRoadmapNodes(nodes []RoadmapNodeInput, startDate string, maxDays int) []RoadmapNodeInput {
	if len(nodes) > maxDays {
		nodes = nodes[:maxDays]
	}
	normalized := make([]RoadmapNodeInput, 0, len(nodes))
	for index, node := range nodes {
		if strings.TrimSpace(node.Title) == "" {
			node.Title = fmt.Sprintf("Day %d", index+1)
		}
		node.DayIndex = index + 1
		if node.DependsOn == nil {
			node.DependsOn = []int{}
		}
		if len(node.Tasks) == 0 {
			node.Tasks = []RoadmapTaskInput{{
				TaskDate: addDays(startDate, index),
				Title:    node.Title,
			}}
		}
		for taskIndex := range node.Tasks {
			if strings.TrimSpace(node.Tasks[taskIndex].TaskDate) == "" {
				node.Tasks[taskIndex].TaskDate = addDays(startDate, index)
			}
			if strings.TrimSpace(node.Tasks[taskIndex].Title) == "" {
				node.Tasks[taskIndex].Title = node.Title
			}
		}
		normalized = append(normalized, node)
	}
	return normalized
}

func roadmapDays(period string) int {
	if period == "monthly" {
		return 30
	}
	return 7
}

func roadmapEndDate(startDate string, days int) string {
	if days < 1 {
		days = 1
	}
	return addDays(startDate, days-1)
}

func addDays(startDate string, offset int) string {
	parsed, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		parsed = time.Now()
	}
	return parsed.AddDate(0, 0, offset).Format("2006-01-02")
}

func compactTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "the goal"
	}
	if len(title) > 48 {
		return title[:48]
	}
	return title
}

func fallbackNodeDescription(goal Goal) string {
	if strings.TrimSpace(goal.Description) != "" {
		return goal.Description
	}
	return "Build measurable progress on this goal through one daily action."
}

func fallbackTaskDescription(goal Goal) string {
	if strings.TrimSpace(goal.Description) != "" {
		return "Use the goal description as context: " + goal.Description
	}
	return "Do one concrete task that makes the goal easier to finish."
}

func mustJSON(value interface{}) string {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func fallbackText(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return value
}
