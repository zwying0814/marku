package repository

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"marku/api/internal/model"
)

func (s *Store) FindAdminByEmail(email string) (*model.Admin, error) {
	var admin model.Admin
	normalized := strings.ToLower(strings.TrimSpace(email))
	if err := s.db.Where("email = ?", normalized).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *Store) FindAdminByID(id uint) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *Store) CreateAdmin(admin *model.Admin) error {
	return s.db.Create(admin).Error
}

func (s *Store) UpdateAdminPassword(id uint, hash string) error {
	result := s.db.Model(&model.Admin{}).Where("id = ?", id).Update("password_hash", hash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Store) TouchAdminLogin(id uint, at time.Time) error {
	result := s.db.Model(&model.Admin{}).Where("id = ?", id).Update("last_login_at", at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
