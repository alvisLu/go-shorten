package model

import (
	"time"

	"gorm.io/gorm"
)

type URL struct {
	ID          int64          `gorm:"primaryKey"`
	Code        string         `gorm:"uniqueIndex;not null"`
	OriginalURL string         `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}