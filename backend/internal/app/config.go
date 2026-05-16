package app

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port           string
	JWTSecret      string
	AdminEmail     string
	AdminPassword  string
	FrontendOrigin string
	NVIDIAAPIKey   string
	NVIDIAModel    string
	DatabasePath   string
}

func LoadConfig(envPath string) (Config, error) {
	_ = loadDotEnv(envPath)

	cfg := Config{
		Port:           envOrDefault("PORT", "4000"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AdminEmail:     os.Getenv("ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
		FrontendOrigin: envOrDefault("FRONTEND_ORIGIN", "*"),
		NVIDIAAPIKey:   os.Getenv("NVIDIA_API_KEY"),
		NVIDIAModel:    envOrDefault("NVIDIA_MODEL", "qwen/qwen3-next-80b-a3b-instruct"),
		DatabasePath:   filepath.Join("data", "progress.db"),
	}

	if cfg.JWTSecret == "" || cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return Config{}, errors.New("missing required auth configuration in backend/.env")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key != "" {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
