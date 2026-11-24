package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashIP(ip, salt string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	if salt == "" {
		sum := sha256.Sum256([]byte(ip))
		return hex.EncodeToString(sum[:])
	}

	mac := hmac.New(sha256.New, []byte(salt))
	_, _ = mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func HashContent(content string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(content)))
	return hex.EncodeToString(sum[:])
}
