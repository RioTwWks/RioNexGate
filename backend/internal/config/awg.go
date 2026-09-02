package config

type StealthAWGConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Port          int    `mapstructure:"port"`
	Subnet        string `mapstructure:"subnet"`
	ServerAddress string `mapstructure:"server_address"`
	PrivateKey    string `mapstructure:"private_key"`
	PublicKey     string `mapstructure:"public_key"`
	ConfigPath    string `mapstructure:"config_path"`
	Jc            int    `mapstructure:"jc"`
	Jmin          int    `mapstructure:"jmin"`
	Jmax          int    `mapstructure:"jmax"`
	S1            int    `mapstructure:"s1"`
	S2            int    `mapstructure:"s2"`
	H1            int64  `mapstructure:"h1"`
	H2            int64  `mapstructure:"h2"`
	H3            int64  `mapstructure:"h3"`
	H4            int64  `mapstructure:"h4"`
	Tag           string `mapstructure:"tag"`
}

func (s *StealthConfig) AWGActive() bool {
	if s == nil || !s.AWG.Enabled {
		return false
	}
	return s.AWG.PrivateKey != "" && s.AWG.PublicKey != ""
}

func (a *StealthAWGConfig) PortOrDefault() int {
	if a == nil || a.Port == 0 { return 51820 }
	return a.Port
}
func (a *StealthAWGConfig) SubnetOrDefault() string {
	if a == nil || a.Subnet == "" { return "10.8.0.0/24" }
	return a.Subnet
}
func (a *StealthAWGConfig) ServerAddressOrDefault() string {
	if a == nil || a.ServerAddress == "" { return "10.8.0.1/24" }
	return a.ServerAddress
}
func (a *StealthAWGConfig) ConfigPathOrDefault() string {
	if a == nil || a.ConfigPath == "" { return "./data/awg/awg0.conf" }
	return a.ConfigPath
}
