package models

import "time"

type ClientStatsReport struct {
	ID          uint      `gorm:"primaryKey"`
	DeviceToken string    `gorm:"index:idx_device_session,unique;size:64;not null"`
	SessionID   string    `gorm:"index:idx_device_session,unique;size:128;not null"`
	BytesIn     int64     `gorm:"not null"`
	BytesOut    int64     `gorm:"not null"`
	ReportedAt  time.Time `gorm:"index;not null"`
}
