package models

import "time"

type WireGuardPeer struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"uniqueIndex" json:"user_id"`
	PrivateKey   string    `gorm:"size:64;not null" json:"-"`
	PublicKey    string    `gorm:"size:64;not null" json:"public_key"`
	PresharedKey string    `gorm:"size:64" json:"-"`
	ClientIP     string    `gorm:"size:18;not null" json:"client_ip"`
	CreatedAt    time.Time `json:"created_at"`
}
