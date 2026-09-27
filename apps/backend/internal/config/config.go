package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	AppEnv  string
	AppPort string

	DatabaseURL string

	// MasterKeyBase64 is the base64-encoded 32-byte AES-256 master key (version 1).
	MasterKeyBase64 string
	JWTAccessSecret string

	CookieDomain string
	CookieSecure bool

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	EmailFrom    string

	CORSOrigins []string
}

// Load reads configuration from the environment and validates required values.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "dev"),
		AppPort:         getEnv("APP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		MasterKeyBase64: os.Getenv("APP_MASTER_KEY_V1_BASE64"),
		JWTAccessSecret: getEnv("JWT_ACCESS_SECRET", "dev-secret"),
		CookieDomain:    getEnv("COOKIE_DOMAIN", "localhost"),
		CookieSecure:    getEnvBool("COOKIE_SECURE", false),
		SMTPHost:        os.Getenv("SMTP_HOST"),
		SMTPPort:        getEnvInt("SMTP_PORT", 587),
		SMTPUser:        os.Getenv("SMTP_USER"),
		SMTPPassword:    os.Getenv("SMTP_PASSWORD"),
		EmailFrom:       getEnv("EMAIL_FROM", "no-reply@example.com"),
		CORSOrigins:     splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173")),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.MasterKeyBase64 == "" {
		return nil, fmt.Errorf("APP_MASTER_KEY_V1_BASE64 is required")
	}

	return cfg, nil
}

// IsProd reports whether the app is running in a production-like environment.
func (c *Config) IsProd() bool {
	return c.AppEnv == "prod" || c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
