package app

import (
	"encoding/json"
	"fmt"
	"strings"
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
