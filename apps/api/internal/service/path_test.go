package service

import (
	"errors"
	"testing"
)

func TestNormalizePagePath(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "query and hash", raw: "/posts/hello?utm=1#top", want: "/posts/hello"},
		{name: "trailing slash", raw: "/posts/hello/", want: "/posts/hello"},
		{name: "leading slash", raw: "posts/hello", want: "/posts/hello"},
		{name: "root", raw: "/", want: "/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizePagePath(test.raw)
			if err != nil {
				t.Fatalf("normalize path: %v", err)
			}
			if got != test.want {
				t.Fatalf("path = %q, want %q", got, test.want)
			}
		})
	}

	for _, raw := range []string{"", "/bad\npath"} {
		if _, err := normalizePagePath(raw); !errors.Is(err, ErrInvalidPagePath) {
			t.Fatalf("error for %q = %v, want ErrInvalidPagePath", raw, err)
		}
	}
}

func TestNormalizePathList(t *testing.T) {
	paths, err := normalizePathList([]string{"/a", "/a/", "/b?x=1"}, 10)
	if err != nil {
		t.Fatalf("normalize list: %v", err)
	}
	if len(paths) != 2 || paths[0] != "/a" || paths[1] != "/b" {
		t.Fatalf("paths = %#v", paths)
	}

	if _, err := normalizePathList(nil, 10); !errors.Is(err, ErrInvalidViewInput) {
		t.Fatalf("empty error = %v", err)
	}
	if _, err := normalizePathList([]string{"/a", "/b"}, 1); !errors.Is(err, ErrInvalidViewInput) {
		t.Fatalf("limit error = %v", err)
	}
}

func TestVisitorValidationAndBotDetection(t *testing.T) {
	if got, err := normalizeVisitorID(" visitor-a "); err != nil || got != "visitor-a" {
		t.Fatalf("visitor = %q err=%v", got, err)
	}
	if got, err := normalizeVisitorID(""); err != nil || got != "" {
		t.Fatalf("empty visitor = %q err=%v", got, err)
	}
	if _, err := normalizeVisitorID("bad visitor"); err == nil {
		t.Fatal("visitor with spaces should fail")
	}

	if !isBotUserAgent("Googlebot/2.1") {
		t.Fatal("Googlebot should be detected")
	}
	if isBotUserAgent("Mozilla/5.0") {
		t.Fatal("browser UA should not be detected")
	}
}
