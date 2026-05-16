package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	config *Config
	store  *Store
	ai     *nvidiaClient
}

func NewServer(cfg Config, store *Store) *http.Server {
	server := &Server{
		config: &cfg,
		store:  store,
	}
	if cfg.NVIDIAAPIKey != "" {
		server.ai = newNVIDIAClient(cfg.NVIDIAAPIKey, cfg.NVIDIAModel)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/auth/login", server.handleLogin)
	mux.HandleFunc("/api/auth/me", server.withAuth(server.handleMe))
	mux.HandleFunc("/api/dashboard", server.withAuth(server.handleDashboard))
	mux.HandleFunc("/api/goals", server.withAuth(server.handleGoals))
	mux.HandleFunc("/api/goals/", server.withAuth(server.handleGoalByID))
	mux.HandleFunc("/api/entries", server.withAuth(server.handleEntries))
	mux.HandleFunc("/api/roadmaps", server.withAuth(server.handleRoadmaps))
	mux.HandleFunc("/api/roadmaps/generate", server.withAuth(server.handleGenerateRoadmapGraph))
	mux.HandleFunc("/api/roadmap-tasks/", server.withAuth(server.handleRoadmapTaskByID))
	mux.HandleFunc("/api/ai/roadmap", server.withAuth(server.handleRoadmap))
	mux.HandleFunc("/api/ai/hints", server.withAuth(server.handleHints))

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.config.FrontendOrigin
		if origin == "*" {
			if reqOrigin := r.Header.Get("Origin"); reqOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type contextKey string

const contextUserKey = contextKey("userEmail")

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Missing bearer token."})
			return
		}

		payload, err := verifyToken(strings.TrimPrefix(header, "Bearer "), s.config.JWTSecret)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token."})
			return
		}

		ctx := context.WithValue(r.Context(), contextUserKey, payload.Sub)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid login payload."})
		return
	}

	if !validEmail(req.Email) || len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid login payload."})
		return
	}

	if !secureStringEqual(strings.ToLower(req.Email), strings.ToLower(s.config.AdminEmail)) ||
		!secureStringEqual(req.Password, s.config.AdminPassword) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid credentials."})
		return
	}

	token, err := issueToken(s.config.AdminEmail, s.config.JWTSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not issue token."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  map[string]string{"email": s.config.AdminEmail},
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]string{"email": r.Context().Value(contextUserKey).(string)},
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	dashboard, err := s.store.Dashboard()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load dashboard."})
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (s *Server) handleGoals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		period := r.URL.Query().Get("period")
		if period != "" && period != "weekly" && period != "monthly" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid period."})
			return
		}

		goals, err := s.store.ListGoals(period)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load goals."})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"goals": goals})
	case http.MethodPost:
		var input GoalInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid goal payload."})
			return
		}
		if err := validateGoal(input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid goal payload.", "details": err.Error()})
			return
		}

		goal, err := s.store.CreateGoal(normalizeGoal(input))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not create goal."})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"goal": goal})
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleGoalByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/api/goals/")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Goal not found."})
		return
	}

	switch r.Method {
	case http.MethodPut:
		var input GoalInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid goal payload."})
			return
		}
		if err := validateGoal(input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid goal payload.", "details": err.Error()})
			return
		}

		goal, err := s.store.UpdateGoal(id, normalizeGoal(input))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not update goal."})
			return
		}
		if goal == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Goal not found."})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"goal": goal})
	case http.MethodDelete:
		if err := s.store.DeleteGoal(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not delete goal."})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleEntries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit := 30
		if raw := r.URL.Query().Get("limit"); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid limit."})
				return
			}
			if value < 1 {
				value = 1
			}
			if value > 120 {
				value = 120
			}
			limit = value
		}

		entries, err := s.store.ListEntries(limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load entries."})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"entries": entries})
	case http.MethodPost:
		var input EntryInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid entry payload."})
			return
		}
		if err := validateEntry(input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid entry payload.", "details": err.Error()})
			return
		}

		entry, err := s.store.UpsertEntry(input)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save entry."})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"entry": entry})
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleRoadmaps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	roadmaps, err := s.store.ListRoadmaps()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load roadmaps."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"roadmaps": roadmaps})
}

func (s *Server) handleGenerateRoadmapGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req RoadmapGenerateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid roadmap request."})
		return
	}
	if err := validateRoadmapGenerate(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid roadmap request.", "details": err.Error()})
		return
	}

	goal, err := s.store.GetGoal(req.GoalID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load goal."})
		return
	}
	if goal == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Goal not found."})
		return
	}

	input := fallbackRoadmapInput(*goal, req.Period, req.StartDate)
	if s.ai != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
		defer cancel()
		content, err := s.ai.generate(
			ctx,
			"You produce strict JSON roadmap graphs for a private progress tracker.",
			structuredRoadmapPrompt(*goal, req.Period, req.StartDate),
		)
		if err == nil {
			input = parseRoadmapInput(*goal, req.Period, req.StartDate, content)
		}
	}

	roadmap, err := s.store.SaveRoadmap(input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save roadmap."})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"roadmap": roadmap})
}

func (s *Server) handleRoadmapTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/api/roadmap-tasks/")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Roadmap task not found."})
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}

	var req RoadmapTaskUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid task payload."})
		return
	}

	roadmap, err := s.store.SetRoadmapTaskDone(id, req.Done)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not update roadmap task."})
		return
	}
	if roadmap == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Roadmap task not found."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"roadmap": roadmap})
}

func (s *Server) handleRoadmap(w http.ResponseWriter, r *http.Request) {
	s.handleAI(w, r, "You create realistic execution roadmaps for one private user.", func(d Dashboard, prompt string) string {
		return roadmapPrompt(d.Goals, d.Entries, prompt)
	})
}

func (s *Server) handleHints(w http.ResponseWriter, r *http.Request) {
	s.handleAI(w, r, "You create short, useful guidance for one private user.", func(d Dashboard, prompt string) string {
		return hintsPrompt(d.Goals, d.Entries, prompt)
	})
}

func (s *Server) handleAI(w http.ResponseWriter, r *http.Request, system string, buildPrompt func(Dashboard, string) string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if s.ai == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "NVIDIA_API_KEY is not configured."})
		return
	}

	var req AIRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid AI request payload."})
		return
	}
	if len(req.Prompt) > 2000 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid AI request payload."})
		return
	}

	dashboard, err := s.store.Dashboard()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load dashboard."})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	content, err := s.ai.generate(ctx, system, buildPrompt(dashboard, req.Prompt))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"content": content})
}

func decodeJSON(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed."})
}

func parseIDFromPath(path, prefix string) (int64, error) {
	raw := strings.TrimPrefix(path, prefix)
	if raw == "" || strings.Contains(raw, "/") {
		return 0, fmt.Errorf("invalid id")
	}
	return strconv.ParseInt(raw, 10, 64)
}

func validEmail(email string) bool {
	return strings.Contains(email, "@") && len(email) <= 320
}

func validateGoal(input GoalInput) error {
	if input.Period != "weekly" && input.Period != "monthly" {
		return errors.New("period must be weekly or monthly")
	}
	if strings.TrimSpace(input.Title) == "" || len(input.Title) > 160 {
		return errors.New("title is required and must be at most 160 characters")
	}
	if len(input.Description) > 2000 {
		return errors.New("description must be at most 2000 characters")
	}
	if input.TargetValue < 1 || input.TargetValue > 100000 {
		return errors.New("target_value must be between 1 and 100000")
	}
	if input.CurrentValue < 0 || input.CurrentValue > 100000 {
		return errors.New("current_value must be between 0 and 100000")
	}
	if input.DueDate != nil && len(*input.DueDate) > 20 {
		return errors.New("due_date must be at most 20 characters")
	}
	if input.Status != "" && input.Status != "active" && input.Status != "completed" && input.Status != "paused" {
		return errors.New("status must be active, completed, or paused")
	}
	return nil
}

func normalizeGoal(input GoalInput) GoalInput {
	if input.Status == "" {
		input.Status = "active"
	}
	return input
}

func validateEntry(input EntryInput) error {
	if len(input.EntryDate) != 10 {
		return errors.New("entry_date must be in YYYY-MM-DD format")
	}
	if len(input.Summary) > 2000 || len(input.Wins) > 2000 || len(input.Blockers) > 2000 {
		return errors.New("summary, wins, and blockers must be at most 2000 characters")
	}
	if len(input.Notes) > 4000 {
		return errors.New("notes must be at most 4000 characters")
	}
	if len(input.Mood) > 80 {
		return errors.New("mood must be at most 80 characters")
	}
	return nil
}

func validateRoadmapGenerate(input RoadmapGenerateRequest) error {
	if input.GoalID <= 0 {
		return errors.New("goal_id is required")
	}
	if input.Period != "weekly" && input.Period != "monthly" {
		return errors.New("period must be weekly or monthly")
	}
	if len(input.StartDate) != 10 {
		return errors.New("start_date must be in YYYY-MM-DD format")
	}
	return nil
}
