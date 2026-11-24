package router

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"marku/api/internal/config"
)

type submittedComment struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type publicCommentDTO struct {
	ID            uint               `json:"id"`
	ParentID      *uint              `json:"parentId"`
	AuthorName    string             `json:"authorName"`
	AuthorEmail   string             `json:"authorEmail"`
	AuthorWebsite string             `json:"authorWebsite"`
	Content       string             `json:"content"`
	IsAdmin       bool               `json:"isAdmin"`
	Replies       []publicCommentDTO `json:"replies"`
}

type adminCommentDTO struct {
	ID          uint   `json:"id"`
	SiteID      uint   `json:"siteId"`
	Path        string `json:"path"`
	ParentID    *uint  `json:"parentId"`
	AuthorName  string `json:"authorName"`
	AuthorEmail string `json:"authorEmail"`
	Content     string `json:"content"`
	Status      string `json:"status"`
	IsAdmin     bool   `json:"isAdmin"`
	IPHash      string `json:"ipHash"`
	UserAgent   string `json:"userAgent"`
}

func TestPublicCommentModerationAndAdminReply(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)

	status, submitted, body := submitComment(t, app, key, commentPayload("/posts/hello", "Hello <b>world</b>"))
	if status != http.StatusCreated || submitted.ID == 0 || submitted.Status != "pending" {
		t.Fatalf("submit comment = status %d body %s", status, body)
	}

	if listStatus, items, total := listPublicComments(t, app, key, "/posts/hello"); listStatus != http.StatusOK || total != 0 || len(items) != 0 {
		t.Fatalf("pending comment leaked: status=%d total=%d items=%d", listStatus, total, len(items))
	}
	if countStatus, count := publicCommentCount(t, app, key, "/posts/hello"); countStatus != http.StatusOK || count != 0 {
		t.Fatalf("pending comment count = status %d count %d", countStatus, count)
	}

	adminStatus, adminItems, adminTotal := listAdminComments(t, app, token, siteID, "")
	if adminStatus != http.StatusOK || adminTotal != 1 || len(adminItems) != 1 {
		t.Fatalf("admin list = status %d total %d items %d", adminStatus, adminTotal, len(adminItems))
	}
	if adminItems[0].AuthorEmail != "alice@example.com" || adminItems[0].IPHash == "" || adminItems[0].UserAgent == "" {
		t.Fatalf("admin comment missing private fields: %+v", adminItems[0])
	}

	approved := perform(app.engine, http.MethodPatch, "/api/v1/admin/comments/"+itoa(submitted.ID),
		strings.NewReader(`{"status":"approved"}`), withJSON(authHeaders(token)))
	if approved.Code != http.StatusOK {
		t.Fatalf("approve comment status = %d body=%s", approved.Code, approved.Body.String())
	}

	listStatus, items, total := listPublicComments(t, app, key, "/posts/hello")
	if listStatus != http.StatusOK || total != 1 || len(items) != 1 {
		t.Fatalf("public approved list = status %d total %d items %d", listStatus, total, len(items))
	}
	if items[0].Content != "Hello <b>world</b>" || items[0].AuthorEmail != "" {
		t.Fatalf("unexpected public comment: %+v", items[0])
	}
	if countStatus, count := publicCommentCount(t, app, key, "/posts/hello"); countStatus != http.StatusOK || count != 1 {
		t.Fatalf("approved count = status %d count %d", countStatus, count)
	}
	publicRaw := perform(app.engine, http.MethodGet, "/api/v1/public/comments?path=/posts/hello", nil, map[string]string{
		"X-Site-Key": key,
	})
	for _, privateField := range []string{"authorEmail", "ipHash", "userAgent"} {
		if strings.Contains(publicRaw.Body.String(), privateField) {
			t.Fatalf("public response leaked %s: %s", privateField, publicRaw.Body.String())
		}
	}

	replyStatus, reply := replyComment(t, app, token, submitted.ID, "Thanks for reading")
	if replyStatus != http.StatusCreated || reply.ID == 0 || !reply.IsAdmin {
		t.Fatalf("admin reply = status %d dto %+v", replyStatus, reply)
	}

	listStatus, items, total = listPublicComments(t, app, key, "/posts/hello")
	if listStatus != http.StatusOK || total != 1 || len(items) != 1 || len(items[0].Replies) != 1 {
		t.Fatalf("threaded list = status %d total %d items %d replies %d", listStatus, total, len(items), len(items[0].Replies))
	}
	if !items[0].Replies[0].IsAdmin || items[0].Replies[0].AuthorName != "Admin" {
		t.Fatalf("unexpected admin reply dto: %+v", items[0].Replies[0])
	}
	if countStatus, count := publicCommentCount(t, app, key, "/posts/hello"); countStatus != http.StatusOK || count != 2 {
		t.Fatalf("thread count = status %d count %d", countStatus, count)
	}
}

