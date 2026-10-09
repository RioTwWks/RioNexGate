package models

import (
	"regexp"
	"strings"
	"time"
)

// UserExit binds a panel user to one exit node with a dedicated inbound identity.
// Multiple rows per user enable country selection inside a single subscription:
// each assignment has its own VLESS UUID + routing email on the entry core.
type UserExit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_exit;not null" json:"user_id"`
	NodeID    uint      `gorm:"uniqueIndex:idx_user_exit;not null;index" json:"node_id"`
	UUID      string    `gorm:"uniqueIndex;size:36;not null" json:"uuid"`
	Email     string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Priority  int       `gorm:"default:100" json:"priority"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

var exitSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// ExitSlug returns a short lowercase label for subscription remarks and route emails.
func ExitSlug(node Node) string {
	raw := strings.TrimSpace(node.Region)
	if raw == "" {
		raw = strings.TrimSpace(node.Name)
	}
	if raw == "" {
		return "exit"
	}
	s := strings.ToLower(raw)
	s = exitSlugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "exit"
	}
	if len(s) > 32 {
		s = s[:32]
	}
	return s
}

// ExitDisplayLabel is the human-readable country/region shown in client profiles.
func ExitDisplayLabel(node Node) string {
	if r := strings.TrimSpace(node.Region); r != "" {
		return r
	}
	if n := strings.TrimSpace(node.Name); n != "" {
		return n
	}
	return "exit"
}

// ExitRouteEmail builds a unique xray client email for routing to a specific exit.
// Primary (first) assignment keeps the panel user email unchanged.
func ExitRouteEmail(userEmail, slug string, primary bool) string {
	if primary || slug == "" {
		return userEmail
	}
	return userEmail + "/" + slug
}

// UserExitAssignment is a resolved exit binding with node metadata for config/subscription.
type UserExitAssignment struct {
	UserExit UserExit
	Node     Node
}
