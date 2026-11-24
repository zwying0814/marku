package router

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"marku/api/internal/config"
	"marku/api/internal/model"
	"marku/api/internal/repository"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

func TestHealthAndRequestID(t *testing.T) {
	engine, _ := newTestRouter(t, nil)

	recorder := perform(engine, http.MethodGet, "/healthz", nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}

	var envelope struct {
		Code response.Code  `json:"code"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if envelope.Code != response.CodeOK || envelope.Data["status"] != "ok" {
		t.Fatalf("unexpected health response: %+v", envelope)
	}

	preserved := perform(engine, http.MethodGet, "/healthz", nil, map[string]string{
		"X-Request-ID": "test-request-123",
	})
	if got := preserved.Header().Get("X-Request-ID"); got != "test-request-123" {
		t.Fatalf("request id = %q, want test-request-123", got)
	}

	replaced := perform(engine, http.MethodGet, "/healthz", nil, map[string]string{
		"X-Request-ID": "bad",
	})
	if got := replaced.Header().Get("X-Request-ID"); got == "bad" || len(got) != 32 {
		t.Fatalf("invalid request id was not replaced: %q", got)
	}
}

func TestReady(t *testing.T) {
	engine, _ := newTestRouter(t, nil)
	recorder := perform(engine, http.MethodGet, "/readyz", nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want 200", recorder.Code)
	}
}

func TestNotFoundAndMethodNotAllowed(t *testing.T) {
	engine, _ := newTestRouter(t, nil)

	notFound := perform(engine, http.MethodGet, "/missing", nil, nil)
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d, want 404", notFound.Code)
	}
	assertErrorCode(t, notFound, response.CodeNotFound)

	notAllowed := perform(engine, http.MethodPost, "/healthz", nil, nil)
	if notAllowed.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method not allowed status = %d, want 405", notAllowed.Code)
	}
	assertErrorCode(t, notAllowed, response.CodeMethodNotAllowed)
}

func TestCORS(t *testing.T) {
	engine, _ := newTestRouter(t, func(cfg *config.Config) {
		cfg.CORS.AllowedOrigins = []string{"https://blog.example.com"}
	})

	preflight := perform(engine, http.MethodOptions, "/api/v1", nil, map[string]string{
		"Origin":                        "https://blog.example.com",
		"Access-Control-Request-Method": "GET",
	})
	if preflight.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", preflight.Code)
	}
	if got := preflight.Header().Get("Access-Control-Allow-Origin"); got != "https://blog.example.com" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := preflight.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "OPTIONS") {
		t.Fatalf("allow methods = %q, want OPTIONS", got)
	}

	denied := perform(engine, http.MethodOptions, "/api/v1", nil, map[string]string{
		"Origin": "https://evil.example.com",
	})
	if denied.Code != http.StatusForbidden {
		t.Fatalf("denied preflight status = %d, want 403", denied.Code)
	}

	normal := perform(engine, http.MethodGet, "/api/v1", nil, map[string]string{
		"Origin": "https://blog.example.com",
	})
	if normal.Code != http.StatusOK {
		t.Fatalf("normal request status = %d, want 200", normal.Code)
	}
	if got := normal.Header().Get("Access-Control-Allow-Origin"); got != "https://blog.example.com" {
		t.Fatalf("normal allow origin = %q", got)
	}
}

func TestRateLimit(t *testing.T) {
	engine, _ := newTestRouter(t, func(cfg *config.Config) {
		cfg.RateLimit = config.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 1,
			Burst:             1,
		}
	})

	first := perform(engine, http.MethodGet, "/api/v1", nil, nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", first.Code)
	}

	second := perform(engine, http.MethodGet, "/api/v1", nil, nil)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", second.Code)
	}
	if got := second.Header().Get("Retry-After"); got != "1" {
		t.Fatalf("retry-after = %q, want 1", got)
	}
	assertErrorCode(t, second, response.CodeRateLimited)
}

func TestBodyLimit(t *testing.T) {
	engine, _ := newTestRouter(t, func(cfg *config.Config) {
		cfg.HTTP.MaxBodyBytes = 8
	})

	recorder := perform(engine, http.MethodPost, "/api/v1", strings.NewReader(strings.Repeat("x", 16)), nil)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("body limit status = %d, want 413", recorder.Code)
	}
	assertErrorCode(t, recorder, response.CodePayloadTooLarge)
}

func TestRecovery(t *testing.T) {
	engine, _ := newTestRouter(t, nil)
	engine.GET("/panic", func(_ *gin.Context) {
		panic("boom")
	})

	recorder := perform(engine, http.MethodGet, "/panic", nil, nil)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("panic status = %d, want 500", recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("panic response should include X-Request-ID")
	}
	assertErrorCode(t, recorder, response.CodeInternal)
}

func newTestRouter(t *testing.T, mutate func(*config.Config)) (*gin.Engine, *gorm.DB) {
	engine, db, _ := newTestApp(t, mutate)
	return engine, db
}

func newTestApp(t *testing.T, mutate func(*config.Config)) (*gin.Engine, *gorm.DB, *service.Service) {
	t.Helper()

	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(model.MigrationModels()...); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	cfg := config.Config{
		Env: config.EnvTest,
		HTTP: config.HTTPConfig{
			Addr:              ":0",
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
			IdleTimeout:       time.Second,
			ShutdownTimeout:   time.Second,
			MaxBodyBytes:      1 << 20,
		},
		DB: config.DBConfig{
			Driver:          "sqlite",
			DSN:             ":memory:",
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: time.Minute,
		},
		Auth: config.AuthConfig{
			JWTSecret: strings.Repeat("j", 32),
			TokenTTL:  time.Hour,
		},
		CORS: config.CORSConfig{MaxAge: time.Hour},
		RateLimit: config.RateLimitConfig{
			Enabled:           false,
			RequestsPerSecond: 10,
			Burst:             20,
		},
		Comment: config.CommentConfig{
			RateLimitWindow:   10 * time.Minute,
			RateLimitMax:      100,
			MinSubmitDuration: time.Second,
			MaxLength:         2000,
			MaxLinks:          5,
			MaxDepth:          3,
			DuplicateWindow:   10 * time.Minute,
			IPHashSalt:        "test-ip-salt",
		},
		View: config.ViewConfig{
			DedupeByVisitorDay: false,
			FilterBots:         false,
			MaxBatchPaths:      50,
		},
		Log: config.LogConfig{Level: slog.LevelInfo, Format: "text"},
	}
	if mutate != nil {
		mutate(&cfg)
	}

	store := repository.New(db)
	services := service.New(store, cfg)
	engine, err := New(cfg, db, services, WithoutWebUI())
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return engine, db, services
}

func perform(engine *gin.Engine, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, body)
	request.RemoteAddr = "192.0.2.1:1234"
	request.Header.Set("User-Agent", "marku-test-agent")
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func assertErrorCode(t *testing.T, recorder *httptest.ResponseRecorder, want response.Code) {
	t.Helper()
	var envelope response.Envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if envelope.Code != want {
		t.Fatalf("response code = %d, want %d (body=%s)", envelope.Code, want, recorder.Body.String())
	}
}
