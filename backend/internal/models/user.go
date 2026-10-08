package models

import "time"

type User struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UUID              string    `gorm:"uniqueIndex;size:36" json:"uuid"`
	Email             string    `gorm:"uniqueIndex" json:"email"`
	SubscriptionToken string    `gorm:"uniqueIndex;size:64" json:"subscription_token,omitempty"`
	TrafficGB         int64     `json:"traffic_gb"`
	UsedBytes         int64     `json:"-"`
	ExpiresAt         time.Time `json:"expires_at"`
	Active            bool      `gorm:"default:true" json:"active"`
	EntryNodeID       *uint     `json:"entry_node_id,omitempty"`
	ExitNodeID        *uint     `json:"exit_node_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// IsExpired reports whether the account past expires_at (zero means no expiry).
func (u *User) IsExpired() bool {
	if u == nil || u.ExpiresAt.IsZero() {
		return false
	}
	return !u.ExpiresAt.After(time.Now())
}

// IsQuotaExceeded reports whether used_bytes reached the traffic_gb limit (0 = unlimited).
func (u *User) IsQuotaExceeded() bool {
	if u == nil || u.TrafficGB <= 0 {
		return false
	}
	limit := u.TrafficGB * 1024 * 1024 * 1024
	return u.UsedBytes >= limit
}

// AccessAllowed is true when the user may use proxy/client APIs.
func (u *User) AccessAllowed() bool {
	return u != nil && u.Active && !u.IsExpired() && !u.IsQuotaExceeded()
}