func TestCommentSiteIsolation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteAID, keyA := createSite(t, app, token, "Site A", "site-a", nil)
	siteBID, keyB := createSite(t, app, token, "Site B", "site-b", nil)
	enableAutoApprove(t, app, token, siteAID)

	if status, submitted, body := submitComment(t, app, keyA, commentPayload("/same-path", "Only site A")); status != http.StatusCreated || submitted.Status != "approved" {
		t.Fatalf("submit site A comment = status %d body %s", status, body)
	}

	if status, items, total := listPublicComments(t, app, keyA, "/same-path"); status != http.StatusOK || total != 1 || len(items) != 1 {
		t.Fatalf("site A public list = status %d total %d items %d", status, total, len(items))
	}
	if status, items, total := listPublicComments(t, app, keyB, "/same-path"); status != http.StatusOK || total != 0 || len(items) != 0 {
		t.Fatalf("site B saw site A comments: status %d total %d items %d", status, total, len(items))
	}
	if status, count := publicCommentCount(t, app, keyB, "/same-path"); status != http.StatusOK || count != 0 {
		t.Fatalf("site B count = status %d count %d", status, count)
	}

	if status, items, total := listAdminComments(t, app, token, siteBID, ""); status != http.StatusOK || total != 0 || len(items) != 0 {
		t.Fatalf("site B admin list = status %d total %d items %d", status, total, len(items))
	}
}

func TestPublicCommentPagination(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)
	enableAutoApprove(t, app, token, siteID)

	for _, content := range []string{"one", "two", "three"} {
		if status, _, body := submitComment(t, app, key, commentPayload("/page", content)); status != http.StatusCreated {
			t.Fatalf("submit %q status = %d body=%s", content, status, body)
		}
	}

	recorder := perform(app.engine, http.MethodGet,
		"/api/v1/public/comments?path="+url.QueryEscape("/page")+"&page=2&pageSize=1",
		nil,
		map[string]string{"X-Site-Key": key},
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("pagination status = %d, want 200", recorder.Code)
	}
	var body struct {
		Data struct {
			Items []publicCommentDTO `json:"items"`
			Total int64              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode pagination response: %v", err)
	}
	if body.Data.Total != 3 || len(body.Data.Items) != 1 || body.Data.Items[0].Content != "two" {
		t.Fatalf("unexpected pagination response: %+v", body.Data)
	}
}

func TestCommentThreadDepth(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)
	enableAutoApprove(t, app, token, siteID)

	_, root, _ := submitComment(t, app, key, commentPayload("/thread", "root"))
	_, reply2, body := submitComment(t, app, key, commentPayloadWithParent("/thread", "reply 2", root.ID))
	if reply2.ID == 0 {
		t.Fatalf("create reply 2 failed: %s", body)
	}
	_, reply3, body := submitComment(t, app, key, commentPayloadWithParent("/thread", "reply 3", reply2.ID))
	if reply3.ID == 0 {
		t.Fatalf("create reply 3 failed: %s", body)
	}
	status, _, body := submitComment(t, app, key, commentPayloadWithParent("/thread", "reply 4", reply3.ID))
	if status != http.StatusBadRequest {
		t.Fatalf("depth limit status = %d, want 400 (body=%s)", status, body)
	}

	listStatus, items, total := listPublicComments(t, app, key, "/thread")
	if listStatus != http.StatusOK || total != 1 || len(items) != 1 {
		t.Fatalf("thread list = status %d total %d items %d", listStatus, total, len(items))
	}
	if len(items[0].Replies) != 1 || len(items[0].Replies[0].Replies) != 1 {
		t.Fatalf("unexpected thread depth: %+v", items[0])
	}
	if status, count := publicCommentCount(t, app, key, "/thread"); status != http.StatusOK || count != 3 {
		t.Fatalf("thread count = status %d count %d", status, count)
	}
}

