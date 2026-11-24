package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"marku/api/internal/model"
	"marku/api/internal/security"
)

func (s *Service) SeedAdmin() (bool, error) {
	email := normalizeEmail(s.auth.AdminEmail)
	if email == "" {
		return false, nil
	}

	if _, err := s.repository.FindAdminByEmail(email); err == nil {
		return false, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("find initial admin: %w", err)
	}

	hash, err := security.HashPassword(s.auth.AdminPassword)
	if err != nil {
		return false, fmt.Errorf("hash initial admin password: %w", err)
	}

	admin := &model.Admin{
		Email:        email,
		PasswordHash: hash,
		Nickname:     "Admin",
		Status:       model.AdminStatusActive,
	}
	if err := s.repository.CreateAdmin(admin); err != nil {
		return false, fmt.Errorf("create initial admin: %w", err)
	}
	return true, nil
}

func (s *Service) Login(email, password string) (*model.Admin, string, time.Time, error) {
	if !s.tokens.Configured() {
		return nil, "", time.Time{}, ErrAuthNotConfigured
	}

	admin, err := s.repository.FindAdminByEmail(normalizeEmail(email))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("find admin: %w", err)
	}
	if admin.Status != model.AdminStatusActive {
		return nil, "", time.Time{}, ErrAdminDisabled
	}
	if !security.CheckPassword(admin.PasswordHash, password) {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	token, expiresAt, err := s.tokens.Issue(admin.ID, admin.Email)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("issue token: %w", err)
	}

	now := time.Now().UTC()
	if err := s.repository.TouchAdminLogin(admin.ID, now); err == nil {
		admin.LastLoginAt = &now
	}

	return admin, token, expiresAt, nil
}

func (s *Service) Authenticate(token string) (*model.Admin, error) {
	if !s.tokens.Configured() {
		return nil, ErrAuthNotConfigured
	}

	claims, err := s.tokens.Parse(strings.TrimSpace(token))
	if err != nil {
		return nil, ErrInvalidToken
	}

	adminID, err := claims.AdminID()
	if err != nil {
		return nil, ErrInvalidToken
	}

	admin, err := s.repository.FindAdminByID(adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("find admin: %w", err)
	}
	if admin.Status != model.AdminStatusActive {
		return nil, ErrAdminDisabled
	}
	return admin, nil
}

func (s *Service) ChangePassword(admin *model.Admin, currentPassword, newPassword string) error {
	if admin == nil {
		return ErrInvalidToken
	}
	if !security.CheckPassword(admin.PasswordHash, currentPassword) {
		return ErrInvalidCurrentPassword
	}
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	if err := s.repository.UpdateAdminPassword(admin.ID, hash); err != nil {
		return fmt.Errorf("update admin password: %w", err)
	}
	admin.PasswordHash = hash
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
