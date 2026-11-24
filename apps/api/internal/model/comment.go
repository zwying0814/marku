package model

import "gorm.io/gorm"

const (
	CommentStatusPending  = "pending"
	CommentStatusApproved = "approved"
	CommentStatusRejected = "rejected"
	CommentStatusSpam     = "spam"
)

type Comment struct {
	Base
	SiteID        uint           `gorm:"not null;index;index:idx_comments_site_path_status,priority:1" json:"siteId"`
	Path          string         `gorm:"size:512;not null;index:idx_comments_site_path_status,priority:2" json:"path"`
	ParentID      *uint          `gorm:"index" json:"parentId"`
	AuthorName    string         `gorm:"size:100;not null" json:"authorName"`
	AuthorEmail   string         `gorm:"size:255" json:"-"`
	AuthorWebsite string         `gorm:"size:255" json:"authorWebsite"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	ContentHash   string         `gorm:"size:64;index" json:"-"`
	Status        string         `gorm:"size:20;not null;default:pending;index;index:idx_comments_site_path_status,priority:3" json:"status"`
	IsAdmin       bool           `gorm:"not null;default:false" json:"isAdmin"`
	IPHash        string         `gorm:"size:64;index" json:"-"`
	UserAgent     string         `gorm:"size:512" json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
