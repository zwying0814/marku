package model

type LikeStat struct {
	Base
	SiteID uint   `gorm:"not null;index;uniqueIndex:idx_like_stats_site_path,priority:1" json:"siteId"`
	Path   string `gorm:"size:512;not null;uniqueIndex:idx_like_stats_site_path,priority:2" json:"path"`
	Likes  uint64 `gorm:"not null;default:0" json:"likes"`
}

type LikeRecord struct {
	Base
	SiteID    uint   `gorm:"not null;index;uniqueIndex:idx_like_records_site_path_visitor,priority:1" json:"siteId"`
	Path      string `gorm:"size:512;not null;uniqueIndex:idx_like_records_site_path_visitor,priority:2" json:"path"`
	VisitorID string `gorm:"size:128;not null;uniqueIndex:idx_like_records_site_path_visitor,priority:3" json:"visitorId"`
}

type LikeDaily struct {
	Base
	SiteID  uint   `gorm:"not null;index;uniqueIndex:idx_like_dailies_site_path_date,priority:1" json:"siteId"`
	Path    string `gorm:"size:512;not null;uniqueIndex:idx_like_dailies_site_path_date,priority:2" json:"path"`
	Date    string `gorm:"size:10;not null;index;uniqueIndex:idx_like_dailies_site_path_date,priority:3" json:"date"`
	Likes   uint64 `gorm:"not null;default:0" json:"likes"`
	Unlikes uint64 `gorm:"not null;default:0" json:"unlikes"`
}