func TestCommentAntiSpam(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)
	enableAutoApprove(t, app, token, siteID)

	honeypot := commentPayload("/spam", "spam")
	honeypot["honeypot"] = "https://spam.example.com"
	if status, _, body := submitComment(t, app, key, honeypot); status != http.StatusBadRequest {
		t.Fatalf("honeypot status = %d, want 400 (body=%s)", status, body)
	}

	tooFast := commentPayload("/spam", "too fast")
	tooFast["formStartedAt"] = time.Now().UnixMilli()
	if status, _, body := submitComment(t, app, key, tooFast); status != http.StatusTooManyRequests {
		t.Fatalf("too fast status = %d, want 429 (body=%s)", status, body)
	}

	duplicate := commentPayload("/spam", "duplicate content")
	duplicate["formStartedAt"] = time.Now().Add(-2 * time.Second).UnixMilli()
	if status, _, body := submitComment(t, app, key, duplicate); status != http.StatusCreated {
		t.Fatalf("first duplicate status = %d, want 201 (body=%s)", status, body)
	}
	if status, _, body := submitComment(t, app, key, duplicate); status != http.StatusConflict {
		t.Fatalf("duplicate status = %d, want 409 (body=%s)", status, body)
	}

	limitedApp := newAdminTestAppWithConfig(t, func(cfg *config.Config) {
		cfg.Comment.RateLimitMax = 1
	})
	limitedToken := login(t, limitedApp, testAdminEmail, testAdminPassword)
	limitedSiteID, limitedKey := createSite(t, limitedApp, limitedToken, "Site Limit", "site-limit", nil)
	enableAutoApprove(t, limitedApp, limitedToken, limitedSiteID)

	if status, _, body := submitComment(t, limitedApp, limitedKey, commentPayload("/limit", "first")); status != http.StatusCreated {
		t.Fatalf("first limited comment status = %d (body=%s)", status, body)
	}
	if status, _, body := submitComment(t, limitedApp, limitedKey, commentPayload("/limit", "second")); status != http.StatusTooManyRequests {
		t.Fatalf("rate limited status = %d, want 429 (body=%s)", status, body)
	}
}

func TestAdminCommentBulkAndDelete(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)

	_, first, _ := submitComment(t, app, key, commentPayload("/bulk", "first"))
	_, second, _ := submitComment(t, app, key, commentPayload("/bulk", "second"))

	status, updated := bulkComments(t, app, token, siteID, []uint{first.ID, second.ID}, "approve")
	if status != http.StatusOK || updated != 2 {
		t.Fatalf("bulk approve = status %d updated %d", status, updated)
	}
	if status, count := publicCommentCount(t, app, key, "/bulk"); status != http.StatusOK || count != 2 {
		t.Fatalf("bulk approved count = status %d count %d", status, count)
	}

	status, updated = bulkComments(t, app, token, siteID, []uint{first.ID, second.ID}, "spam")
	if status != http.StatusOK || updated != 2 {
		t.Fatalf("bulk spam = status %d updated %d", status, updated)
	}
	if status, count := publicCommentCount(t, app, key, "/bulk"); status != http.StatusOK || count != 0 {
		t.Fatalf("bulk spam count = status %d count %d", status, count)
	}

	status, updated = bulkComments(t, app, token, siteID, []uint{first.ID, second.ID}, "delete")
	if status != http.StatusOK || updated != 2 {
		t.Fatalf("bulk delete = status %d updated %d", status, updated)
	}
	if status, items, total := listAdminComments(t, app, token, siteID, ""); status != http.StatusOK || total != 0 || len(items) != 0 {
		t.Fatalf("admin list after bulk delete = status %d total %d items %d", status, total, len(items))
	}
}

