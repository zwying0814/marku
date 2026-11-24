package security

import (
	"errors"
	"testing"
	"time"
)

func TestTokenIssueAndParse(t *testing.T) {
	manager := NewTokenManager("test-secret-at-least-32-characters-long", time.Hour)
	token, expiresAt, err := manager.Issue(42, "admin@example.com")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if token == "" || expiresAt.Before(time.Now()) {
		t.Fatal("invalid issued token metadata")
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	adminID, err := claims.AdminID()
	if err != nil {
		t.Fatalf("admin id: %v", err)
	}
	if adminID != 42 || claims.Email != "admin@example.com" {
		t.Fatalf("unexpected claims: id=%d email=%q", adminID, claims.Email)
	}
}

func TestTokenRejectsInvalidState(t *testing.T) {
	manager := NewTokenManager("test-secret-at-least-32-characters-long", time.Hour)
	token, _, err := manager.Issue(1, "admin@example.com")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	other := NewTokenManager("another-secret-at-least-32-characters", time.Hour)
	if _, err := other.Parse(token); err == nil {
		t.Fatal("token signed with another secret should fail")
	}

	expired := NewTokenManager("test-secret-at-least-32-characters-long", -time.Minute)
	expiredToken, _, err := expired.Issue(1, "admin@example.com")
	if err != nil {
		t.Fatalf("issue expired token: %v", err)
	}
	if _, err := expired.Parse(expiredToken); err == nil {
		t.Fatal("expired token should fail")
	}

	unconfigured := NewTokenManager("", time.Hour)
	if _, _, err := unconfigured.Issue(1, "admin@example.com"); !errors.Is(err, ErrTokenNotConfigured) {
		t.Fatalf("issue error = %v, want ErrTokenNotConfigured", err)
	}
	if _, err := unconfigured.Parse(token); !errors.Is(err, ErrTokenNotConfigured) {
		t.Fatalf("parse error = %v, want ErrTokenNotConfigured", err)
	}
}
