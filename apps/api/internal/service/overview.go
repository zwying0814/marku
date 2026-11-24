package service

import (
	"fmt"
	"time"
)

type AdminOverview struct {
	Sites           int64                `json:"sites"`
	Comments        int64                `json:"comments"`
	PendingComments int64                `json:"pendingComments"`
	Views           uint64               `json:"views"`
	Likes           uint64               `json:"likes"`
	Trend           []OverviewTrendPoint `json:"trend"`
}

type OverviewTrendPoint struct {
	Date     string `json:"date"`
	Views    uint64 `json:"views"`
	Likes    uint64 `json:"likes"`
	Unlikes  uint64 `json:"unlikes"`
	NetLikes int64  `json:"netLikes"`
}

func (s *Service) GetAdminOverview(days int) (*AdminOverview, error) {
	if days < 1 {
		days = 7
	}
	if days > 365 {
		days = 365
	}

	totals, err := s.repository.GetAdminOverviewTotals()
	if err != nil {
		return nil, fmt.Errorf("get admin overview totals: %w", err)
	}

	now := time.Now().UTC()
	since := now.AddDate(0, 0, -(days - 1))
	viewRows, err := s.repository.ListOverviewViewTrend(since.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("list overview view trend: %w", err)
	}
	likeRows, err := s.repository.ListOverviewLikeTrend(since.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("list overview like trend: %w", err)
	}

	type point struct {
		views   uint64
		likes   uint64
		unlikes uint64
	}
	byDate := make(map[string]point, len(viewRows)+len(likeRows))
	for _, row := range viewRows {
		value := byDate[row.Date]
		value.views = row.Views
		byDate[row.Date] = value
	}
	for _, row := range likeRows {
		value := byDate[row.Date]
		value.likes = row.Likes
		value.unlikes = row.Unlikes
		byDate[row.Date] = value
	}

	trend := make([]OverviewTrendPoint, 0, days)
	for offset := days - 1; offset >= 0; offset-- {
		date := now.AddDate(0, 0, -offset).Format("2006-01-02")
		value := byDate[date]
		trend = append(trend, OverviewTrendPoint{
			Date:     date,
			Views:    value.views,
			Likes:    value.likes,
			Unlikes:  value.unlikes,
			NetLikes: int64(value.likes) - int64(value.unlikes),
		})
	}

	return &AdminOverview{
		Sites:           totals.Sites,
		Comments:        totals.Comments,
		PendingComments: totals.PendingComments,
		Views:           totals.Views,
		Likes:           totals.Likes,
		Trend:           trend,
	}, nil
}
