package repository

import (
	"time"

	"gorm.io/gorm"

	"marku/api/internal/model"
)

func (s *Store) FindSiteByID(id uint) (*model.Site, error) {
	var site model.Site
	if err := s.db.First(&site, id).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *Store) FindSiteBySlug(slug string) (*model.Site, error) {
	var site model.Site
	if err := s.db.Where("slug = ?", slug).First(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *Store) SiteSlugExists(slug string) (bool, error) {
	var count int64
	if err := s.db.Unscoped().Model(&model.Site{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) FindSiteByKeyPrefix(prefix string) (*model.Site, error) {
	var site model.Site
	if err := s.db.Where("site_key_prefix = ?", prefix).First(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *Store) CreateSite(site *model.Site) error {
	return s.db.Create(site).Error
}

func (s *Store) UpdateSite(id uint, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now().UTC()

	result := s.db.Model(&model.Site{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Store) UpdateSiteKey(id uint, prefix, hash string) error {
	return s.UpdateSite(id, map[string]any{
		"site_key_prefix": prefix,
		"site_key_hash":   hash,
	})
}

func (s *Store) DeleteSite(id uint) error {
	result := s.db.Delete(&model.Site{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Store) ListSites(page, pageSize int, search string) ([]model.Site, int64, error) {
	query := s.db.Model(&model.Site{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR slug LIKE ? OR domain LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var sites []model.Site
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&sites).Error; err != nil {
		return nil, 0, err
	}

	return sites, total, nil
}
