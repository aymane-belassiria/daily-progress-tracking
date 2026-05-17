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

type Roadmap struct {
	ID        int64         `json:"id"`
	GoalID    int64         `json:"goal_id"`
	Period    string        `json:"period"`
	StartDate string        `json:"start_date"`
	EndDate   string        `json:"end_date"`
	Nodes     []RoadmapNode `json:"nodes"`
	Score     RoadmapScore  `json:"score"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

type RoadmapNode struct {
	ID              int64         `json:"id"`
	RoadmapID       int64         `json:"roadmap_id"`
	Sequence        int           `json:"sequence"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	DayIndex        int           `json:"day_index"`
	DependsOn       []int         `json:"depends_on"`
	SuccessCriteria string        `json:"success_criteria"`
	Tasks           []RoadmapTask `json:"tasks"`
}

type RoadmapTask struct {
	ID          int64  `json:"id"`
	NodeID      int64  `json:"node_id"`
	TaskDate    string `json:"task_date"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

type RoadmapScore struct {
	Overall          int    `json:"overall"`
	TaskCompletion   int    `json:"task_completion"`
	GoalProgress     int    `json:"goal_progress"`
	EntryConsistency int    `json:"entry_consistency"`
	Diagnosis        string `json:"diagnosis"`
	NextAction       string `json:"next_action"`
}

type RoadmapInput struct {
	GoalID    int64               `json:"goal_id"`
	Period    string              `json:"period"`
	StartDate string              `json:"start_date"`
	EndDate   string              `json:"end_date"`
	Nodes     []RoadmapNodeInput  `json:"nodes"`
}

type RoadmapNodeInput struct {
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	DayIndex        int                 `json:"day_index"`
	DependsOn       []int               `json:"depends_on"`
	SuccessCriteria string              `json:"success_criteria"`
	Tasks           []RoadmapTaskInput  `json:"tasks"`
}

type RoadmapTaskInput struct {
	TaskDate    string `json:"task_date"`
	Title       string `json:"title"`
	Description string `json:"description"`
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

type RoadmapGenerateRequest struct {
	GoalID    int64  `json:"goal_id"`
	Period    string `json:"period"`
	StartDate string `json:"start_date"`
}

type RoadmapTaskUpdateRequest struct {
	Done bool `json:"done"`
}

type AdaptRequest struct {
	EntryDate string `json:"entry_date"`
	Blockers  string `json:"blockers"`
	Summary   string `json:"summary"`
}

type AdaptSuggestion struct {
	Analysis    string   `json:"analysis"`
	Suggestions []string `json:"suggestions"`
	Enhancement string   `json:"enhancement"`
}