func TestAdminCommentDateFilterAndStats(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)
	date := time.Now().UTC().Format("2006-01-02")

	_, first, body := submitComment(t, app, key, commentPayload("/stats", "first"))
	if first.ID == 0 {
		t.Fatalf("create first comment failed: %s", body)
	}
	_, second, body := submitComment(t, app, key, commentPayload("/stats", "second"))
	if second.ID == 0 {
		t.Fatalf("create second comment failed: %s", body)
	}

	approved := perform(app.engine, http.MethodPatch, "/api/v1/admin/comments/"+itoa(first.ID),
		strings.NewReader(`{"status":"approved"}`), withJSON(authHeaders(token)))
	if approved.Code != http.StatusOK {
		t.Fatalf("approve comment status = %d body=%s", approved.Code, approved.Body.String())
	}

	status, items, total := listAdminComments(t, app, token, siteID,
		"from="+date+"&to="+date+"&page=1&pageSize=20")
	if status != http.StatusOK || total != 2 || len(items) != 2 {
		t.Fatalf("date filtered comments = status %d total %d items %d", status, total, len(items))
	}

	status, _, total = listAdminComments(t, app, token, siteID,
		"from=2099-01-01&to=2099-01-02")
	if status != http.StatusOK || total != 0 {
		t.Fatalf("future date comments = status %d total %d", status, total)
	}

	invalid := perform(app.engine, http.MethodGet,
		"/api/v1/admin/sites/"+itoa(siteID)+"/comments?from=24-11-2026",
		nil, authHeaders(token))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid date status = %d, want 400 (body=%s)", invalid.Code, invalid.Body.String())
	}

	stats := perform(app.engine, http.MethodGet,
		"/api/v1/admin/sites/"+itoa(siteID)+"/comments/stats?range=7d",
		nil, authHeaders(token))
	if stats.Code != http.StatusOK {
		t.Fatalf("comment stats status = %d body=%s", stats.Code, stats.Body.String())
	}
	var statsBody struct {
		Data struct {
			Total    int64 `json:"total"`
			Pending  int64 `json:"pending"`
			Approved int64 `json:"approved"`
			Trend    []struct {
				Date     string `json:"date"`
				Comments int64  `json:"comments"`
				Pending  int64  `json:"pending"`
			} `json:"trend"`
		} `json:"data"`
	}
	if err := json.Unmarshal(stats.Body.Bytes(), &statsBody); err != nil {
		t.Fatalf("decode comment stats: %v", err)
	}
	if statsBody.Data.Total != 2 || statsBody.Data.Pending != 1 || statsBody.Data.Approved != 1 {
		t.Fatalf("unexpected comment totals: %+v", statsBody.Data)
	}
	if len(statsBody.Data.Trend) != 7 {
		t.Fatalf("comment trend length = %d, want 7", len(statsBody.Data.Trend))
	}
	today := statsBody.Data.Trend[len(statsBody.Data.Trend)-1]
	if today.Date != date || today.Comments != 2 || today.Pending != 1 {
		t.Fatalf("unexpected today trend: %+v", today)
	}
}

func TestCommentValidation(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)
	enableAutoApprove(t, app, token, siteID)

	missingContent := commentPayload("/validation", "")
	if status, _, body := submitComment(t, app, key, missingContent); status != http.StatusBadRequest {
		t.Fatalf("missing content status = %d, want 400 (body=%s)", status, body)
	}

	invalidEmail := commentPayload("/validation", "email")
	invalidEmail["authorEmail"] = "not-an-email"
	if status, _, body := submitComment(t, app, key, invalidEmail); status != http.StatusBadRequest {
		t.Fatalf("invalid email status = %d, want 400 (body=%s)", status, body)
	}

	invalidWebsite := commentPayload("/validation", "website")
	invalidWebsite["authorWebsite"] = "javascript:alert(1)"
	if status, _, body := submitComment(t, app, key, invalidWebsite); status != http.StatusBadRequest {
		t.Fatalf("invalid website status = %d, want 400 (body=%s)", status, body)
	}

	tooManyLinks := commentPayload("/validation", "https://a.example.com https://b.example.com https://c.example.com https://d.example.com https://e.example.com https://f.example.com")
	if status, _, body := submitComment(t, app, key, tooManyLinks); status != http.StatusBadRequest {
		t.Fatalf("too many links status = %d, want 400 (body=%s)", status, body)
	}

	if status, _ := listPublicCommentsRaw(t, app, key, ""); status != http.StatusBadRequest {
		t.Fatalf("missing list path status = %d, want 400", status)
	}
}

func commentPayload(path, content string) map[string]any {
	return map[string]any{
		"path":          path,
		"authorName":    "Alice",
		"authorEmail":   "alice@example.com",
		"authorWebsite": "https://alice.example.com",
		"content":       content,
	}
}

func commentPayloadWithParent(path, content string, parentID uint) map[string]any {
	payload := commentPayload(path, content)
	payload["parentId"] = parentID
	return payload
}

