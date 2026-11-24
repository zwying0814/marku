package security

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenNotConfigured = errors.New("token manager is not configured")

const tokenIssuer = "marku-api"

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func (c Claims) AdminID() (uint, error) {
	value, err := strconv.ParseUint(c.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid subject: %w", err)
	}
	return uint(value), nil
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *TokenManager) Configured() bool {
	return len(m.secret) > 0
}

func (m *TokenManager) Issue(adminID uint, email string) (string, time.Time, error) {
	if !m.Configured() {
		return "", time.Time{}, ErrTokenNotConfigured
	}

	now := time.Now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(adminID), 10),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	if !m.Configured() {
		return nil, ErrTokenNotConfigured
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method %q", token.Method.Alg())
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
