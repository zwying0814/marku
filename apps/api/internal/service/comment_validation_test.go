package service

import "testing"

func TestCommentValidationHelpers(t *testing.T) {
	if _, err := normalizeAuthorName(""); err == nil {
		t.Fatal("empty author name should fail")
	}
	if got, err := normalizeAuthorEmail("alice@example.com"); err != nil || got != "alice@example.com" {
		t.Fatalf("email = %q err=%v", got, err)
	}
	if _, err := normalizeAuthorEmail("not-an-email"); err == nil {
		t.Fatal("invalid email should fail")
	}
	if _, err := normalizeAuthorWebsite("javascript:alert(1)"); err == nil {
		t.Fatal("javascript URL should fail")
	}

	content, err := normalizeCommentContent(" hello ", 20, 1)
	if err != nil || content != "hello" {
		t.Fatalf("content = %q err=%v", content, err)
	}
	if _, err := normalizeCommentContent("https://a.example https://b.example", 100, 1); err == nil {
		t.Fatal("too many links should fail")
	}
	if err := validateCommentStatus("approved"); err != nil {
		t.Fatalf("valid status: %v", err)
	}
	if err := validateCommentStatus("unknown"); err == nil {
		t.Fatal("invalid status should fail")
	}
	if _, err := normalizeRequiredVisitorID(""); err == nil {
		t.Fatal("required visitor ID should fail when empty")
	}
}
