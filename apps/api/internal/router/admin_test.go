package router

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"marku/api/internal/config"
	"marku/api/internal/model"
	"marku/api/internal/response"
	"marku/api/internal/service"
)

type adminTestApp struct {
	engine   *gin.Engine
	db       *gorm.DB
	services *service.Service
}

const testAdminEmail = "admin@example.com"

var (
	testAdminPassword = strings.Repeat("p", 16)
	testNewPassword   = strings.Repeat("n", 16)
	testWrongPassword = strings.Repeat("w", 16)
)

func TestAdminAuthentication(t *testing.T) {
	app := newAdminTestApp(t)

	unauthorized := perform(app.engine, http.MethodGet, "/api/v1/admin/auth/me", nil, nil)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", unauthorized.Code)
	}
	if unauthorized.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatal("missing WWW-Authenticate header")
	}
	assertErrorCode(t, unauthorized, response.CodeUnauthorized)

	badLogin := perform(app.engine, http.MethodPost, "/api/v1/admin/auth/login",
		strings.NewReader(loginJSON(testAdminEmail, testWrongPassword)),
		map[string]string{"Content-Type": "application/json"},
	)
	if badLogin.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, want 401", badLogin.Code)
	}

	token := login(t, app, testAdminEmail, testAdminPassword)
	me := perform(app.engine, http.MethodGet, "/api/v1/admin/auth/me", nil, authHeaders(token))
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want 200", me.Code)
	}
	var meBody struct {
		Data struct {
			Email  string `json:"email"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(me.Body.Bytes(), &meBody); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if meBody.Data.Email != testAdminEmail || meBody.Data.Status != model.AdminStatusActive {
		t.Fatalf("unexpected me response: %+v", meBody.Data)
	}

	if err := app.db.Model(&model.Admin{}).
		Where("email = ?", testAdminEmail).
		Update("status", model.AdminStatusDisabled).Error; err != nil {
		t.Fatalf("disable admin: %v", err)
	}

	disabled := perform(app.engine, http.MethodGet, "/api/v1/admin/auth/me", nil, authHeaders(token))
	if disabled.Code != http.StatusForbidden {
		t.Fatalf("disabled admin status = %d, want 403", disabled.Code)
	}
	assertErrorCode(t, disabled, response.CodeForbidden)
}

func TestAdminChangePassword(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)

	headers := authHeaders(token)
	headers["Content-Type"] = "application/json"
	changed := perform(app.engine, http.MethodPost, "/api/v1/admin/auth/change-password",
		strings.NewReader(changePasswordJSON(testAdminPassword, testNewPassword)),
		headers,
	)
	if changed.Code != http.StatusNoContent {
		t.Fatalf("change password status = %d, want 204 (body=%s)", changed.Code, changed.Body.String())
	}

	oldLogin := perform(app.engine, http.MethodPost, "/api/v1/admin/auth/login",
		strings.NewReader(loginJSON(testAdminEmail, testAdminPassword)),
		map[string]string{"Content-Type": "application/json"},
	)
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want 401", oldLogin.Code)
	}

	login(t, app, testAdminEmail, testNewPassword)
}

func TestSiteLifecycleAndIsolation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)

	siteAID, siteAKey := createSite(t, app, token, "Site A", "site-a", []string{"https://a.example.com"})
	siteBID, siteBKey := createSite(t, app, token, "Site B", "site-b", []string{"https://b.example.com"})
	if siteAID == siteBID {
		t.Fatal("expected distinct site IDs")
	}

	list := perform(app.engine, http.MethodGet, "/api/v1/admin/sites", nil, authHeaders(token))
	if list.Code != http.StatusOK {
		t.Fatalf("list sites status = %d, want 200", list.Code)
	}
	var listBody struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode site list: %v", err)
	}
	if listBody.Data.Total != 2 {
		t.Fatalf("site total = %d, want 2", listBody.Data.Total)
	}

	if status, id := publicSite(t, app.engine, siteAKey); status != http.StatusOK || id != siteAID {
		t.Fatalf("site A public lookup = status %d id %d", status, id)
	}
	if status, id := publicSite(t, app.engine, siteBKey); status != http.StatusOK || id != siteBID {
		t.Fatalf("site B public lookup = status %d id %d", status, id)
	}

	missingKey := perform(app.engine, http.MethodGet, "/api/v1/public/site", nil, nil)
	if missingKey.Code != http.StatusUnauthorized {
		t.Fatalf("missing site key status = %d, want 401", missingKey.Code)
	}
	invalidKey := perform(app.engine, http.MethodGet, "/api/v1/public/site", nil, map[string]string{
		"X-Site-Key": "mk_invalid_key",
	})
	if invalidKey.Code != http.StatusUnauthorized {
		t.Fatalf("invalid site key status = %d, want 401", invalidKey.Code)
	}

	rotated := perform(app.engine, http.MethodPost, "/api/v1/admin/sites/"+itoa(siteAID)+"/rotate-key", nil, authHeaders(token))
	if rotated.Code != http.StatusOK {
		t.Fatalf("rotate key status = %d, want 200", rotated.Code)
	}
	var rotatedBody struct {
		Data struct {
			SiteKey string `json:"siteKey"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rotated.Body.Bytes(), &rotatedBody); err != nil {
		t.Fatalf("decode rotate response: %v", err)
	}
	if rotatedBody.Data.SiteKey == "" || rotatedBody.Data.SiteKey == siteAKey {
		t.Fatal("rotated key should be new")
	}
	if status, _ := publicSite(t, app.engine, siteAKey); status != http.StatusUnauthorized {
		t.Fatalf("old site key status = %d, want 401", status)
	}
	if status, id := publicSite(t, app.engine, rotatedBody.Data.SiteKey); status != http.StatusOK || id != siteAID {
		t.Fatalf("new site key lookup = status %d id %d", status, id)
	}

	updateHeaders := authHeaders(token)
	updateHeaders["Content-Type"] = "application/json"
	disabled := perform(app.engine, http.MethodPatch, "/api/v1/admin/sites/"+itoa(siteAID),
		strings.NewReader(`{"status":"disabled"}`), updateHeaders)
	if disabled.Code != http.StatusOK {
		t.Fatalf("disable site status = %d, want 200", disabled.Code)
	}
	if status, _ := publicSite(t, app.engine, rotatedBody.Data.SiteKey); status != http.StatusForbidden {
		t.Fatalf("disabled site public status = %d, want 403", status)
	}

	enabled := perform(app.engine, http.MethodPatch, "/api/v1/admin/sites/"+itoa(siteAID),
		strings.NewReader(`{"status":"active"}`), updateHeaders)
	if enabled.Code != http.StatusOK {
		t.Fatalf("enable site status = %d, want 200", enabled.Code)
	}

	revoked := perform(app.engine, http.MethodDelete, "/api/v1/admin/sites/"+itoa(siteBID)+"/key", nil, authHeaders(token))
	if revoked.Code != http.StatusNoContent {
		t.Fatalf("revoke key status = %d, want 204", revoked.Code)
	}
	if status, _ := publicSite(t, app.engine, siteBKey); status != http.StatusUnauthorized {
		t.Fatalf("revoked site key status = %d, want 401", status)
	}

	deleted := perform(app.engine, http.MethodDelete, "/api/v1/admin/sites/"+itoa(siteAID), nil, authHeaders(token))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete site status = %d, want 204", deleted.Code)
	}
	if status, _ := publicSite(t, app.engine, rotatedBody.Data.SiteKey); status != http.StatusUnauthorized {
		t.Fatalf("deleted site public status = %d, want 401", status)
	}
	missing := perform(app.engine, http.MethodGet, "/api/v1/admin/sites/"+itoa(siteAID), nil, authHeaders(token))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("deleted site admin status = %d, want 404", missing.Code)
	}
}

