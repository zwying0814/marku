package security

import (
	"strings"
	"testing"
)

func TestGenerateSiteKey(t *testing.T) {
	key, prefix, hash, err := GenerateSiteKey()
	if err != nil {
		t.Fatalf("generate site key: %v", err)
	}
	if !ValidSiteKey(key) {
		t.Fatalf("generated key is invalid: %q", key)
	}
	if !strings.HasPrefix(key, "mk_") {
		t.Fatalf("key prefix = %q, want mk_", key)
	}
	if len(prefix) != SiteKeyPrefixLength {
		t.Fatalf("prefix length = %d, want %d", len(prefix), SiteKeyPrefixLength)
	}
	if prefix != key[:SiteKeyPrefixLength] {
		t.Fatalf("prefix = %q, want key prefix", prefix)
	}
	if !VerifySiteKey(key, hash) {
		t.Fatal("expected generated key to verify")
	}
	if VerifySiteKey(key+"x", hash) {
		t.Fatal("tampered key should not verify")
	}
	if VerifySiteKey("", hash) {
		t.Fatal("empty key should not verify")
	}
}
