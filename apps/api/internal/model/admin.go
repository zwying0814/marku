package model

import "time"

const (
	AdminStatusActive   = "active"
	AdminStatusDisabled = "disabled"
)

type Admin struct {
	Base
	Email        string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Nickname     string     `gorm:"size:100;not null" json:"nickname"`
	Status       string     `gorm:"size:20;not null;default:active" json:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
}
