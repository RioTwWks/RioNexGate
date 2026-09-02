package models

import "encoding/json"

const (
	NodeRoleEntry = "entry"
	NodeRoleExit  = "exit"
)

// NodeCredentials holds relay connection parameters stored as JSON in Node.Credentials.
type NodeCredentials struct {
	UUID        string `json:"uuid,omitempty"`
	Encryption  string `json:"encryption,omitempty"`
	Flow        string `json:"flow,omitempty"`
	PublicKey   string `json:"public_key,omitempty"`
	ShortID     string `json:"short_id,omitempty"`
	SNI         string `json:"sni,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Network     string `json:"network,omitempty"`
	Security    string `json:"security,omitempty"`
	Path        string `json:"path,omitempty"`
	Mode        string `json:"mode,omitempty"`
}

func (c NodeCredentials) EncryptionOrDefault() string {
	if c.Encryption != "" {
		return c.Encryption
	}
	return "none"
}

func (c NodeCredentials) NetworkOrDefault() string {
	if c.Network != "" {
		return c.Network
	}
	return "tcp"
}

func (c NodeCredentials) SecurityOrDefault() string {
	if c.Security != "" {
		return c.Security
	}
	return "none"
}

func (c NodeCredentials) FingerprintOrDefault() string {
	if c.Fingerprint != "" {
		return c.Fingerprint
	}
	return "firefox"
}

func ParseNodeCredentials(raw string) NodeCredentials {
	if raw == "" {
		return NodeCredentials{}
	}
	var creds NodeCredentials
	_ = json.Unmarshal([]byte(raw), &creds)
	return creds
}

type Node struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex" json:"name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Active      bool   `gorm:"default:true" json:"active"`
	Role        string `gorm:"default:entry" json:"role"`
	Protocol    string `gorm:"default:vless" json:"protocol"`
	Credentials string `json:"credentials,omitempty"`
	Region      string `json:"region,omitempty"`
	Priority    int    `gorm:"default:100" json:"priority"`
}

func (n Node) ParsedCredentials() NodeCredentials {
	return ParseNodeCredentials(n.Credentials)
}

func (n Node) OutboundTag() string {
	return "exit-" + n.Name
}
