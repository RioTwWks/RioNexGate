package config

import (
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
}

type ServerConfig struct {
	Port   int    `mapstructure:"port"`
	APIKey string `mapstructure:"api_key"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type CoreConfig struct {
	Type       string        `mapstructure:"type"`
	ListenPort int           `mapstructure:"listen_port"`
	PublicHost string        `mapstructure:"public_host"`
	Xray       XrayConfig    `mapstructure:"xray"`
	Singbox    SingboxConfig `mapstructure:"singbox"`
	Stealth    StealthConfig `mapstructure:"stealth"`
	StatsPoll  int           `mapstructure:"stats_poll_seconds"`
}

type XrayConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	BinaryPath string `mapstructure:"binary_path"`
	APIAddress string `mapstructure:"api_address"`
}

type SingboxConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	BinaryPath string `mapstructure:"binary_path"`
	APIAddress string `mapstructure:"api_address"`
}

// StealthConfig holds anti-DPI transport presets (Reality + XHTTP / Vision / TLS).
type StealthConfig struct {
	Enabled     bool                 `mapstructure:"enabled"`
	Fingerprint string               `mapstructure:"fingerprint"`
	Reality     StealthRealityConfig `mapstructure:"reality"`
	XHTTP       StealthXHTTPConfig   `mapstructure:"xhttp"`
	Vision      StealthVisionConfig  `mapstructure:"vision"`
	TLS         StealthTLSConfig     `mapstructure:"tls"`
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
	return s.XHTTP.Enabled || s.Vision.Enabled || s.TLS.Enabled
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

	return &cfg, nil
}

func (c *Config) SetCoreType(t string) {
	c.Core.Type = t
}
