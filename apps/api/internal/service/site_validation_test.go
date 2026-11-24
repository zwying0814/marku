package service

import (
	"strings"
	"testing"
)

func TestSiteValidationHelpers(t *testing.T) {
	if err := validateSiteName("My Site"); err != nil {
		t.Fatalf("valid name: %v", err)
	}
	if err := validateSiteName(""); err == nil {
		t.Fatal("empty name should fail")
	}
	if err := validateSlug("my-site"); err != nil {
		t.Fatalf("valid slug: %v", err)
	}
	if err := validateSlug("Bad Slug"); err == nil {
		t.Fatal("invalid slug should fail")
	}
	if err := validateDomain(strings.Repeat("a", 256)); err == nil {
		t.Fatal("long domain should fail")
	}
	if err := validateStatus("active"); err != nil {
		t.Fatalf("valid status: %v", err)
	}
	if err := validateStatus("unknown"); err == nil {
		t.Fatal("invalid status should fail")
	}
	if err := validateModeration("pending"); err != nil {
		t.Fatalf("valid moderation: %v", err)
	}
	if err := validateModeration("unknown"); err == nil {
		t.Fatal("invalid moderation should fail")
	}

	origins, err := normalizeOrigins([]string{"https://example.com/", "https://example.com"})
	if err != nil {
		t.Fatalf("normalize origins: %v", err)
	}
	if len(origins) != 1 || origins[0] != "https://example.com" {
		t.Fatalf("origins = %#v", origins)
	}
	if _, err := normalizeOrigins([]string{"*"}); err == nil {
		t.Fatal("wildcard origin should fail")
	}
	if _, err := normalizeOrigins([]string{"https://example.com/path"}); err == nil {
		t.Fatal("origin with path should fail")
	}
}
