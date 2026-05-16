package app

type Goal struct {
	ID           int64   `json:"id"`
	Period       string  `json:"period"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	TargetValue  int     `json:"target_value"`
	CurrentValue int     `json:"current_value"`
	DueDate      *string `json:"due_date"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type Entry struct {
	ID        int64  `json:"id"`
	EntryDate string `json:"entry_date"`
	Summary   string `json:"summary"`
	Wins      string `json:"wins"`
	Blockers  string `json:"blockers"`
	Notes     string `json:"notes"`
	Mood      string `json:"mood"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Stat struct {
	Period            string `json:"period"`
	Total             int    `json:"total"`
	Completed         int    `json:"completed"`
	AverageCompletion int    `json:"averageCompletion"`
}

type Dashboard struct {
	Stats   []Stat  `json:"stats"`
	Goals   []Goal  `json:"goals"`
	Entries []Entry `json:"entries"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GoalInput struct {
	Period       string  `json:"period"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	TargetValue  int     `json:"target_value"`
	CurrentValue int     `json:"current_value"`
	DueDate      *string `json:"due_date"`
	Status       string  `json:"status"`
}

type EntryInput struct {
	EntryDate string `json:"entry_date"`
	Summary   string `json:"summary"`
	Wins      string `json:"wins"`
	Blockers  string `json:"blockers"`
	Notes     string `json:"notes"`
	Mood      string `json:"mood"`
}

type AIRequest struct {
	Prompt string `json:"prompt"`
}
