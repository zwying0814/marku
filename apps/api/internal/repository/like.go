package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"marku/api/internal/model"
)

type DailyLikeRow struct {
	Date    string `gorm:"column:date"`
	Likes   uint64 `gorm:"column:likes"`
	Unlikes uint64 `gorm:"column:unlikes"`
}

func (s *Store) SetLike(siteID uint, path, visitorID string, liked bool, date string, now time.Time) (uint64, bool, error) {
	var likes uint64
	var changed bool

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if liked {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.LikeRecord{
				SiteID:    siteID,
				Path:      path,
				VisitorID: visitorID,
			})
			if result.Error != nil {
				return result.Error
			}
			changed = result.RowsAffected > 0
			if changed {
				if err := incrementLikeStat(tx, siteID, path, date, now); err != nil {
					return err
				}
			}
		} else {
			result := tx.Where("site_id = ? AND path = ? AND visitor_id = ?", siteID, path, visitorID).
				Delete(&model.LikeRecord{})
			if result.Error != nil {
				return result.Error
			}
			changed = result.RowsAffected > 0
			if changed {
				if err := decrementLikeStat(tx, siteID, path, date, now); err != nil {
					return err
				}
			}
		}

		current, err := findLikeStat(tx, siteID, path)
		if err != nil {
			return err
		}
		likes = current
		return nil
	})

	return likes, changed, err
}

func (s *Store) ToggleLike(siteID uint, path, visitorID, date string, now time.Time) (uint64, bool, error) {
	var likes uint64
	var liked bool

	err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("site_id = ? AND path = ? AND visitor_id = ?", siteID, path, visitorID).
			Delete(&model.LikeRecord{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			liked = false
			if err := decrementLikeStat(tx, siteID, path, date, now); err != nil {
				return err
			}
		} else {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.LikeRecord{
				SiteID:    siteID,
				Path:      path,
				VisitorID: visitorID,
			})
			if result.Error != nil {
				return result.Error
			}
			liked = true
			if result.RowsAffected > 0 {
				if err := incrementLikeStat(tx, siteID, path, date, now); err != nil {
					return err
				}
			}
		}

		current, err := findLikeStat(tx, siteID, path)
		if err != nil {
			return err
		}
		likes = current
		return nil
	})

	return likes, liked, err
}

func (s *Store) GetLikeState(siteID uint, path, visitorID string) (uint64, bool, error) {
	var likes uint64
	var stat model.LikeStat
	err := s.db.Where("site_id = ? AND path = ?", siteID, path).First(&stat).Error
	if err == nil {
		likes = stat.Likes
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, err
	}

	if visitorID == "" {
		return likes, false, nil
	}
	var count int64
	if err := s.db.Model(&model.LikeRecord{}).
		Where("site_id = ? AND path = ? AND visitor_id = ?", siteID, path, visitorID).
		Count(&count).Error; err != nil {
		return 0, false, err
	}
	return likes, count > 0, nil
}

func (s *Store) GetSiteLikeTotals(siteID uint) (uint64, error) {
	var total uint64
	err := s.db.Model(&model.LikeStat{}).
		Select("COALESCE(SUM(likes), 0)").
		Where("site_id = ?", siteID).
		Scan(&total).Error
	return total, err
}

func (s *Store) GetPathLikeTotal(siteID uint, path string) (uint64, error) {
	var stat model.LikeStat
	err := s.db.Where("site_id = ? AND path = ?", siteID, path).First(&stat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return stat.Likes, nil
}

func (s *Store) ListTopLikedPages(siteID uint, limit int) ([]model.LikeStat, error) {
	var stats []model.LikeStat
	err := s.db.
		Where("site_id = ? AND likes > 0", siteID).
		Order("likes DESC, path ASC").
		Limit(limit).
		Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *Store) ListDailyLikeTrend(siteID uint, path, sinceDate string) ([]DailyLikeRow, error) {
	query := s.db.Model(&model.LikeDaily{}).
		Select("date, COALESCE(SUM(likes), 0) AS likes, COALESCE(SUM(unlikes), 0) AS unlikes").
		Where("site_id = ? AND date >= ?", siteID, sinceDate)
	if path != "" {
		query = query.Where("path = ?", path)
	}

	var rows []DailyLikeRow
	if err := query.Group("date").Order("date ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func incrementLikeStat(tx *gorm.DB, siteID uint, path, date string, now time.Time) error {
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "site_id"},
			{Name: "path"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"likes":      gorm.Expr("likes + 1"),
			"updated_at": now,
		}),
	}).Create(&model.LikeStat{
		SiteID: siteID,
		Path:   path,
		Likes:  1,
	}).Error; err != nil {
		return err
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "site_id"},
			{Name: "path"},
			{Name: "date"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"likes":      gorm.Expr("likes + 1"),
			"updated_at": now,
		}),
	}).Create(&model.LikeDaily{
		SiteID: siteID,
		Path:   path,
		Date:   date,
		Likes:  1,
	}).Error
}

func decrementLikeStat(tx *gorm.DB, siteID uint, path, date string, now time.Time) error {
	if err := tx.Model(&model.LikeStat{}).
		Where("site_id = ? AND path = ?", siteID, path).
		Updates(map[string]any{
			"likes":      gorm.Expr("CASE WHEN likes > 0 THEN likes - 1 ELSE 0 END"),
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "site_id"},
			{Name: "path"},
			{Name: "date"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"unlikes":    gorm.Expr("unlikes + 1"),
			"updated_at": now,
		}),
	}).Create(&model.LikeDaily{
		SiteID:  siteID,
		Path:    path,
		Date:    date,
		Unlikes: 1,
	}).Error
}

func findLikeStat(tx *gorm.DB, siteID uint, path string) (uint64, error) {
	var stat model.LikeStat
	err := tx.Where("site_id = ? AND path = ?", siteID, path).First(&stat).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return stat.Likes, nil
}
