package models

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UUID       string    `gorm:"uniqueIndex;size:36" json:"uuid"`
	Email      string    `gorm:"uniqueIndex" json:"email"`
	TrafficGB  int64     `json:"traffic_gb"`
	UsedBytes  int64     `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	Active     bool      `gorm:"default:true" json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}
