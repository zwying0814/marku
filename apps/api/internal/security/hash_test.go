package security

import "testing"

func TestHashIPAndContent(t *testing.T) {
	first := HashIP("192.0.2.10", "test-salt")
	second := HashIP("192.0.2.10", "test-salt")
	if first == "" || first != second {
		t.Fatal("IP hash should be deterministic and non-empty")
	}
	if first == HashIP("192.0.2.11", "test-salt") {
		t.Fatal("different IPs should have different hashes")
	}
	if first == HashIP("192.0.2.10", "another-salt") {
		t.Fatal("different salts should have different hashes")
	}

	if HashContent("hello") != HashContent(" hello ") {
		t.Fatal("content hash should trim surrounding whitespace")
	}
	if HashContent("hello") == HashContent("hello!") {
		t.Fatal("different content should have different hashes")
	}
}
