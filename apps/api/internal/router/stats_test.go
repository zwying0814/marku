package router

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestPublicAggregateStats(t *testing.T) {
	app := newAdminTestApp(t)
	token := login(t, app, testAdminEmail, testAdminPassword)
	siteID, key := createSite(t, app, token, "Site A", "site-a", nil)
	enableAutoApprove(t, app, token, siteID)

	submitComment(t, app, key, commentPayload("/stats", "first"))
	submitComment(t, app, key, commentPayload("/stats", "second"))
	recordView(t, app, key, "/stats", "", "Mozilla/5.0")
	recordView(t, app, key, "/stats", "", "Mozilla/5.0")
	sendLike(t, app, key, http.MethodPut, "/stats", "visitor-a")
	sendLike(t, app, key, http.MethodPut, "/stats", "visitor-b")

	query := url.Values{}
	query.Set("paths", "/stats,/missing")
	query.Set("visitorId", "visitor-a")
	recorder := perform(app.engine, http.MethodGet, "/api/v1/public/stats?"+query.Encode(), nil, map[string]string{
		"X-Site-Key": key,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("aggregate stats status = %d body=%s", recorder.Code, recorder.Body.String())
	}

	var body struct {
		Data struct {
			Items []struct {
				Path        string `json:"path"`
				Comments    int64  `json:"comments"`
				Views       uint64 `json:"views"`
				UniqueViews uint64 `json:"uniqueViews"`
				Likes       uint64 `json:"likes"`
				Liked       bool   `json:"liked"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode aggregate stats: %v", err)
	}
	if len(body.Data.Items) != 2 {
		t.Fatalf("aggregate items = %+v", body.Data.Items)
	}
	first := body.Data.Items[0]
	if first.Path != "/stats" || first.Comments != 2 || first.Views != 2 || first.Likes != 2 || !first.Liked {
		t.Fatalf("unexpected aggregate stats: %+v", first)
	}
	missing := body.Data.Items[1]
	if missing.Path != "/missing" || missing.Comments != 0 || missing.Views != 0 || missing.Likes != 0 || missing.Liked {
		t.Fatalf("unexpected missing stats: %+v", missing)
	}
}
