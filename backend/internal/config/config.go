package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Core     CoreConfig     `mapstructure:"core"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Limits   LimitsConfig   `mapstructure:"limits"`

	// path is the loaded CONFIG_PATH (not from YAML).
	path string
}

type ServerConfig struct {
	Port             int    `mapstructure:"port"`
	APIKey           string `mapstructure:"api_key"`
	PublicBaseURL    string `mapstructure:"public_base_url"`
	ClientSOCKS5Port int    `mapstructure:"client_socks5_port"`
	// RegistrationSecret, when set, allows POST /api/client/register with
	// header X-Registration-Secret (for RioNexTunnel without exposing the admin API key).
	RegistrationSecret string `mapstructure:"registration_secret"`
	// AllowOpenRegister enables unauthenticated device registration (insecure; LAN/dev only).
	AllowOpenRegister bool `mapstructure:"allow_open_register"`
	// AllowPrivateProbes permits SSRF-sensitive probes (stealth test-dest, node health)
	// against loopback/private/link-local addresses. Needed for LAN multihop; keep false on public hosts.
	AllowPrivateProbes bool `mapstructure:"allow_private_probes"`
	// CORSOrigins lists allowed browser origins. Empty defaults to localhost panel URLs.
	// Set to ["*"] only if you intentionally allow any origin.
	CORSOrigins []string `mapstructure:"cors_origins"`
	// EnableDocs serves /api/docs and /api/openapi.yaml (default true).
	EnableDocs *bool `mapstructure:"enable_docs"`
	// SubscriptionSigSecret signs time-limited subscription URLs. Empty → use api_key.
	SubscriptionSigSecret string `mapstructure:"subscription_sig_secret"`
	// SubscriptionDefaultTTLHours is the default TTL for minted signed subscription links (0 = 168h).
	SubscriptionDefaultTTLHours int `mapstructure:"subscription_default_ttl_hours"`
	// RequireSubscriptionSig rejects bare /api/subscription/{token} without exp+sig.
	RequireSubscriptionSig bool `mapstructure:"require_subscription_sig"`
	// AutoDeactivateOnQuota sets user.active=false when traffic quota is exceeded (default true).
	AutoDeactivateOnQuota *bool `mapstructure:"auto_deactivate_on_quota"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type CoreConfig struct {
	Type       string         `mapstructure:"type"`
	ListenPort int            `mapstructure:"listen_port"`
	PublicHost string         `mapstructure:"public_host"`
	Xray       XrayConfig     `mapstructure:"xray"`
	Singbox    SingboxConfig  `mapstructure:"singbox"`
	Skadi      SkadiConfig    `mapstructure:"skadi"`
	Stealth    StealthConfig  `mapstructure:"stealth"`
	Multihop   MultihopConfig `mapstructure:"multihop"`
	StatsPoll  int            `mapstructure:"stats_poll_seconds"`
}

// MultihopConfig controls entry→exit chain generation on the local core.
type MultihopConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	LocalRole string `mapstructure:"local_role"`
}

func (m *MultihopConfig) IsEntryNode() bool {
	return m.Enabled && m.LocalRole == "entry"
}

type XrayConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	BinaryPath string `mapstructure:"binary_path"`
	// APIAddress is where the backend connects to query Xray stats.
	APIAddress string `mapstructure:"api_address"`
	// APIListen is the bind address written into the generated Xray config.
	// Defaults to loopback derived from APIAddress (never 0.0.0.0).
	APIListen string `mapstructure:"api_listen"`
}

type SingboxConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	BinaryPath string `mapstructure:"binary_path"`
	APIAddress string `mapstructure:"api_address"`
}

// SkadiConfig holds paths and API settings for SkadiCore (https://github.com/RioTwWks/SkadiCore).
type SkadiConfig struct {
	ConfigPath     string `mapstructure:"config_path"`
	BinaryPath     string `mapstructure:"binary_path"`
	APIAddress     string `mapstructure:"api_address"`
	APIToken       string `mapstructure:"api_token"`
	MetricsAddress string `mapstructure:"metrics_address"`
}

// StealthConfig holds anti-DPI transport presets (Reality + XHTTP / Vision / TLS).
type StealthConfig struct {
	Enabled     bool                 `mapstructure:"enabled"`
	Fingerprint string               `mapstructure:"fingerprint"`
	Reality     StealthRealityConfig `mapstructure:"reality"`
	XHTTP       StealthXHTTPConfig   `mapstructure:"xhttp"`
	Vision      StealthVisionConfig  `mapstructure:"vision"`
	TLS           StealthTLSConfig           `mapstructure:"tls"`
	Fragmentation StealthFragmentationConfig `mapstructure:"fragmentation"`
	AWG           StealthAWGConfig           `mapstructure:"awg"`
}

