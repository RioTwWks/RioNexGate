package models

import "time"

// Invite is a per-user registration token for RioNexTunnel device onboarding.
type Invite struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"index;not null" json:"user_id"`
	Token     string     `gorm:"uniqueIndex;size:64;not null" json:"token"`
	Label     string     `json:"label,omitempty"`
	MaxUses   int        `gorm:"not null;default:1" json:"max_uses"`
	UsedCount int        `gorm:"not null;default:0" json:"used_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Usable reports whether the invite can still be consumed.
func (inv *Invite) Usable() bool {
	if inv == nil {
		return false
	}
	if inv.RevokedAt != nil {
		return false
	}
	if inv.MaxUses > 0 && inv.UsedCount >= inv.MaxUses {
		return false
	}
	if inv.ExpiresAt != nil && !inv.ExpiresAt.After(time.Now()) {
		return false
	}
	return true
}
