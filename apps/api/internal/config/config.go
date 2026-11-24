package config

import (
	"fmt"
	"log/slog"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

type Config struct {
	Env       string
	HTTP      HTTPConfig
	DB        DBConfig
	Auth      AuthConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Comment   CommentConfig
	View      ViewConfig
	Log       LogConfig
}

type HTTPConfig struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxBodyBytes      int64
	TrustedProxies    []string
}

type DBConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	AutoMigrate     bool
}

type AuthConfig struct {
	JWTSecret     string
	TokenTTL      time.Duration
	AdminEmail    string
	AdminPassword string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond float64
	Burst             int
}

type CommentConfig struct {
	RateLimitWindow   time.Duration
	RateLimitMax      int
	MinSubmitDuration time.Duration
	MaxLength         int
	MaxLinks          int
	MaxDepth          int
	DuplicateWindow   time.Duration
	IPHashSalt        string
}

type ViewConfig struct {
	DedupeByVisitorDay bool
	FilterBots         bool
	MaxBatchPaths      int
}

type LogConfig struct {
	Level  slog.Level
	Format string
}

type lookupFunc func(string) (string, bool)

func Load() (Config, error) {
	_ = godotenv.Load()
	return loadFrom(os.LookupEnv)
}

func (c Config) IsProduction() bool {
	return c.Env == EnvProduction
}

