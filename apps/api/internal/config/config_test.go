package config

import (
	"log/slog"
	"strings"
	"testing"
)

func TestLoadFromDefaults(t *testing.T) {
	cfg, err := loadFrom(values(nil))
	if err != nil {
		t.Fatalf("load defaults: %v", err)
	}

	if cfg.Env != EnvDevelopment {
		t.Fatalf("Env = %q, want %q", cfg.Env, EnvDevelopment)
	}
	if cfg.HTTP.Addr != ":36580" {
		t.Fatalf("HTTP.Addr = %q, want :36580", cfg.HTTP.Addr)
	}
	if cfg.DB.Driver != "sqlite" || cfg.DB.DSN != "./data/marku.db" {
		t.Fatalf("unexpected database defaults: %+v", cfg.DB)
	}
	if !cfg.DB.AutoMigrate {
		t.Fatal("AutoMigrate should default to true in development")
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("default CORS origins = %#v, want two localhost origins", cfg.CORS.AllowedOrigins)
	}
	if cfg.Log.Format != "text" || cfg.Log.Level != slog.LevelInfo {
		t.Fatalf("unexpected log defaults: %+v", cfg.Log)
	}
	if !cfg.RateLimit.Enabled {
		t.Fatal("rate limiting should be enabled by default")
	}
	if cfg.Comment.MaxLength != 2000 || cfg.Comment.MaxDepth != 3 || cfg.Comment.RateLimitMax != 5 {
		t.Fatalf("unexpected comment defaults: %+v", cfg.Comment)
	}
	if cfg.View.MaxBatchPaths != 50 || !cfg.View.FilterBots || cfg.View.DedupeByVisitorDay {
		t.Fatalf("unexpected view defaults: %+v", cfg.View)
	}
}

func TestLoadFromProduction(t *testing.T) {
	cfg, err := loadFrom(values(map[string]string{
		"MARKU_ENV":          EnvProduction,
		"MARKU_JWT_SECRET":   strings.Repeat("s", 32),
		"MARKU_IP_HASH_SALT": strings.Repeat("i", 16),
	}))
	if err != nil {
		t.Fatalf("load production config: %v", err)
	}

	if cfg.DB.AutoMigrate {
		t.Fatal("AutoMigrate should default to false in production")
	}
	if cfg.Log.Format != "json" {
		t.Fatalf("Log.Format = %q, want json", cfg.Log.Format)
	}
	if len(cfg.CORS.AllowedOrigins) != 0 {
		t.Fatalf("production CORS origins should be empty by default, got %#v", cfg.CORS.AllowedOrigins)
	}
}

func TestLoadFromOverrides(t *testing.T) {
	cfg, err := loadFrom(values(map[string]string{
		"MARKU_ENV":                        EnvTest,
		"MARKU_HTTP_ADDR":                  "127.0.0.1:18080",
		"MARKU_HTTP_MAX_BODY_BYTES":        "4096",
		"MARKU_HTTP_TRUSTED_PROXIES":       "127.0.0.1, 10.0.0.0/8,127.0.0.1",
		"MARKU_DB_MAX_OPEN_CONNS":          "4",
		"MARKU_DB_MAX_IDLE_CONNS":          "2",
		"MARKU_JWT_SECRET":                 strings.Repeat("j", 32),
		"MARKU_CORS_ALLOWED_ORIGINS":       "https://blog.example.com, https://blog.example.com/, http://localhost:36583",
		"MARKU_CORS_ALLOW_CREDENTIALS":     "true",
		"MARKU_RATE_LIMIT_RPS":             "2.5",
		"MARKU_RATE_LIMIT_BURST":           "7",
		"MARKU_COMMENT_RATE_LIMIT_MAX":     "9",
		"MARKU_COMMENT_MAX_LENGTH":         "1500",
		"MARKU_COMMENT_MAX_LINKS":          "2",
		"MARKU_COMMENT_MAX_DEPTH":          "4",
		"MARKU_IP_HASH_SALT":               "test-ip-salt",
		"MARKU_VIEW_DEDUPE_BY_VISITOR_DAY": "true",
		"MARKU_VIEW_FILTER_BOTS":           "false",
		"MARKU_VIEW_MAX_BATCH_PATHS":       "20",
		"MARKU_LOG_LEVEL":                  "debug",
		"MARKU_LOG_FORMAT":                 "json",
	}))
	if err != nil {
		t.Fatalf("load overrides: %v", err)
	}

	if cfg.HTTP.Addr != "127.0.0.1:18080" || cfg.HTTP.MaxBodyBytes != 4096 {
		t.Fatalf("unexpected HTTP config: %+v", cfg.HTTP)
	}
	if len(cfg.HTTP.TrustedProxies) != 2 {
		t.Fatalf("TrustedProxies = %#v, want two entries", cfg.HTTP.TrustedProxies)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins = %#v, want two normalized entries", cfg.CORS.AllowedOrigins)
	}
	if cfg.CORS.AllowedOrigins[0] != "https://blog.example.com" {
		t.Fatalf("first origin = %q", cfg.CORS.AllowedOrigins[0])
	}
	if !cfg.CORS.AllowCredentials {
		t.Fatal("AllowCredentials should be true")
	}
	if cfg.RateLimit.RequestsPerSecond != 2.5 || cfg.RateLimit.Burst != 7 {
		t.Fatalf("unexpected rate limit config: %+v", cfg.RateLimit)
	}
	if cfg.Comment.RateLimitMax != 9 || cfg.Comment.MaxLength != 1500 || cfg.Comment.MaxLinks != 2 || cfg.Comment.MaxDepth != 4 {
		t.Fatalf("unexpected comment config: %+v", cfg.Comment)
	}
	if !cfg.View.DedupeByVisitorDay || cfg.View.FilterBots || cfg.View.MaxBatchPaths != 20 {
		t.Fatalf("unexpected view config: %+v", cfg.View)
	}
	if cfg.Log.Level != slog.LevelDebug || cfg.Log.Format != "json" {
		t.Fatalf("unexpected log config: %+v", cfg.Log)
	}
}

func TestLoadFromValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "invalid env", env: map[string]string{"MARKU_ENV": "staging"}, want: "MARKU_ENV"},
		{name: "invalid driver", env: map[string]string{"MARKU_DB_DRIVER": "oracle"}, want: "MARKU_DB_DRIVER"},
		{name: "invalid timeout", env: map[string]string{"MARKU_HTTP_READ_HEADER_TIMEOUT": "nope"}, want: "MARKU_HTTP_READ_HEADER_TIMEOUT"},
		{name: "invalid bool", env: map[string]string{"MARKU_RATE_LIMIT_ENABLED": "sometimes"}, want: "MARKU_RATE_LIMIT_ENABLED"},
		{name: "invalid log level", env: map[string]string{"MARKU_LOG_LEVEL": "verbose"}, want: "MARKU_LOG_LEVEL"},
		{name: "invalid log format", env: map[string]string{"MARKU_LOG_FORMAT": "xml"}, want: "MARKU_LOG_FORMAT"},
		{name: "wildcard credentials", env: map[string]string{"MARKU_CORS_ALLOWED_ORIGINS": "*", "MARKU_CORS_ALLOW_CREDENTIALS": "true"}, want: "credentials"},
		{name: "origin path", env: map[string]string{"MARKU_CORS_ALLOWED_ORIGINS": "https://example.com/path"}, want: "must not contain a path"},
		{name: "wildcard proxy", env: map[string]string{"MARKU_HTTP_TRUSTED_PROXIES": "*"}, want: "explicit proxies"},
		{name: "short jwt", env: map[string]string{"MARKU_JWT_SECRET": "short"}, want: "at least 16"},
		{name: "idle exceeds open", env: map[string]string{"MARKU_DB_MAX_IDLE_CONNS": "2"}, want: "MARKU_DB_MAX_IDLE_CONNS"},
		{name: "zero body limit", env: map[string]string{"MARKU_HTTP_MAX_BODY_BYTES": "0"}, want: "MARKU_HTTP_MAX_BODY_BYTES"},
		{name: "zero rate limiter", env: map[string]string{"MARKU_RATE_LIMIT_RPS": "0"}, want: "MARKU_RATE_LIMIT_RPS"},
		{name: "invalid comment depth", env: map[string]string{"MARKU_COMMENT_MAX_DEPTH": "0"}, want: "MARKU_COMMENT_MAX_DEPTH"},
		{name: "invalid comment length", env: map[string]string{"MARKU_COMMENT_MAX_LENGTH": "0"}, want: "MARKU_COMMENT_MAX_LENGTH"},
		{name: "invalid view batch", env: map[string]string{"MARKU_VIEW_MAX_BATCH_PATHS": "0"}, want: "MARKU_VIEW_MAX_BATCH_PATHS"},
		{name: "production without jwt", env: map[string]string{"MARKU_ENV": EnvProduction}, want: "MARKU_JWT_SECRET"},
		{name: "production short jwt", env: map[string]string{"MARKU_ENV": EnvProduction, "MARKU_JWT_SECRET": "short"}, want: "at least 32"},
		{name: "production without ip salt", env: map[string]string{"MARKU_ENV": EnvProduction, "MARKU_JWT_SECRET": strings.Repeat("s", 32)}, want: "MARKU_IP_HASH_SALT"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := loadFrom(values(test.env)); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func values(input map[string]string) lookupFunc {
	return func(key string) (string, bool) {
		value, ok := input[key]
		return value, ok
	}
}
