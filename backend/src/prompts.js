export function roadmapPrompt(goals, entries, userPrompt) {
  return `
You are a personal execution coach.

Build a practical roadmap for the next 7 to 30 days based on the user's current goals and recent progress.

Requirements:
- Prioritize the most important goals.
- Break work into small milestones.
- Point out risks, blockers, and dependencies.
- End with a short "next 3 actions" list.
- Use concise markdown.

Current goals:
${JSON.stringify(goals, null, 2)}

Recent entries:
${JSON.stringify(entries, null, 2)}

Extra user instruction:
${userPrompt || "None"}
`.trim();
}

export function hintsPrompt(goals, entries, userPrompt) {
  return `
You are a strict but practical accountability assistant.

Give the user actionable hints for today's work.

Requirements:
- Keep the answer short and direct.
- Recommend what to focus on today.
- Mention one thing to avoid.
- Mention one quick win.
- Use concise markdown bullets.

Current goals:
${JSON.stringify(goals, null, 2)}

Recent entries:
${JSON.stringify(entries, null, 2)}

Extra user instruction:
${userPrompt || "None"}
`.trim();
}