type StealthRealityConfig struct {
	Dest        string   `mapstructure:"dest"`
	ServerNames []string `mapstructure:"server_names"`
	PrivateKey  string   `mapstructure:"private_key"`
	PublicKey   string   `mapstructure:"public_key"`
	ShortIDs    []string `mapstructure:"short_ids"`
	Show        bool     `mapstructure:"show"`
	Xver        int      `mapstructure:"xver"`
}

type StealthXHTTPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
	Mode    string `mapstructure:"mode"`
	Tag     string `mapstructure:"tag"`
}

type StealthVisionConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Tag     string `mapstructure:"tag"`
}

type StealthTLSConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Port    int      `mapstructure:"port"`
	SNI     string   `mapstructure:"sni"`
	ALPN    []string `mapstructure:"alpn"`
	Tag     string   `mapstructure:"tag"`
}

type TelegramConfig struct {
	BotToken string  `mapstructure:"bot_token"`
	AdminIDs []int64 `mapstructure:"admin_ids"`
}

type LimitsConfig struct {
	DefaultTrafficGB  int64 `mapstructure:"default_traffic_gb"`
	DefaultExpireDays int   `mapstructure:"default_expire_days"`
}

// IsActive reports whether stealth presets are enabled and minimally configured.
func (s *StealthConfig) IsActive() bool {
	if s == nil || !s.Enabled {
		return false
	}
	if s.Reality.PrivateKey == "" || s.Reality.PublicKey == "" || s.Reality.Dest == "" {
		return false
	}
	if len(s.Reality.ServerNames) == 0 || len(s.Reality.ShortIDs) == 0 {
		return false
	}
	return s.XHTTP.Enabled || s.Vision.Enabled || s.TLS.Enabled || s.AWG.Enabled
}

// FingerprintOrDefault returns the configured uTLS fingerprint or firefox.
func (s *StealthConfig) FingerprintOrDefault() string {
	if s == nil || s.Fingerprint == "" {
		return "firefox"
	}
	return s.Fingerprint
}

// PrimarySNI returns the first Reality server name.
func (s *StealthConfig) PrimarySNI() string {
	if s == nil || len(s.Reality.ServerNames) == 0 {
		return ""
	}
	return s.Reality.ServerNames[0]
}

// PrimaryShortID returns the first Reality short ID.
func (s *StealthConfig) PrimaryShortID() string {
	if s == nil || len(s.Reality.ShortIDs) == 0 {
		return ""
	}
	return s.Reality.ShortIDs[0]
}

func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetDefault("server.port", 8080)
	v.SetDefault("core.type", "xray")
	v.SetDefault("core.listen_port", 443)
	v.SetDefault("core.public_host", "127.0.0.1")
	v.SetDefault("core.stats_poll_seconds", 60)
	v.SetDefault("limits.default_traffic_gb", 50)
	v.SetDefault("limits.default_expire_days", 30)
	v.SetDefault("server.client_socks5_port", 10808)

	v.SetDefault("core.stealth.fingerprint", "firefox")
	v.SetDefault("core.stealth.xhttp.enabled", true)
	v.SetDefault("core.stealth.xhttp.port", 443)
	v.SetDefault("core.stealth.xhttp.path", "/api/v1/data")
	v.SetDefault("core.stealth.xhttp.mode", "stream-one")
	v.SetDefault("core.stealth.xhttp.tag", "vless-xhttp-reality")
	v.SetDefault("core.stealth.vision.enabled", true)
	v.SetDefault("core.stealth.vision.port", 8443)
	v.SetDefault("core.stealth.vision.tag", "vless-vision-reality")
	v.SetDefault("core.stealth.tls.enabled", false)
	v.SetDefault("core.stealth.tls.port", 2053)
	v.SetDefault("core.stealth.tls.alpn", []string{"h2", "http/1.1"})
	v.SetDefault("core.stealth.tls.tag", "vless-tls")
	v.SetDefault("core.stealth.fragmentation.enabled", false)
	v.SetDefault("core.stealth.fragmentation.strategy", "serverhello")
	v.SetDefault("core.stealth.fragmentation.length", "50-100")
	v.SetDefault("core.stealth.fragmentation.delay", "10-20")
	v.SetDefault("core.stealth.fragmentation.max_split", "2-4")
	v.SetDefault("core.stealth.awg.enabled", false)
	v.SetDefault("core.stealth.awg.port", 51820)
	v.SetDefault("core.stealth.awg.subnet", "10.8.0.0/24")
	v.SetDefault("core.stealth.awg.server_address", "10.8.0.1/24")
	v.SetDefault("core.stealth.awg.config_path", "./data/awg/awg0.conf")
	v.SetDefault("core.stealth.awg.jc", 4)
	v.SetDefault("core.stealth.awg.jmin", 40)
	v.SetDefault("core.stealth.awg.jmax", 70)
	v.SetDefault("core.stealth.awg.s1", 84)
	v.SetDefault("core.stealth.awg.s2", 0)
	v.SetDefault("core.stealth.awg.h1", int64(1286472620))
	v.SetDefault("core.stealth.awg.h2", int64(2995958389))
	v.SetDefault("core.stealth.awg.h3", int64(641212874))
	v.SetDefault("core.stealth.awg.h4", int64(3523276991))
	v.SetDefault("core.stealth.awg.tag", "awg-reserve")
	v.SetDefault("core.multihop.enabled", false)
	v.SetDefault("core.multihop.local_role", "entry")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.Telegram.BotToken = token
	} else if strings.HasPrefix(cfg.Telegram.BotToken, "${") {
		cfg.Telegram.BotToken = os.ExpandEnv(cfg.Telegram.BotToken)
	}

	if cfg.Core.ListenPort == 0 {
		cfg.Core.ListenPort = 443
	}
	if cfg.Core.PublicHost == "" {
		cfg.Core.PublicHost = "127.0.0.1"
	}
	if cfg.Core.StatsPoll == 0 {
		cfg.Core.StatsPoll = 60
	}
	if cfg.Server.ClientSOCKS5Port == 0 {
		cfg.Server.ClientSOCKS5Port = 10808
	}

	if err := ValidateAPIKey(cfg.Server.APIKey); err != nil {
		return nil, err
	}

	if cfg.Server.EnableDocs == nil {
		t := true
		cfg.Server.EnableDocs = &t
	}
	if len(cfg.Server.CORSOrigins) == 0 {
		cfg.Server.CORSOrigins = []string{
			"http://localhost:8888",
			"http://127.0.0.1:8888",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}
	}
	if cfg.Server.SubscriptionDefaultTTLHours <= 0 {
		cfg.Server.SubscriptionDefaultTTLHours = 168
	}
	if cfg.Server.AutoDeactivateOnQuota == nil {
		t := true
		cfg.Server.AutoDeactivateOnQuota = &t
	}

	cfg.path = path
	return &cfg, nil
}

