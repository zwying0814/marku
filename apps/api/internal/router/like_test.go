package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type likeStateDTO struct {
	Likes uint64 `json:"likes"`
	Liked bool   `json:"liked"`
}

func TestLikeSetAndGet(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, key := createSite(t, app, token, "Site A", "site-a", nil)

	status, state := sendLike(t, app, key, http.MethodPut, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 1 || !state.Liked {
		t.Fatalf("first like = status %d state %+v", status, state)
	}
	status, state = sendLike(t, app, key, http.MethodPut, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 1 || !state.Liked {
		t.Fatalf("idempotent like = status %d state %+v", status, state)
	}

	status, state = sendLike(t, app, key, http.MethodGet, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 1 || !state.Liked {
		t.Fatalf("visitor a state = status %d state %+v", status, state)
	}
	status, state = sendLike(t, app, key, http.MethodGet, "/posts/hello", "visitor-b")
	if status != http.StatusOK || state.Likes != 1 || state.Liked {
		t.Fatalf("visitor b state = status %d state %+v", status, state)
	}

	status, state = sendLike(t, app, key, http.MethodDelete, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 0 || state.Liked {
		t.Fatalf("unlike = status %d state %+v", status, state)
	}
	status, state = sendLike(t, app, key, http.MethodDelete, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 0 || state.Liked {
		t.Fatalf("idempotent unlike = status %d state %+v", status, state)
	}

	status, state = sendLike(t, app, key, http.MethodPost, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 1 || !state.Liked {
		t.Fatalf("toggle on = status %d state %+v", status, state)
	}
	status, state = sendLike(t, app, key, http.MethodPost, "/posts/hello", "visitor-a")
	if status != http.StatusOK || state.Likes != 0 || state.Liked {
		t.Fatalf("toggle off = status %d state %+v", status, state)
	}
}

func TestLikeSiteIsolation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, keyA := createSite(t, app, token, "Site A", "site-a", nil)
	_, keyB := createSite(t, app, token, "Site B", "site-b", nil)

	if status, state := sendLike(t, app, keyA, http.MethodPut, "/same", "visitor-a"); status != http.StatusOK || state.Likes != 1 {
		t.Fatalf("site A like = status %d state %+v", status, state)
	}
	if status, state := sendLike(t, app, keyA, http.MethodGet, "/same", "visitor-a"); status != http.StatusOK || state.Likes != 1 {
		t.Fatalf("site A state = status %d state %+v", status, state)
	}
	if status, state := sendLike(t, app, keyB, http.MethodGet, "/same", "visitor-a"); status != http.StatusOK || state.Likes != 0 || state.Liked {
		t.Fatalf("site B saw site A likes: status %d state %+v", status, state)
	}
}

func TestLikeAdminStats(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)

	sendLike(t, app, key, http.MethodPut, "/a", "visitor-a")
	sendLike(t, app, key, http.MethodPut, "/a", "visitor-b")
	sendLike(t, app, key, http.MethodPut, "/b", "visitor-c")

	stats := getAdminLikeStats(t, app, token, siteID, "?range=7d&limit=5")
	if stats.TotalLikes != 3 || len(stats.TopPages) != 2 || stats.TopPages[0].Path != "/a" || stats.TopPages[0].Likes != 2 {
		t.Fatalf("unexpected site like stats: %+v", stats)
	}
	if len(stats.Trend) != 7 || stats.Trend[6].Likes != 3 || stats.Trend[6].Unlikes != 0 || stats.Trend[6].Net != 3 {
		t.Fatalf("unexpected like trend: %+v", stats.Trend)
	}

	pathStats := getAdminLikeStats(t, app, token, siteID, "?range=7d&path=%2Fa")
	if pathStats.TotalLikes != 2 || pathStats.Path != "/a" || len(pathStats.TopPages) != 0 {
		t.Fatalf("unexpected path like stats: %+v", pathStats)
	}
	if pathStats.Trend[6].Net != 2 {
		t.Fatalf("unexpected path trend: %+v", pathStats.Trend)
	}

	sendLike(t, app, key, http.MethodDelete, "/a", "visitor-a")
	stats = getAdminLikeStats(t, app, token, siteID, "?range=7d&limit=5")
	if stats.TotalLikes != 2 || stats.Trend[6].Likes != 3 || stats.Trend[6].Unlikes != 1 || stats.Trend[6].Net != 2 {
		t.Fatalf("unexpected stats after unlike: %+v", stats)
	}
}

func TestLikeValidation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	_, key := createSite(t, app, token, "Site A", "site-a", nil)

	missingVisitor := perform(app.engine, http.MethodPut, "/api/v1/public/likes",
		strings.NewReader(`{"path":"/missing-visitor"}`),
		map[string]string{"Content-Type": "application/json", "X-Site-Key": key},
	)
	if missingVisitor.Code != http.StatusBadRequest {
		t.Fatalf("missing visitor status = %d, want 400", missingVisitor.Code)
	}

	missingPath := perform(app.engine, http.MethodPut, "/api/v1/public/likes",
		strings.NewReader(`{"visitorId":"visitor-a"}`),
		map[string]string{"Content-Type": "application/json", "X-Site-Key": key},
	)
	if missingPath.Code != http.StatusBadRequest {
		t.Fatalf("missing path status = %d, want 400", missingPath.Code)
	}

	getMissingPath := perform(app.engine, http.MethodGet, "/api/v1/public/likes", nil, map[string]string{"X-Site-Key": key})
	if getMissingPath.Code != http.StatusBadRequest {
		t.Fatalf("get missing path status = %d, want 400", getMissingPath.Code)
	}
}

func sendLike(t *testing.T, app *adminTestApp, key, method, path, visitorID string) (int, likeStateDTO) {
	t.Helper()

	var recorderBody string
	target := "/api/v1/public/likes"
	if method == http.MethodGet {
		query := url.Values{}
		query.Set("path", path)
		if visitorID != "" {
			query.Set("visitorId", visitorID)
		}
		target += "?" + query.Encode()
	} else {
		payload, _ := json.Marshal(map[string]string{
			"path":      path,
			"visitorId": visitorID,
		})
		recorderBody = string(payload)
	}

	var body io.Reader
	if recorderBody != "" {
		body = strings.NewReader(recorderBody)
	}

	headers := map[string]string{"X-Site-Key": key}
	if recorderBody != "" {
		headers["Content-Type"] = "application/json"
	}
	recorder := perform(app.engine, method, target, body, headers)

	var envelope struct {
		Data likeStateDTO `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode like response: %v", err)
		}
	}
	return recorder.Code, envelope.Data
}

type adminLikeStatsDTO struct {
	Path       string `json:"path"`
	TotalLikes uint64 `json:"totalLikes"`
	TopPages   []struct {
		Path  string `json:"path"`
		Likes uint64 `json:"likes"`
	} `json:"topPages"`
	Trend []struct {
		Date    string `json:"date"`
		Likes   uint64 `json:"likes"`
		Unlikes uint64 `json:"unlikes"`
		Net     int64  `json:"net"`
	} `json:"trend"`
}

func getAdminLikeStats(t *testing.T, app *adminTestApp, token string, siteID uint, query string) adminLikeStatsDTO {
	t.Helper()
	recorder := perform(app.engine, http.MethodGet,
		"/api/v1/admin/sites/"+itoa(siteID)+"/likes"+query,
		nil,
		authHeaders(token),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("admin like stats status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Data adminLikeStatsDTO `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode admin like stats: %v", err)
	}
	return body.Data
}