func loadFrom(lookup lookupFunc) (Config, error) {
	env := strings.ToLower(get(lookup, "MARKU_ENV", EnvDevelopment))
	switch env {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		return Config{}, fmt.Errorf("MARKU_ENV must be one of development, test, production")
	}

	isProduction := env == EnvProduction
	defaultOrigins := "http://localhost:36581,http://127.0.0.1:36581"
	defaultLogFormat := "text"
	defaultAutoMigrate := "true"
	defaultJWTSecret := ""
	if isProduction {
		defaultOrigins = ""
		defaultLogFormat = "json"
		defaultAutoMigrate = "false"
	}

	driver := strings.ToLower(get(lookup, "MARKU_DB_DRIVER", "sqlite"))
	defaultMaxOpen := 25
	defaultMaxIdle := 10
	if driver == "sqlite" {
		defaultMaxOpen = 1
		defaultMaxIdle = 1
	}

	readHeaderTimeout, err := durationEnv(lookup, "MARKU_HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := durationEnv(lookup, "MARKU_HTTP_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := durationEnv(lookup, "MARKU_HTTP_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := durationEnv(lookup, "MARKU_HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationEnv(lookup, "MARKU_HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	connMaxLifetime, err := durationEnv(lookup, "MARKU_DB_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	tokenTTL, err := durationEnv(lookup, "MARKU_JWT_TTL", 8*time.Hour)
	if err != nil {
		return Config{}, err
	}
	corsMaxAge, err := durationEnv(lookup, "MARKU_CORS_MAX_AGE", 12*time.Hour)
	if err != nil {
		return Config{}, err
	}
	commentRateLimitWindow, err := durationEnv(lookup, "MARKU_COMMENT_RATE_LIMIT_WINDOW", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}
	commentMinSubmitDuration, err := durationEnv(lookup, "MARKU_COMMENT_MIN_SUBMIT_DURATION", 2*time.Second)
	if err != nil {
		return Config{}, err
	}
	commentDuplicateWindow, err := durationEnv(lookup, "MARKU_COMMENT_DUPLICATE_WINDOW", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}

	maxBodyBytes, err := int64Env(lookup, "MARKU_HTTP_MAX_BODY_BYTES", 1<<20)
	if err != nil {
		return Config{}, err
	}
	maxOpenConns, err := intEnv(lookup, "MARKU_DB_MAX_OPEN_CONNS", defaultMaxOpen)
	if err != nil {
		return Config{}, err
	}
	maxIdleConns, err := intEnv(lookup, "MARKU_DB_MAX_IDLE_CONNS", defaultMaxIdle)
	if err != nil {
		return Config{}, err
	}
	rateLimitBurst, err := intEnv(lookup, "MARKU_RATE_LIMIT_BURST", 40)
	if err != nil {
		return Config{}, err
	}
	commentRateLimitMax, err := intEnv(lookup, "MARKU_COMMENT_RATE_LIMIT_MAX", 5)
	if err != nil {
		return Config{}, err
	}
	commentMaxLength, err := intEnv(lookup, "MARKU_COMMENT_MAX_LENGTH", 2000)
	if err != nil {
		return Config{}, err
	}
	commentMaxLinks, err := intEnv(lookup, "MARKU_COMMENT_MAX_LINKS", 3)
	if err != nil {
		return Config{}, err
	}
	commentMaxDepth, err := intEnv(lookup, "MARKU_COMMENT_MAX_DEPTH", 3)
	if err != nil {
		return Config{}, err
	}
	viewMaxBatchPaths, err := intEnv(lookup, "MARKU_VIEW_MAX_BATCH_PATHS", 50)
	if err != nil {
		return Config{}, err
	}
	rateLimitRPS, err := floatEnv(lookup, "MARKU_RATE_LIMIT_RPS", 20)
	if err != nil {
		return Config{}, err
	}
	autoMigrate, err := boolEnv(lookup, "MARKU_DB_AUTO_MIGRATE", defaultAutoMigrate == "true")
	if err != nil {
		return Config{}, err
	}
	rateLimitEnabled, err := boolEnv(lookup, "MARKU_RATE_LIMIT_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	allowCredentials, err := boolEnv(lookup, "MARKU_CORS_ALLOW_CREDENTIALS", false)
	if err != nil {
		return Config{}, err
	}
	viewDedupe, err := boolEnv(lookup, "MARKU_VIEW_DEDUPE_BY_VISITOR_DAY", false)
	if err != nil {
		return Config{}, err
	}
	viewFilterBots, err := boolEnv(lookup, "MARKU_VIEW_FILTER_BOTS", true)
	if err != nil {
		return Config{}, err
	}

	logLevel := slog.LevelInfo
	rawLogLevel := strings.ToLower(get(lookup, "MARKU_LOG_LEVEL", "info"))
	if err := logLevel.UnmarshalText([]byte(rawLogLevel)); err != nil {
		return Config{}, fmt.Errorf("MARKU_LOG_LEVEL must be one of debug, info, warn, error")
	}

	cfg := Config{
		Env: env,
		HTTP: HTTPConfig{
			Addr:              get(lookup, "MARKU_HTTP_ADDR", ":36580"),
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			ShutdownTimeout:   shutdownTimeout,
			MaxBodyBytes:      maxBodyBytes,
			TrustedProxies:    splitList(get(lookup, "MARKU_HTTP_TRUSTED_PROXIES", "")),
		},
		DB: DBConfig{
			Driver:          driver,
			DSN:             get(lookup, "MARKU_DB_DSN", "./data/marku.db"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			AutoMigrate:     autoMigrate,
		},
		Auth: AuthConfig{
			JWTSecret:     get(lookup, "MARKU_JWT_SECRET", defaultJWTSecret),
			TokenTTL:      tokenTTL,
			AdminEmail:    get(lookup, "MARKU_ADMIN_EMAIL", ""),
			AdminPassword: get(lookup, "MARKU_ADMIN_PASSWORD", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins:   splitList(get(lookup, "MARKU_CORS_ALLOWED_ORIGINS", defaultOrigins)),
			AllowCredentials: allowCredentials,
			MaxAge:           corsMaxAge,
		},
		RateLimit: RateLimitConfig{
			Enabled:           rateLimitEnabled,
			RequestsPerSecond: rateLimitRPS,
			Burst:             rateLimitBurst,
		},
		Comment: CommentConfig{
			RateLimitWindow:   commentRateLimitWindow,
			RateLimitMax:      commentRateLimitMax,
			MinSubmitDuration: commentMinSubmitDuration,
			MaxLength:         commentMaxLength,
			MaxLinks:          commentMaxLinks,
			MaxDepth:          commentMaxDepth,
			DuplicateWindow:   commentDuplicateWindow,
			IPHashSalt:        get(lookup, "MARKU_IP_HASH_SALT", ""),
		},
		View: ViewConfig{
			DedupeByVisitorDay: viewDedupe,
			FilterBots:         viewFilterBots,
			MaxBatchPaths:      viewMaxBatchPaths,
		},
		Log: LogConfig{
			Level:  logLevel,
			Format: strings.ToLower(get(lookup, "MARKU_LOG_FORMAT", defaultLogFormat)),
		},
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.HTTP.Addr) == "" {
		return fmt.Errorf("MARKU_HTTP_ADDR must not be empty")
	}
	if c.HTTP.ReadHeaderTimeout <= 0 || c.HTTP.ReadTimeout <= 0 || c.HTTP.WriteTimeout <= 0 || c.HTTP.IdleTimeout <= 0 {
		return fmt.Errorf("HTTP timeouts must be greater than zero")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("MARKU_HTTP_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.HTTP.MaxBodyBytes <= 0 {
		return fmt.Errorf("MARKU_HTTP_MAX_BODY_BYTES must be greater than zero")
	}
	for _, proxy := range c.HTTP.TrustedProxies {
		if proxy == "*" {
			return fmt.Errorf("MARKU_HTTP_TRUSTED_PROXIES must list explicit proxies; \"*\" is not allowed")
		}
	}

	switch c.DB.Driver {
	case "sqlite", "mysql", "postgres":
	default:
		return fmt.Errorf("MARKU_DB_DRIVER must be one of sqlite, mysql, postgres")
	}
	if strings.TrimSpace(c.DB.DSN) == "" {
		return fmt.Errorf("MARKU_DB_DSN must not be empty")
	}
	if c.DB.MaxOpenConns < 1 {
		return fmt.Errorf("MARKU_DB_MAX_OPEN_CONNS must be at least 1")
	}
	if c.DB.MaxIdleConns < 0 || c.DB.MaxIdleConns > c.DB.MaxOpenConns {
		return fmt.Errorf("MARKU_DB_MAX_IDLE_CONNS must be between 0 and MARKU_DB_MAX_OPEN_CONNS")
	}
	if c.DB.ConnMaxLifetime <= 0 {
		return fmt.Errorf("MARKU_DB_CONN_MAX_LIFETIME must be greater than zero")
	}

	if c.Auth.JWTSecret == "" {
		if c.IsProduction() {
			return fmt.Errorf("MARKU_JWT_SECRET is required in production")
		}
	} else if c.IsProduction() && len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("MARKU_JWT_SECRET must be at least 32 characters in production")
	} else if len(c.Auth.JWTSecret) < 16 {
		return fmt.Errorf("MARKU_JWT_SECRET must be at least 16 characters")
	}
	if c.Auth.TokenTTL <= 0 {
		return fmt.Errorf("MARKU_JWT_TTL must be greater than zero")
	}
	if (c.Auth.AdminEmail == "") != (c.Auth.AdminPassword == "") {
		return fmt.Errorf("MARKU_ADMIN_EMAIL and MARKU_ADMIN_PASSWORD must be set together")
	}
	if c.Auth.AdminEmail != "" {
		address, err := mail.ParseAddress(c.Auth.AdminEmail)
		if err != nil || address.Address != c.Auth.AdminEmail {
			return fmt.Errorf("MARKU_ADMIN_EMAIL must be a valid email address")
		}
		if len(c.Auth.AdminPassword) < 8 {
			return fmt.Errorf("MARKU_ADMIN_PASSWORD must be at least 8 characters")
		}
		if c.Auth.JWTSecret == "" {
			return fmt.Errorf("MARKU_JWT_SECRET is required when admin credentials are configured")
		}
	}
	if c.IsProduction() && c.Auth.AdminEmail == "" {
		return fmt.Errorf("MARKU_ADMIN_EMAIL and MARKU_ADMIN_PASSWORD are required in production")
	}

	if c.CORS.MaxAge <= 0 {
		return fmt.Errorf("MARKU_CORS_MAX_AGE must be greater than zero")
	}
	if c.CORS.AllowCredentials {
		for _, origin := range c.CORS.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("MARKU_CORS_ALLOWED_ORIGINS cannot contain \"*\" when credentials are enabled")
			}
		}
	}
	normalizedOrigins, err := normalizeOrigins(c.CORS.AllowedOrigins)
	if err != nil {
		return err
	}
	c.CORS.AllowedOrigins = normalizedOrigins

	if c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerSecond <= 0 {
			return fmt.Errorf("MARKU_RATE_LIMIT_RPS must be greater than zero")
		}
		if c.RateLimit.Burst < 1 {
			return fmt.Errorf("MARKU_RATE_LIMIT_BURST must be at least 1")
		}
	}

	if c.Comment.RateLimitWindow <= 0 {
		return fmt.Errorf("MARKU_COMMENT_RATE_LIMIT_WINDOW must be greater than zero")
	}
	if c.Comment.RateLimitMax < 1 {
		return fmt.Errorf("MARKU_COMMENT_RATE_LIMIT_MAX must be at least 1")
	}
	if c.Comment.MinSubmitDuration <= 0 {
		return fmt.Errorf("MARKU_COMMENT_MIN_SUBMIT_DURATION must be greater than zero")
	}
	if c.Comment.MaxLength < 1 || c.Comment.MaxLength > 10000 {
		return fmt.Errorf("MARKU_COMMENT_MAX_LENGTH must be between 1 and 10000")
	}
	if c.Comment.MaxLinks < 0 {
		return fmt.Errorf("MARKU_COMMENT_MAX_LINKS must not be negative")
	}
	if c.Comment.MaxDepth < 1 || c.Comment.MaxDepth > 10 {
		return fmt.Errorf("MARKU_COMMENT_MAX_DEPTH must be between 1 and 10")
	}
	if c.Comment.DuplicateWindow <= 0 {
		return fmt.Errorf("MARKU_COMMENT_DUPLICATE_WINDOW must be greater than zero")
	}
	if c.IsProduction() && len(c.Comment.IPHashSalt) < 16 {
		return fmt.Errorf("MARKU_IP_HASH_SALT must be at least 16 characters in production")
	}

	if c.View.MaxBatchPaths < 1 || c.View.MaxBatchPaths > 100 {
		return fmt.Errorf("MARKU_VIEW_MAX_BATCH_PATHS must be between 1 and 100")
	}

	switch c.Log.Format {
	case "text", "json":
	default:
		return fmt.Errorf("MARKU_LOG_FORMAT must be one of text, json")
	}

	return nil
}

func normalizeOrigins(origins []string) ([]string, error) {
	normalized := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			if _, ok := seen[origin]; !ok {
				seen[origin] = struct{}{}
				normalized = append(normalized, origin)
			}
			continue
		}

		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, fmt.Errorf("MARKU_CORS_ALLOWED_ORIGINS contains invalid origin %q", origin)
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return nil, fmt.Errorf("MARKU_CORS_ALLOWED_ORIGINS origin %q must not contain a path", origin)
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("MARKU_CORS_ALLOWED_ORIGINS origin %q must not contain query or fragment", origin)
		}

		value := parsed.Scheme + "://" + parsed.Host
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized, nil
}

func get(lookup lookupFunc, key, fallback string) string {
	if value, ok := lookup(key); ok {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return fallback
}

func splitList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func durationEnv(lookup lookupFunc, key string, fallback time.Duration) (time.Duration, error) {
	raw := get(lookup, key, fallback.String())
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func intEnv(lookup lookupFunc, key string, fallback int) (int, error) {
	raw := get(lookup, key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return value, nil
}

func int64Env(lookup lookupFunc, key string, fallback int64) (int64, error) {
	raw := get(lookup, key, strconv.FormatInt(fallback, 10))
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return value, nil
}

func floatEnv(lookup lookupFunc, key string, fallback float64) (float64, error) {
	raw := get(lookup, key, strconv.FormatFloat(fallback, 'f', -1, 64))
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}
	return value, nil
}

func boolEnv(lookup lookupFunc, key string, fallback bool) (bool, error) {
	raw := get(lookup, key, strconv.FormatBool(fallback))
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return value, nil
}
