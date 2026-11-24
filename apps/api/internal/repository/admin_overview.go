package repository

import "marku/api/internal/model"

type AdminOverviewTotals struct {
	Sites           int64
	Comments        int64
	PendingComments int64
	Views           uint64
	Likes           uint64
}

type DailyOverviewViews struct {
	Date  string `gorm:"column:date"`
	Views uint64 `gorm:"column:views"`
}

type DailyOverviewLikes struct {
	Date    string `gorm:"column:date"`
	Likes   uint64 `gorm:"column:likes"`
	Unlikes uint64 `gorm:"column:unlikes"`
}

func (s *Store) GetAdminOverviewTotals() (AdminOverviewTotals, error) {
	var totals AdminOverviewTotals
	if err := s.db.Model(&model.Site{}).Count(&totals.Sites).Error; err != nil {
		return totals, err
	}
	if err := s.db.Model(&model.Comment{}).Count(&totals.Comments).Error; err != nil {
		return totals, err
	}
	if err := s.db.Model(&model.Comment{}).
		Where("status = ?", model.CommentStatusPending).
		Count(&totals.PendingComments).Error; err != nil {
		return totals, err
	}
	if err := s.db.Model(&model.PageStat{}).
		Select("COALESCE(SUM(views), 0)").
		Scan(&totals.Views).Error; err != nil {
		return totals, err
	}
	if err := s.db.Model(&model.LikeStat{}).
		Select("COALESCE(SUM(likes), 0)").
		Scan(&totals.Likes).Error; err != nil {
		return totals, err
	}
	return totals, nil
}

func (s *Store) ListOverviewViewTrend(sinceDate string) ([]DailyOverviewViews, error) {
	var rows []DailyOverviewViews
	err := s.db.Model(&model.PageViewDaily{}).
		Select("date, COALESCE(SUM(views), 0) AS views").
		Where("date >= ?", sinceDate).
		Group("date").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}

func (s *Store) ListOverviewLikeTrend(sinceDate string) ([]DailyOverviewLikes, error) {
	var rows []DailyOverviewLikes
	err := s.db.Model(&model.LikeDaily{}).
		Select("date, COALESCE(SUM(likes), 0) AS likes, COALESCE(SUM(unlikes), 0) AS unlikes").
		Where("date >= ?", sinceDate).
		Group("date").
		Order("date ASC").
		Scan(&rows).Error
	return rows, err
}
