package models

import "time"

type Traffic struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index"`
	BytesUp    int64
	BytesDown  int64
	RecordedAt time.Time `gorm:"index"`
}
