package service

import (
	"fmt"
	"time"

	"marku/api/internal/model"
)

type LikeState struct {
	Likes uint64 `json:"likes"`
	Liked bool   `json:"liked"`
}

type LikePageItem struct {
	Path  string `json:"path"`
	Likes uint64 `json:"likes"`
}

type LikeTrendPoint struct {
	Date    string `json:"date"`
	Likes   uint64 `json:"likes"`
	Unlikes uint64 `json:"unlikes"`
	Net     int64  `json:"net"`
}

type SiteLikeStats struct {
	Path       string           `json:"path,omitempty"`
	TotalLikes uint64           `json:"totalLikes"`
	TopPages   []LikePageItem   `json:"topPages,omitempty"`
	Trend      []LikeTrendPoint `json:"trend"`
}

func (s *Service) SetLike(site *model.Site, rawPath, visitorID string, liked bool) (LikeState, error) {
	if site == nil {
		return LikeState{}, fmt.Errorf("%w: site is required", ErrInvalidLikeInput)
	}
	if site.Status != model.SiteStatusActive {
		return LikeState{}, ErrSiteDisabled
	}

	path, err := normalizePagePath(rawPath)
	if err != nil {
		return LikeState{}, err
	}
	visitorID, err = normalizeRequiredVisitorID(visitorID)
	if err != nil {
		return LikeState{}, err
	}

	now := time.Now().UTC()
	likes, _, err := s.repository.SetLike(
		site.ID,
		path,
		visitorID,
		liked,
		now.Format("2006-01-02"),
		now,
	)
	if err != nil {
		return LikeState{}, fmt.Errorf("set like: %w", err)
	}
	return LikeState{Likes: likes, Liked: liked}, nil
}

func (s *Service) ToggleLike(site *model.Site, rawPath, visitorID string) (LikeState, error) {
	if site == nil {
		return LikeState{}, fmt.Errorf("%w: site is required", ErrInvalidLikeInput)
	}
	if site.Status != model.SiteStatusActive {
		return LikeState{}, ErrSiteDisabled
	}

	path, err := normalizePagePath(rawPath)
	if err != nil {
		return LikeState{}, err
	}
	visitorID, err = normalizeRequiredVisitorID(visitorID)
	if err != nil {
		return LikeState{}, err
	}

	now := time.Now().UTC()
	likes, liked, err := s.repository.ToggleLike(
		site.ID,
		path,
		visitorID,
		now.Format("2006-01-02"),
		now,
	)
	if err != nil {
		return LikeState{}, fmt.Errorf("toggle like: %w", err)
	}
	return LikeState{Likes: likes, Liked: liked}, nil
}

func (s *Service) GetLikeState(siteID uint, rawPath, visitorID string) (LikeState, error) {
	path, err := normalizePagePath(rawPath)
	if err != nil {
		return LikeState{}, err
	}
	visitorID, err = normalizeVisitorID(visitorID)
	if err != nil {
		return LikeState{}, fmt.Errorf("%w: %v", ErrInvalidLikeInput, err)
	}

	likes, liked, err := s.repository.GetLikeState(siteID, path, visitorID)
	if err != nil {
		return LikeState{}, fmt.Errorf("get like state: %w", err)
	}
	return LikeState{Likes: likes, Liked: liked}, nil
}

func (s *Service) GetSiteLikeStats(siteID uint, rawPath string, days, topLimit int) (*SiteLikeStats, error) {
	if _, err := s.GetSite(siteID); err != nil {
		return nil, err
	}
	if days < 1 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	if topLimit < 1 {
		topLimit = 10
	}
	if topLimit > 50 {
		topLimit = 50
	}

	path := ""
	if rawPath != "" {
		normalized, err := normalizePagePath(rawPath)
		if err != nil {
			return nil, err
		}
		path = normalized
	}

	var total uint64
	var topPages []LikePageItem
	if path != "" {
		value, err := s.repository.GetPathLikeTotal(siteID, path)
		if err != nil {
			return nil, fmt.Errorf("get path like total: %w", err)
		}
		total = value
	} else {
		value, err := s.repository.GetSiteLikeTotals(siteID)
		if err != nil {
			return nil, fmt.Errorf("get site like totals: %w", err)
		}
		total = value

		topStats, err := s.repository.ListTopLikedPages(siteID, topLimit)
		if err != nil {
			return nil, fmt.Errorf("list top liked pages: %w", err)
		}
		topPages = make([]LikePageItem, 0, len(topStats))
		for _, stat := range topStats {
			topPages = append(topPages, LikePageItem{Path: stat.Path, Likes: stat.Likes})
		}
	}

	now := time.Now().UTC()
	since := now.AddDate(0, 0, -(days - 1))
	rows, err := s.repository.ListDailyLikeTrend(siteID, path, since.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("list like trend: %w", err)
	}
	byDate := make(map[string]LikeTrendPoint, len(rows))
	for _, row := range rows {
		byDate[row.Date] = LikeTrendPoint{
			Date:    row.Date,
			Likes:   row.Likes,
			Unlikes: row.Unlikes,
			Net:     int64(row.Likes) - int64(row.Unlikes),
		}
	}

	trend := make([]LikeTrendPoint, 0, days)
	for offset := days - 1; offset >= 0; offset-- {
		date := now.AddDate(0, 0, -offset).Format("2006-01-02")
		if point, ok := byDate[date]; ok {
			trend = append(trend, point)
		} else {
			trend = append(trend, LikeTrendPoint{Date: date})
		}
	}

	return &SiteLikeStats{
		Path:       path,
		TotalLikes: total,
		TopPages:   topPages,
		Trend:      trend,
	}, nil
}

func normalizeRequiredVisitorID(raw string) (string, error) {
	value, err := normalizeVisitorID(raw)
	if err != nil {
		return "", fmt.Errorf("%w: visitorId is invalid", ErrInvalidLikeInput)
	}
	if value == "" {
		return "", fmt.Errorf("%w: visitorId is required", ErrInvalidLikeInput)
	}
	return value, nil
}
