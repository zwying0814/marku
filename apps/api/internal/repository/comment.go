package repository

import (
	"time"

	"marku/api/internal/model"
)

type CommentListFilter struct {
	SiteID   uint
	Status   string
	Path     string
	Search   string
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
	Sort     string
}

type SiteCommentTotals struct {
	Total    int64
	Pending  int64
	Approved int64
}

type DailyCommentCount struct {
	Date     string
	Comments int64
	Pending  int64
}

func (s *Store) CreateComment(comment *model.Comment) error {
	return s.db.Create(comment).Error
}

func (s *Store) FindCommentByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := s.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *Store) FindCommentByIDAndSite(id, siteID uint) (*model.Comment, error) {
	var comment model.Comment
	if err := s.db.Where("id = ? AND site_id = ?", id, siteID).First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *Store) ListApprovedCommentsByPath(siteID uint, path string) ([]model.Comment, error) {
	var comments []model.Comment
	err := s.db.
		Where("site_id = ? AND path = ? AND status = ?", siteID, path, model.CommentStatusApproved).
		Order("created_at ASC, id ASC").
		Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *Store) CountCommentsByIP(siteID uint, ipHash string, since time.Time) (int64, error) {
	var count int64
	err := s.db.Model(&model.Comment{}).
		Where("site_id = ? AND ip_hash = ? AND created_at >= ?", siteID, ipHash, since).
		Count(&count).Error
	return count, err
}

func (s *Store) CommentExistsSince(siteID uint, path, ipHash, contentHash string, since time.Time) (bool, error) {
	var count int64
	err := s.db.Model(&model.Comment{}).
		Where(
			"site_id = ? AND path = ? AND ip_hash = ? AND content_hash = ? AND created_at >= ?",
			siteID, path, ipHash, contentHash, since,
		).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) ListComments(filter CommentListFilter) ([]model.Comment, int64, error) {
	query := s.db.Model(&model.Comment{}).Where("site_id = ?", filter.SiteID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Path != "" {
		query = query.Where("path = ?", filter.Path)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("author_name LIKE ? OR author_email LIKE ? OR content LIKE ?", like, like, like)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at < ?", *filter.To)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at DESC, id DESC"
	if filter.Sort == "asc" {
		order = "created_at ASC, id ASC"
	}

	var comments []model.Comment
	if err := query.
		Order(order).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (s *Store) UpdateCommentStatus(id uint, status string) (int64, error) {
	result := s.db.Model(&model.Comment{}).Where("id = ?", id).Update("status", status)
	return result.RowsAffected, result.Error
}

func (s *Store) UpdateCommentsStatusInSite(siteID uint, ids []uint, status string) (int64, error) {
	result := s.db.Model(&model.Comment{}).
		Where("site_id = ? AND id IN ?", siteID, ids).
		Update("status", status)
	return result.RowsAffected, result.Error
}

func (s *Store) DeleteComment(id uint) (int64, error) {
	result := s.db.Delete(&model.Comment{}, id)
	return result.RowsAffected, result.Error
}

func (s *Store) DeleteCommentsInSite(siteID uint, ids []uint) (int64, error) {
	result := s.db.
		Where("site_id = ? AND id IN ?", siteID, ids).
		Delete(&model.Comment{})
	return result.RowsAffected, result.Error
}

func (s *Store) GetSiteCommentTotals(siteID uint) (SiteCommentTotals, error) {
	var totals SiteCommentTotals
	err := s.db.Model(&model.Comment{}).
		Where("site_id = ?", siteID).
		Select(
			"COUNT(*) AS total, "+
				"COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS pending, "+
				"COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS approved",
			model.CommentStatusPending,
			model.CommentStatusApproved,
		).
		Scan(&totals).Error
	return totals, err
}

func (s *Store) ListDailyCommentCounts(siteID uint, since time.Time) ([]DailyCommentCount, error) {
	var comments []model.Comment
	err := s.db.
		Select("created_at, status").
		Where("site_id = ? AND created_at >= ?", siteID, since).
		Order("created_at ASC").
		Find(&comments).Error
	if err != nil {
		return nil, err
	}

	byDate := make(map[string]DailyCommentCount)
	for _, comment := range comments {
		date := comment.CreatedAt.UTC().Format("2006-01-02")
		count := byDate[date]
		count.Date = date
		count.Comments++
		if comment.Status == model.CommentStatusPending {
			count.Pending++
		}
		byDate[date] = count
	}

	counts := make([]DailyCommentCount, 0, len(byDate))
	for _, count := range byDate {
		counts = append(counts, count)
	}
	return counts, nil
}
