package router

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"marku/api/internal/config"
)

type recordViewDTO struct {
	Views       uint64 `json:"views"`
	UniqueViews uint64 `json:"uniqueViews"`
	Recorded    bool   `json:"recorded"`
}

type pageViewDTO struct {
	Path        string `json:"path"`
	Views       uint64 `json:"views"`
	UniqueViews uint64 `json:"uniqueViews"`
}

func TestPageViewRecordAndBatch(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, key := createSite(t, app, token, "Site A", "site-a", nil)

	if status, result := recordView(t, app, key, "/posts/hello?utm=1", "", "Mozilla/5.0"); status != http.StatusOK || !result.Recorded || result.Views != 1 {
		t.Fatalf("first view = status %d result %+v", status, result)
	}
	if status, result := recordView(t, app, key, "/posts/hello", "", "Mozilla/5.0"); status != http.StatusOK || !result.Recorded || result.Views != 2 {
		t.Fatalf("second view = status %d result %+v", status, result)
	}
	if status, result := recordView(t, app, key, "/about", "", "Mozilla/5.0"); status != http.StatusOK || result.Views != 1 {
		t.Fatalf("about view = status %d result %+v", status, result)
	}

	status, items := getViews(t, app, key, []string{"/posts/hello", "/about", "/missing"})
	if status != http.StatusOK || len(items) != 3 {
		t.Fatalf("batch views = status %d items %+v", status, items)
	}
	if items[0].Views != 2 || items[1].Views != 1 || items[2].Views != 0 {
		t.Fatalf("unexpected batch views: %+v", items)
	}

	missingPath := perform(app.engine, http.MethodPost, "/api/v1/public/views",
		strings.NewReader(`{}`),
		map[string]string{"Content-Type": "application/json", "X-Site-Key": key},
	)
	if missingPath.Code != http.StatusBadRequest {
		t.Fatalf("missing path status = %d, want 400", missingPath.Code)
	}

	emptyBatch := perform(app.engine, http.MethodGet, "/api/v1/public/views", nil, map[string]string{"X-Site-Key": key})
	if emptyBatch.Code != http.StatusBadRequest {
		t.Fatalf("empty batch status = %d, want 400", emptyBatch.Code)
	}
}

func TestPageViewSiteIsolation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, keyA := createSite(t, app, token, "Site A", "site-a", nil)
	_, keyB := createSite(t, app, token, "Site B", "site-b", nil)

	if status, _ := recordView(t, app, keyA, "/same", "", "Mozilla/5.0"); status != http.StatusOK {
		t.Fatalf("site A view status = %d", status)
	}

	_, itemsA := getViews(t, app, keyA, []string{"/same"})
	_, itemsB := getViews(t, app, keyB, []string{"/same"})
	if len(itemsA) != 1 || itemsA[0].Views != 1 {
		t.Fatalf("site A views = %+v", itemsA)
	}
	if len(itemsB) != 1 || itemsB[0].Views != 0 {
		t.Fatalf("site B saw site A views: %+v", itemsB)
	}
}

func TestPageViewVisitorDedupe(t *testing.T) {
	app := newAdminTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.View.DedupeByVisitorDay = true
	})
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, key := createSite(t, app, token, "Site A", "site-a", nil)

	first := recordViewWithVisitor(t, app, key, "/dedupe", "visitor-a")
	second := recordViewWithVisitor(t, app, key, "/dedupe", "visitor-a")
	third := recordViewWithVisitor(t, app, key, "/dedupe", "visitor-b")

	if first.Views != 1 || first.UniqueViews != 1 {
		t.Fatalf("first dedupe view = %+v", first)
	}
	if second.Views != 2 || second.UniqueViews != 1 {
		t.Fatalf("same visitor view = %+v", second)
	}
	if third.Views != 3 || third.UniqueViews != 2 {
		t.Fatalf("second visitor view = %+v", third)
	}
}

