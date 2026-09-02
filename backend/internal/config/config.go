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
	Port             int    `mapstructure:"port"`
	APIKey           string `mapstructure:"api_key"`
	PublicBaseURL    string `mapstructure:"public_base_url"`
	ClientSOCKS5Port int    `mapstructure:"client_socks5_port"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type CoreConfig struct {
	Type       string          `mapstructure:"type"`
	ListenPort int             `mapstructure:"listen_port"`
	PublicHost string          `mapstructure:"public_host"`
	Xray       XrayConfig      `mapstructure:"xray"`
	Singbox    SingboxConfig   `mapstructure:"singbox"`
	StatsPoll  int             `mapstructure:"stats_poll_seconds"`
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

type TelegramConfig struct {
	BotToken string  `mapstructure:"bot_token"`
	AdminIDs []int64 `mapstructure:"admin_ids"`
}

type LimitsConfig struct {
	DefaultTrafficGB  int64 `mapstructure:"default_traffic_gb"`
	DefaultExpireDays int   `mapstructure:"default_expire_days"`
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

	return &cfg, nil
}

func (c *Config) SetCoreType(t string) {
	c.Core.Type = t
}
