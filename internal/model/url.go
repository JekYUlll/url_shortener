package model

import "time"

type URL struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OriginalURL string    `gorm:"column:original_url;type:text;not null"`
	ShortCode   string    `gorm:"column:short_code;type:text;not null;size:100;uniqueIndex"`
	IsCustom    bool      `gorm:"column:is_custom;not null;default:false"`
	ExpiredAt   time.Time `gorm:"column:expired_at;type:timestamp;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
}

func (u *URL) TableName() string {
	return "urls"
}