func TestSiteValidation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, _ := createSite(t, app, token, "Site A", "site-a", nil)

	duplicate := perform(app.engine, http.MethodPost, "/api/v1/admin/sites",
		strings.NewReader(`{"name":"Duplicate","slug":"site-a"}`),
		withJSON(authHeaders(token)),
	)
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate slug status = %d, want 409", duplicate.Code)
	}

	invalidSlug := perform(app.engine, http.MethodPost, "/api/v1/admin/sites",
		strings.NewReader(`{"name":"Bad Slug","slug":"Bad Slug"}`),
		withJSON(authHeaders(token)),
	)
	if invalidSlug.Code != http.StatusBadRequest {
		t.Fatalf("invalid slug status = %d, want 400", invalidSlug.Code)
	}

	invalidOrigin := perform(app.engine, http.MethodPost, "/api/v1/admin/sites",
		strings.NewReader(`{"name":"Bad Origin","slug":"bad-origin","allowedOrigins":["*"]}`),
		withJSON(authHeaders(token)),
	)
	if invalidOrigin.Code != http.StatusBadRequest {
		t.Fatalf("invalid origin status = %d, want 400", invalidOrigin.Code)
	}

	emptyUpdate := perform(app.engine, http.MethodPatch, "/api/v1/admin/sites/"+itoa(siteID),
		strings.NewReader(`{}`), withJSON(authHeaders(token)))
	if emptyUpdate.Code != http.StatusBadRequest {
		t.Fatalf("empty update status = %d, want 400", emptyUpdate.Code)
	}
}

