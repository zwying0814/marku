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

const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "123456"
)

type UpdateAdminProfileInput struct {
	CurrentPassword string
	Username        string
	NewPassword     string
}

func (s *Service) SeedAdmin() (bool, error) {
	count, err := s.repository.CountAdmins()
	if err != nil {
		return false, fmt.Errorf("count admins: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	hash, err := security.HashPassword(DefaultAdminPassword)
	if err != nil {
		return false, fmt.Errorf("hash initial admin password: %w", err)
	}

	admin := &model.Admin{
		Email:        DefaultAdminUsername,
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

func (s *Service) UpdateAdminProfile(admin *model.Admin, input UpdateAdminProfileInput) (*model.Admin, error) {
	if admin == nil {
		return nil, ErrInvalidToken
	}
	if !security.CheckPassword(admin.PasswordHash, input.CurrentPassword) {
		return nil, ErrInvalidCurrentPassword
	}

	username := normalizeEmail(input.Username)
	if username == "" {
		username = admin.Email
	}
	if username != admin.Email {
		if err := validateAdminUsername(username); err != nil {
			return nil, err
		}
		existing, err := s.repository.FindAdminByEmail(username)
		switch {
		case err == nil && existing.ID != admin.ID:
			return nil, ErrUsernameTaken
		case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf("find admin by username: %w", err)
		}
	}

	passwordHash := admin.PasswordHash
	if input.NewPassword != "" {
		if len(input.NewPassword) < 8 {
			return nil, ErrWeakPassword
		}
		hash, err := security.HashPassword(input.NewPassword)
		if err != nil {
			return nil, fmt.Errorf("hash new password: %w", err)
		}
		passwordHash = hash
	}

	if err := s.repository.UpdateAdminProfile(admin.ID, username, passwordHash); err != nil {
		return nil, fmt.Errorf("update admin profile: %w", err)
	}

	admin.Email = username
	admin.PasswordHash = passwordHash
	return admin, nil
}

func validateAdminUsername(username string) error {
	if username == "" || len([]rune(username)) > 64 || strings.ContainsAny(username, " \t\r\n") {
		return ErrInvalidAdminUsername
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
