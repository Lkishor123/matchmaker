package config

import (
	"fmt"
	"os"
	"strings"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func require(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("missing %s", key)
	}
	return v, nil
}

// Auth holds configuration for the auth service.
type Auth struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTPrivateKey      string
	UserServiceURL     string
}

// LoadAuth reads environment variables and validates required fields.
func LoadAuth() (*Auth, error) {
	var missing []string
	cfg := &Auth{
		GoogleClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GoogleRedirectURL:  getenv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8081/api/v1/auth/google/callback"),
		JWTPrivateKey:      os.Getenv("JWT_PRIVATE_KEY"),
		UserServiceURL:     getenv("USER_SERVICE_URL", "http://localhost:8084"),
	}
	if cfg.GoogleClientID == "" {
		missing = append(missing, "GOOGLE_OAUTH_CLIENT_ID")
	}
	if cfg.GoogleClientSecret == "" {
		missing = append(missing, "GOOGLE_OAUTH_CLIENT_SECRET")
	}
	if cfg.JWTPrivateKey == "" {
		missing = append(missing, "JWT_PRIVATE_KEY")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing env vars: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// User holds configuration for the user service.
type User struct {
	PostgresURL string
}

// LoadUser loads the configuration for the user service.
func LoadUser() (*User, error) {
	dsn, err := require("POSTGRES_URL")
	if err != nil {
		return nil, err
	}
	return &User{PostgresURL: dsn}, nil
}

// Report holds configuration for the astrology report service.
type Report struct {
	MongoURL              string
	RedisURL              string
	AstrologyEngineURL    string
	AstrologyEngineAPIKey string
}

// LoadReport reads config for the report service.
func LoadReport() (*Report, error) {
	var missing []string
	cfg := &Report{
		MongoURL:              os.Getenv("MONGO_URL"),
		RedisURL:              os.Getenv("REDIS_URL"),
		AstrologyEngineURL:    os.Getenv("ASTROLOGY_ENGINE_URL"),
		AstrologyEngineAPIKey: os.Getenv("ASTROLOGY_ENGINE_API_KEY"),
	}
	if cfg.MongoURL == "" {
		missing = append(missing, "MONGO_URL")
	}
	if cfg.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if cfg.AstrologyEngineURL == "" {
		missing = append(missing, "ASTROLOGY_ENGINE_URL")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing env vars: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// Match holds configuration for the match analysis service.
type Match struct {
	ReportServiceURL string
}

// LoadMatch returns config for the match service.
func LoadMatch() (*Match, error) {
	return &Match{
		ReportServiceURL: getenv("REPORT_SERVICE_URL", "http://localhost:8082"),
	}, nil
}

// Chat holds configuration for the chat service.
type Chat struct {
	RedisURL  string
	LLMAPIURL string
	LLMAPIKey string
}

// LoadChat loads config for the chat service.
func LoadChat() (*Chat, error) {
	if os.Getenv("REDIS_URL") == "" {
		return nil, fmt.Errorf("missing REDIS_URL")
	}
	cfg := &Chat{
		RedisURL:  os.Getenv("REDIS_URL"),
		LLMAPIURL: getenv("LLM_API_URL", "https://example.com/api/chat"),
		LLMAPIKey: os.Getenv("LLM_API_KEY"),
	}
	if cfg.LLMAPIKey == "" {
		return nil, fmt.Errorf("missing LLM_API_KEY")
	}
	return cfg, nil
}
