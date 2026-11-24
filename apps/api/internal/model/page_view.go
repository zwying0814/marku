package model

import "time"

type PageStat struct {
	Base
	SiteID       uint      `gorm:"not null;index;uniqueIndex:idx_page_stats_site_path,priority:1" json:"siteId"`
	Path         string    `gorm:"size:512;not null;uniqueIndex:idx_page_stats_site_path,priority:2" json:"path"`
	Views        uint64    `gorm:"not null;default:0" json:"views"`
	UniqueViews  uint64    `gorm:"not null;default:0" json:"uniqueViews"`
	LastViewedAt time.Time `json:"lastViewedAt"`
}

type PageViewDaily struct {
	Base
	SiteID      uint   `gorm:"not null;index;uniqueIndex:idx_page_view_daily_site_path_date,priority:1" json:"siteId"`
	Path        string `gorm:"size:512;not null;uniqueIndex:idx_page_view_daily_site_path_date,priority:2" json:"path"`
	Date        string `gorm:"size:10;not null;index;uniqueIndex:idx_page_view_daily_site_path_date,priority:3" json:"date"`
	Views       uint64 `gorm:"not null;default:0" json:"views"`
	UniqueViews uint64 `gorm:"not null;default:0" json:"uniqueViews"`
}

type PageViewVisitor struct {
	Base
	SiteID    uint   `gorm:"not null;uniqueIndex:idx_page_view_visitors_site_path_date_visitor,priority:1" json:"siteId"`
	Path      string `gorm:"size:512;not null;uniqueIndex:idx_page_view_visitors_site_path_date_visitor,priority:2" json:"path"`
	Date      string `gorm:"size:10;not null;uniqueIndex:idx_page_view_visitors_site_path_date_visitor,priority:3" json:"date"`
	VisitorID string `gorm:"size:128;not null;uniqueIndex:idx_page_view_visitors_site_path_date_visitor,priority:4" json:"visitorId"`
}
