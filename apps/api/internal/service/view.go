package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"marku/api/internal/model"
)

var botUserAgentMarkers = []string{
	"bot",
	"crawler",
	"spider",
	"slurp",
	"bingpreview",
	"facebookexternalhit",
	"headlesschrome",
	"lighthouse",
	"curl",
	"wget",
	"python-requests",
	"go-http-client",
}

type RecordViewResult struct {
	Views       uint64 `json:"views"`
	UniqueViews uint64 `json:"uniqueViews"`
	Recorded    bool   `json:"recorded"`
}

type PageViewItem struct {
	Path        string `json:"path"`
	Views       uint64 `json:"views"`
	UniqueViews uint64 `json:"uniqueViews"`
}

type ViewTrendPoint struct {
	Date        string `json:"date"`
	Views       uint64 `json:"views"`
	UniqueViews uint64 `json:"uniqueViews"`
}

type SiteViewStats struct {
	TotalViews       uint64           `json:"totalViews"`
	TotalUniqueViews uint64           `json:"totalUniqueViews"`
	TopPages         []PageViewItem   `json:"topPages"`
	Trend            []ViewTrendPoint `json:"trend"`
}

func (s *Service) RecordPageView(site *model.Site, rawPath, visitorID, userAgent string) (RecordViewResult, error) {
	if site == nil {
		return RecordViewResult{}, fmt.Errorf("%w: site is required", ErrInvalidViewInput)
	}
	if site.Status != model.SiteStatusActive {
		return RecordViewResult{}, ErrSiteDisabled
	}

	path, err := normalizePagePath(rawPath)
	if err != nil {
		return RecordViewResult{}, err
	}
	visitorID, err = normalizeVisitorID(visitorID)
	if err != nil {
		return RecordViewResult{}, err
	}

	now := time.Now().UTC()
	if s.view.FilterBots && isBotUserAgent(userAgent) {
		stat, err := s.repository.FindPageStat(site.ID, path)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RecordViewResult{Recorded: false}, nil
		}
		if err != nil {
			return RecordViewResult{}, fmt.Errorf("find page stat: %w", err)
		}
		return RecordViewResult{
			Views:       stat.Views,
			UniqueViews: stat.UniqueViews,
			Recorded:    false,
		}, nil
	}

	views, uniqueViews, err := s.repository.RecordPageView(
		site.ID,
		path,
		now.Format("2006-01-02"),
		visitorID,
		s.view.DedupeByVisitorDay,
		now,
	)
	if err != nil {
		return RecordViewResult{}, fmt.Errorf("record page view: %w", err)
	}
	return RecordViewResult{
		Views:       views,
		UniqueViews: uniqueViews,
		Recorded:    true,
	}, nil
}

func (s *Service) GetPageViews(siteID uint, rawPaths []string) ([]PageViewItem, error) {
	paths, err := normalizePathList(rawPaths, s.view.MaxBatchPaths)
	if err != nil {
		return nil, err
	}

	stats, err := s.repository.GetPageStats(siteID, paths)
	if err != nil {
		return nil, fmt.Errorf("get page stats: %w", err)
	}

	items := make([]PageViewItem, 0, len(paths))
	for _, path := range paths {
		item := PageViewItem{Path: path}
		if stat, ok := stats[path]; ok {
			item.Views = stat.Views
			item.UniqueViews = stat.UniqueViews
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) GetSiteViewStats(siteID uint, days, topLimit int) (*SiteViewStats, error) {
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

	totals, err := s.repository.GetSiteViewTotals(siteID)
	if err != nil {
		return nil, fmt.Errorf("get site view totals: %w", err)
	}
	topStats, err := s.repository.ListTopPageStats(siteID, topLimit)
	if err != nil {
		return nil, fmt.Errorf("list top page stats: %w", err)
	}

	topPages := make([]PageViewItem, 0, len(topStats))
	for _, stat := range topStats {
		topPages = append(topPages, PageViewItem{
			Path:        stat.Path,
			Views:       stat.Views,
			UniqueViews: stat.UniqueViews,
		})
	}

	now := time.Now().UTC()
	since := now.AddDate(0, 0, -(days - 1))
	rows, err := s.repository.ListDailyViewTrend(siteID, since.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("list view trend: %w", err)
	}
	byDate := make(map[string]ViewTrendPoint, len(rows))
	for _, row := range rows {
		byDate[row.Date] = ViewTrendPoint{
			Date:        row.Date,
			Views:       row.Views,
			UniqueViews: row.UniqueViews,
		}
	}

	trend := make([]ViewTrendPoint, 0, days)
	for offset := days - 1; offset >= 0; offset-- {
		date := now.AddDate(0, 0, -offset).Format("2006-01-02")
		if point, ok := byDate[date]; ok {
			trend = append(trend, point)
		} else {
			trend = append(trend, ViewTrendPoint{Date: date})
		}
	}

	return &SiteViewStats{
		TotalViews:       totals.Views,
		TotalUniqueViews: totals.UniqueViews,
		TopPages:         topPages,
		Trend:            trend,
	}, nil
}

func normalizeVisitorID(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if len(value) > 128 || strings.ContainsAny(value, "\r\n\t ") {
		return "", fmt.Errorf("%w: visitorId is invalid", ErrInvalidViewInput)
	}
	return value, nil
}

func isBotUserAgent(userAgent string) bool {
	lower := strings.ToLower(userAgent)
	for _, marker := range botUserAgentMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