func TestPublicCORSAllowsAnyOrigin(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, key := createSite(t, app, token, "Site A", "site-a", []string{"https://a.example.com"})

	preflight := perform(app.engine, http.MethodOptions, "/api/v1/public/site", nil, map[string]string{
		"Origin":                        "https://blog.example.com",
		"Access-Control-Request-Method": "GET",
	})
	if preflight.Code != http.StatusNoContent {
		t.Fatalf("public preflight status = %d, want 204", preflight.Code)
	}
	if got := preflight.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("public preflight allow origin = %q, want *", got)
	}
	if got := preflight.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Site-Key") {
		t.Fatalf("public preflight allow headers = %q, missing X-Site-Key", got)
	}

	request := perform(app.engine, http.MethodGet, "/api/v1/public/site", nil, map[string]string{
		"Origin":     "https://blog.example.com",
		"X-Site-Key": key,
	})
	if request.Code != http.StatusOK {
		t.Fatalf("public request status = %d, want 200", request.Code)
	}
	if got := request.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("public request allow origin = %q, want *", got)
	}
}

func newAdminTestApp(t *testing.T) *adminTestApp {
	return newAdminTestAppWithConfig(t, nil)
}

func newAdminTestAppWithConfig(t *testing.T, mutate func(*config.Config)) *adminTestApp {
	t.Helper()

	engine, db, services := newTestApp(t, func(cfg *config.Config) {
		cfg.Auth = config.AuthConfig{
			JWTSecret:     strings.Repeat("j", 32),
			TokenTTL:      time.Hour,
			AdminEmail:    testAdminEmail,
			AdminPassword: testAdminPassword,
		}
		if mutate != nil {
			mutate(cfg)
		}
	})
	if _, err := services.SeedAdmin(); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	return &adminTestApp{
		engine:   engine,
		db:       db,
		services: services,
	}
}

func login(t *testing.T, app *adminTestApp, email, password string) string {
	t.Helper()

	payload, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("encode login: %v", err)
	}
	recorder := perform(app.engine, http.MethodPost, "/api/v1/admin/auth/login",
		strings.NewReader(string(payload)),
		map[string]string{"Content-Type": "application/json"},
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200 (body=%s)", recorder.Code, recorder.Body.String())
	}

	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if body.Data.Token == "" {
		t.Fatal("login returned empty token")
	}
	return body.Data.Token
}

func createSite(t *testing.T, app *adminTestApp, token, name, slug string, origins []string) (uint, string) {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"name":           name,
		"slug":           slug,
		"allowedOrigins": origins,
	})
	if err != nil {
		t.Fatalf("encode site: %v", err)
	}

	recorder := perform(app.engine, http.MethodPost, "/api/v1/admin/sites",
		strings.NewReader(string(payload)),
		withJSON(authHeaders(token)),
	)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create site status = %d, want 201 (body=%s)", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "siteKeyHash") {
		t.Fatal("site response must not expose siteKeyHash")
	}

	var body struct {
		Data struct {
			ID      uint   `json:"id"`
			SiteKey string `json:"siteKey"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode site response: %v", err)
	}
	if body.Data.ID == 0 || body.Data.SiteKey == "" {
		t.Fatalf("invalid create site response: %+v", body.Data)
	}
	return body.Data.ID, body.Data.SiteKey
}

func publicSite(t *testing.T, engine *gin.Engine, key string) (int, uint) {
	t.Helper()

	recorder := perform(engine, http.MethodGet, "/api/v1/public/site", nil, map[string]string{
		"X-Site-Key": key,
	})
	if recorder.Code != http.StatusOK {
		return recorder.Code, 0
	}

	var body struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode public site response: %v", err)
	}
	return recorder.Code, body.Data.ID
}

func authHeaders(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func withJSON(headers map[string]string) map[string]string {
	headers["Content-Type"] = "application/json"
	return headers
}

func loginJSON(email, password string) string {
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	return string(payload)
}

func changePasswordJSON(currentPassword, newPassword string) string {
	payload, _ := json.Marshal(map[string]string{
		"currentPassword": currentPassword,
		"newPassword":     newPassword,
	})
	return string(payload)
}

func itoa(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}