func submitComment(t *testing.T, app *adminTestApp, key string, payload map[string]any) (int, submittedComment, string) {
	t.Helper()

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode comment: %v", err)
	}
	recorder := perform(app.engine, http.MethodPost, "/api/v1/public/comments",
		strings.NewReader(string(encoded)),
		map[string]string{
			"Content-Type": "application/json",
			"X-Site-Key":   key,
		},
	)
	var body struct {
		Data submittedComment `json:"data"`
	}
	if recorder.Code == http.StatusCreated {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode comment response: %v", err)
		}
	}
	return recorder.Code, body.Data, recorder.Body.String()
}

func listPublicComments(t *testing.T, app *adminTestApp, key, path string) (int, []publicCommentDTO, int64) {
	t.Helper()
	status, body := listPublicCommentsRaw(t, app, key, path)
	if status != http.StatusOK {
		return status, nil, 0
	}
	return status, body.Items, body.Total
}

func listPublicCommentsRaw(t *testing.T, app *adminTestApp, key, path string) (int, struct {
	Items []publicCommentDTO `json:"items"`
	Total int64              `json:"total"`
}) {
	t.Helper()

	target := "/api/v1/public/comments"
	if path != "" {
		target += "?path=" + url.QueryEscape(path)
	}
	recorder := perform(app.engine, http.MethodGet, target, nil, map[string]string{"X-Site-Key": key})
	var envelope struct {
		Data struct {
			Items []publicCommentDTO `json:"items"`
			Total int64              `json:"total"`
		} `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode public comments: %v", err)
		}
	}
	return recorder.Code, envelope.Data
}

func publicCommentCount(t *testing.T, app *adminTestApp, key, path string) (int, int64) {
	t.Helper()
	recorder := perform(app.engine, http.MethodGet,
		"/api/v1/public/comments/count?path="+url.QueryEscape(path), nil,
		map[string]string{"X-Site-Key": key},
	)
	var body struct {
		Data struct {
			Count int64 `json:"count"`
		} `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode comment count: %v", err)
		}
	}
	return recorder.Code, body.Data.Count
}

func listAdminComments(t *testing.T, app *adminTestApp, token string, siteID uint, query string) (int, []adminCommentDTO, int64) {
	t.Helper()
	target := "/api/v1/admin/sites/" + itoa(siteID) + "/comments"
	if query != "" {
		target += "?" + query
	}
	recorder := perform(app.engine, http.MethodGet, target, nil, authHeaders(token))
	var body struct {
		Data struct {
			Items []adminCommentDTO `json:"items"`
			Total int64             `json:"total"`
		} `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode admin comments: %v", err)
		}
	}
	return recorder.Code, body.Data.Items, body.Data.Total
}

func replyComment(t *testing.T, app *adminTestApp, token string, commentID uint, content string) (int, adminCommentDTO) {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"content": content})
	recorder := perform(app.engine, http.MethodPost,
		"/api/v1/admin/comments/"+itoa(commentID)+"/reply",
		strings.NewReader(string(payload)),
		withJSON(authHeaders(token)),
	)
	var body struct {
		Data adminCommentDTO `json:"data"`
	}
	if recorder.Code == http.StatusCreated {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode reply: %v", err)
		}
	}
	return recorder.Code, body.Data
}

func bulkComments(t *testing.T, app *adminTestApp, token string, siteID uint, ids []uint, action string) (int, int64) {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{
		"ids":    ids,
		"action": action,
	})
	recorder := perform(app.engine, http.MethodPost,
		"/api/v1/admin/sites/"+itoa(siteID)+"/comments/bulk",
		strings.NewReader(string(payload)),
		withJSON(authHeaders(token)),
	)
	var body struct {
		Data struct {
			Updated int64 `json:"updated"`
		} `json:"data"`
	}
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode bulk action: %v", err)
		}
	}
	return recorder.Code, body.Data.Updated
}

func enableAutoApprove(t *testing.T, app *adminTestApp, token string, siteID uint) {
	t.Helper()
	recorder := perform(app.engine, http.MethodPatch, "/api/v1/admin/sites/"+itoa(siteID),
		strings.NewReader(`{"commentModeration":"approved"}`),
		withJSON(authHeaders(token)),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("enable auto approve status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}
