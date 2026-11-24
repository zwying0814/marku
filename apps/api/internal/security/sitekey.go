package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	siteKeyPrefix       = "mk_"
	SiteKeyPrefixLength = 12
)

func GenerateSiteKey() (plain string, prefix string, hash string, err error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", "", "", fmt.Errorf("generate site key: %w", err)
	}

	plain = siteKeyPrefix + base64.RawURLEncoding.EncodeToString(random)
	return plain, SiteKeyPrefix(plain), HashSiteKey(plain), nil
}

func SiteKeyPrefix(key string) string {
	if len(key) <= SiteKeyPrefixLength {
		return key
	}
	return key[:SiteKeyPrefixLength]
}

func HashSiteKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func VerifySiteKey(key, hash string) bool {
	if key == "" || hash == "" {
		return false
	}
	expected := HashSiteKey(key)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(hash)) == 1
}

func ValidSiteKey(key string) bool {
	return strings.HasPrefix(key, siteKeyPrefix) && len(key) >= 32
}
