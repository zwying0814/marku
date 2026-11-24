package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"marku/api/internal/model"
)

type ViewTotals struct {
	Views       uint64 `gorm:"column:views"`
	UniqueViews uint64 `gorm:"column:unique_views"`
}

type DailyViewRow struct {
	Date        string `gorm:"column:date"`
	Views       uint64 `gorm:"column:views"`
	UniqueViews uint64 `gorm:"column:unique_views"`
}

func (s *Store) RecordPageView(siteID uint, path, date, visitorID string, dedupe bool, now time.Time) (uint64, uint64, error) {
	var views uint64
	var uniqueViews uint64

	err := s.db.Transaction(func(tx *gorm.DB) error {
		stat := model.PageStat{
			SiteID:       siteID,
			Path:         path,
			Views:        1,
			UniqueViews:  0,
			LastViewedAt: now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "site_id"},
				{Name: "path"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"views":          gorm.Expr("views + 1"),
				"last_viewed_at": now,
				"updated_at":     now,
			}),
		}).Create(&stat).Error; err != nil {
			return err
		}

		daily := model.PageViewDaily{
			SiteID:      siteID,
			Path:        path,
			Date:        date,
			Views:       1,
			UniqueViews: 0,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "site_id"},
				{Name: "path"},
				{Name: "date"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"views":      gorm.Expr("views + 1"),
				"updated_at": now,
			}),
		}).Create(&daily).Error; err != nil {
			return err
		}

		if dedupe && visitorID != "" {
			visitor := model.PageViewVisitor{
				SiteID:    siteID,
				Path:      path,
				Date:      date,
				VisitorID: visitorID,
			}
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&visitor)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				if err := tx.Model(&model.PageStat{}).
					Where("site_id = ? AND path = ?", siteID, path).
					Updates(map[string]any{
						"unique_views": gorm.Expr("unique_views + 1"),
						"updated_at":   now,
					}).Error; err != nil {
					return err
				}
				if err := tx.Model(&model.PageViewDaily{}).
					Where("site_id = ? AND path = ? AND date = ?", siteID, path, date).
					Updates(map[string]any{
						"unique_views": gorm.Expr("unique_views + 1"),
						"updated_at":   now,
					}).Error; err != nil {
					return err
				}
			}
		}

		var current model.PageStat
		if err := tx.Where("site_id = ? AND path = ?", siteID, path).First(&current).Error; err != nil {
			return err
		}
		views = current.Views
		uniqueViews = current.UniqueViews
		return nil
	})

	return views, uniqueViews, err
}

func (s *Store) FindPageStat(siteID uint, path string) (*model.PageStat, error) {
	var stat model.PageStat
	if err := s.db.Where("site_id = ? AND path = ?", siteID, path).First(&stat).Error; err != nil {
		return nil, err
	}
	return &stat, nil
}

func (s *Store) GetPageStats(siteID uint, paths []string) (map[string]model.PageStat, error) {
	if len(paths) == 0 {
		return map[string]model.PageStat{}, nil
	}

	var stats []model.PageStat
	if err := s.db.Where("site_id = ? AND path IN ?", siteID, paths).Find(&stats).Error; err != nil {
		return nil, err
	}

	result := make(map[string]model.PageStat, len(stats))
	for _, stat := range stats {
		result[stat.Path] = stat
	}
	return result, nil
}

func (s *Store) GetSiteViewTotals(siteID uint) (ViewTotals, error) {
	var totals ViewTotals
	err := s.db.Model(&model.PageStat{}).
		Select("COALESCE(SUM(views), 0) AS views, COALESCE(SUM(unique_views), 0) AS unique_views").
		Where("site_id = ?", siteID).
		Scan(&totals).Error
	return totals, err
}

func (s *Store) ListTopPageStats(siteID uint, limit int) ([]model.PageStat, error) {
	var stats []model.PageStat
	err := s.db.
		Where("site_id = ?", siteID).
		Order("views DESC, unique_views DESC, path ASC").
		Limit(limit).
		Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *Store) ListDailyViewTrend(siteID uint, sinceDate string) ([]DailyViewRow, error) {
	var rows []DailyViewRow
	err := s.db.Model(&model.PageViewDaily{}).
		Select("date, COALESCE(SUM(views), 0) AS views, COALESCE(SUM(unique_views), 0) AS unique_views").
		Where("site_id = ? AND date >= ?", siteID, sinceDate).
		Group("date").
		Order("date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
