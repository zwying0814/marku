package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

const (
	SiteStatusActive   = "active"
	SiteStatusDisabled = "disabled"

	CommentModerationPending  = "pending"
	CommentModerationApproved = "approved"
)

type Site struct {
	Base
	Name              string         `gorm:"size:100;not null" json:"name"`
	Slug              string         `gorm:"uniqueIndex;size:64;not null" json:"slug"`
	Domain            string         `gorm:"size:255" json:"domain"`
	Status            string         `gorm:"size:20;not null;default:active" json:"status"`
	SiteKeyPrefix     string         `gorm:"index;size:16" json:"siteKeyPrefix"`
	SiteKeyHash       string         `gorm:"size:64" json:"-"`
	AllowedOrigins    string         `gorm:"column:allowed_origins;type:text" json:"-"`
	CommentModeration string         `gorm:"size:20;not null;default:pending" json:"commentModeration"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *Site) SetAllowedOrigins(origins []string) error {
	if origins == nil {
		origins = []string{}
	}
	encoded, err := json.Marshal(origins)
	if err != nil {
		return err
	}
	s.AllowedOrigins = string(encoded)
	return nil
}

func (s Site) AllowedOriginList() []string {
	if s.AllowedOrigins == "" {
		return nil
	}

	var origins []string
	if err := json.Unmarshal([]byte(s.AllowedOrigins), &origins); err != nil {
		return nil
	}
	return origins
}