// Path returns the config file path used by Load.
func (c *Config) Path() string {
	if c == nil {
		return ""
	}
	return c.path
}

// SetPath sets the config file path (tests / programmatic use).
func (c *Config) SetPath(path string) {
	if c != nil {
		c.path = path
	}
}

// APIKey returns the current API key (safe for dynamic middleware).
func (c *Config) APIKey() string {
	if c == nil {
		return ""
	}
	return c.Server.APIKey
}

// RotateAPIKey validates, persists, and hot-swaps the API key in memory.
func (c *Config) RotateAPIKey(newKey string) error {
	if c == nil {
		return fmt.Errorf("nil config")
	}
	if err := ValidateAPIKey(newKey); err != nil {
		return err
	}
	if c.path != "" {
		if err := PersistAPIKey(c.path, newKey); err != nil {
			return err
		}
	}
	c.Server.APIKey = newKey
	return nil
}

// SubscriptionHMACSecret is used to sign subscription URLs.
func (c *Config) SubscriptionHMACSecret() string {
	if c == nil {
		return ""
	}
	if s := strings.TrimSpace(c.Server.SubscriptionSigSecret); s != "" {
		return s
	}
	return c.Server.APIKey
}

// AutoDeactivateOnQuotaEnabled reports whether quota exhaustion should flip active=false.
func (c *Config) AutoDeactivateOnQuotaEnabled() bool {
	return c != nil && c.Server.AutoDeactivateOnQuota != nil && *c.Server.AutoDeactivateOnQuota
}

// DocsEnabled reports whether OpenAPI/Swagger routes should be registered.
// Nil EnableDocs defaults to true (Load also sets this explicitly).
func (c *Config) DocsEnabled() bool {
	if c == nil {
		return false
	}
	if c.Server.EnableDocs == nil {
		return true
	}
	return *c.Server.EnableDocs
}

// insecureAPIKeys are placeholders that must never be used in a running panel.
var insecureAPIKeys = map[string]struct{}{
	"":                        {},
	"change-me":               {},
	"change-me-to-secure-key": {},
	"YOUR_API_KEY":            {},
}

// ValidateAPIKey rejects empty and well-known insecure placeholder keys.
// Test configs may set RIONEXGATE_ALLOW_INSECURE_API_KEY=1 to bypass.
func ValidateAPIKey(key string) error {
	if os.Getenv("RIONEXGATE_ALLOW_INSECURE_API_KEY") == "1" {
		if key == "" {
			return fmt.Errorf("server.api_key is empty")
		}
		return nil
	}
	if _, bad := insecureAPIKeys[key]; bad {
		return fmt.Errorf("server.api_key is missing or insecure; set a long random value in config.yaml (refusing placeholders like change-me-to-secure-key)")
	}
	if len(key) < 16 {
		return fmt.Errorf("server.api_key is too short (minimum 16 characters)")
	}
	return nil
}

func (c *Config) SetCoreType(t string) {
	c.Core.Type = t
}
