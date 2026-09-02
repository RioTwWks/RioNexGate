package models

import "time"

type Device struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"index;not null" json:"user_id"`
	Token            string     `gorm:"uniqueIndex;size:64;not null" json:"token"`
	Label            string     `json:"label"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	CachedConfig     string     `gorm:"type:text" json:"-"`
	CachedConfigHash string     `gorm:"size:64" json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
}
