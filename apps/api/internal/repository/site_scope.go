package repository

import "gorm.io/gorm"

// SiteScope is the required entry point for future site-owned content queries.
// Every content repository must use ForSite(siteID) and must not query by path alone.
type SiteScope struct {
	db     *gorm.DB
	siteID uint
}

func (s *Store) ForSite(siteID uint) *SiteScope {
	return &SiteScope{
		db:     s.db,
		siteID: siteID,
	}
}

func (s *SiteScope) SiteID() uint {
	return s.siteID
}

func (s *SiteScope) Query() *gorm.DB {
	return s.db.Where("site_id = ?", s.siteID)
}