func TestPageViewBotFilter(t *testing.T) {
	app := newAdminTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.View.FilterBots = true
	})
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, key := createSite(t, app, token, "Site A", "site-a", nil)

	botStatus, bot := recordView(t, app, key, "/bot", "", "Googlebot/2.1")
	if botStatus != http.StatusOK || bot.Recorded || bot.Views != 0 {
		t.Fatalf("bot view = status %d result %+v", botStatus, bot)
	}

	humanStatus, human := recordView(t, app, key, "/bot", "", "Mozilla/5.0")
	if humanStatus != http.StatusOK || !human.Recorded || human.Views != 1 {
		t.Fatalf("human view = status %d result %+v", humanStatus, human)
	}
}

func TestAdminViewStats(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)

	recordView(t, app, key, "/a", "", "Mozilla/5.0")
	recordView(t, app, key, "/a", "", "Mozilla/5.0")
	recordView(t, app, key, "/b", "", "Mozilla/5.0")

	recorder := perform(app.engine, http.MethodGet,
		"/api/v1/admin/sites/"+itoa(siteID)+"/stats?range=7d&limit=5",
		nil,
		authHeaders(token),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("admin stats status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Data struct {
			TotalViews       uint64        `json:"totalViews"`
			TotalUniqueViews uint64        `json:"totalUniqueViews"`
			TopPages         []pageViewDTO `json:"topPages"`
			Trend            []struct {
				Date  string `json:"date"`
				Views uint64 `json:"views"`
			} `json:"trend"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if body.Data.TotalViews != 3 || body.Data.TotalUniqueViews != 0 {
		t.Fatalf("unexpected totals: %+v", body.Data)
	}
	if len(body.Data.TopPages) != 2 || body.Data.TopPages[0].Path != "/a" || body.Data.TopPages[0].Views != 2 {
		t.Fatalf("unexpected top pages: %+v", body.Data.TopPages)
	}
	if len(body.Data.Trend) != 7 || body.Data.Trend[6].Views != 3 {
		t.Fatalf("unexpected trend: %+v", body.Data.Trend)
	}

	invalidRange := perform(app.engine, http.MethodGet,
		"/api/v1/admin/sites/"+itoa(siteID)+"/stats?range=abc",
		nil,
		authHeaders(token),
	)
	if invalidRange.Code != http.StatusBadRequest {
		t.Fatalf("invalid range status = %d, want 400", invalidRange.Code)
	}
}

func recordView(t *testing.T, app *adminTestApp, key, path, visitorID, userAgent string) (int, recordViewDTO) {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{
		"path":      path,
		"visitorId": visitorID,
	})
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Site-Key":   key,
	}
	if userAgent != "" {
		headers["User-Agent"] = userAgent
	}
	recorder := perform(app.engine, http.MethodPost, "/api/v1/public/views",
		strings.NewReader(string(payload)),
		headers,
	)
	var body struct {
		Data recordViewDTO `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode view response: %v", err)
		}
	}
	return recorder.Code, body.Data
}

func recordViewWithVisitor(t *testing.T, app *adminTestApp, key, path, visitorID string) recordViewDTO {
	t.Helper()
	status, result := recordView(t, app, key, path, visitorID, "Mozilla/5.0")
	if status != http.StatusOK {
		t.Fatalf("record view status = %d", status)
	}
	return result
}

func getViews(t *testing.T, app *adminTestApp, key string, paths []string) (int, []pageViewDTO) {
	t.Helper()
	query := url.Values{}
	query.Set("paths", strings.Join(paths, ","))
	recorder := perform(app.engine, http.MethodGet, "/api/v1/public/views?"+query.Encode(), nil, map[string]string{
		"X-Site-Key": key,
	})
	var body struct {
		Data struct {
			Items []pageViewDTO `json:"items"`
		} `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode views: %v", err)
		}
	}
	return recorder.Code, body.Data.Items
}
