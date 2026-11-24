package model

import "time"

type Base struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func MigrationModels() []any {
	return []any{
		&Admin{},
		&Site{},
		&Comment{},
		&PageStat{},
		&PageViewDaily{},
		&PageViewVisitor{},
		&LikeStat{},
		&LikeRecord{},
		&LikeDaily{},
	}
}
