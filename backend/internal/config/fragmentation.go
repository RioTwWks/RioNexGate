package config

import "strings"

type StealthFragmentationConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Strategy string `mapstructure:"strategy"`
	Length   string `mapstructure:"length"`
	Delay    string `mapstructure:"delay"`
	MaxSplit string `mapstructure:"max_split"`
}

const FragmentationRealityLimitation = "Xray-core finalmask.fragment on REALITY inbounds crashes the process on the first connection (missing CloseWrite on fragmentConn, confirmed through v26.7.28). Fragmentation is emitted only on the optional VLESS+TLS inbound until upstream fixes this."

func (f *StealthFragmentationConfig) IsEnabled() bool { return f != nil && f.Enabled }
func (f *StealthFragmentationConfig) StrategyOrDefault() string {
	if f == nil || f.Strategy == "" || f.Strategy == "serverhello" { return "serverhello" }
	return f.Strategy
}
func (f *StealthFragmentationConfig) PacketsValue() string {
	if f != nil && f.Strategy == "all" { return "1-3" }
	return "tlshello"
}
func (f *StealthFragmentationConfig) LengthOrDefault() string {
	if f != nil && strings.TrimSpace(f.Length) != "" { return strings.TrimSpace(f.Length) }
	return "50-100"
}
func (f *StealthFragmentationConfig) DelayOrDefault() string {
	if f != nil && strings.TrimSpace(f.Delay) != "" { return strings.TrimSpace(f.Delay) }
	return "10-20"
}
func (f *StealthFragmentationConfig) MaxSplitOrDefault() string {
	if f != nil && strings.TrimSpace(f.MaxSplit) != "" { return strings.TrimSpace(f.MaxSplit) }
	return "2-4"
}
func (s *StealthConfig) FragmentationApplicable() bool {
	return s != nil && s.Fragmentation.IsEnabled() && s.TLS.Enabled
}
func (s *StealthConfig) FragmentationLimitation() string {
	if s == nil || !s.Fragmentation.IsEnabled() { return "" }
	if s.XHTTP.Enabled || s.Vision.Enabled { return FragmentationRealityLimitation }
	return ""
}
func (f *StealthFragmentationConfig) Aggressive() bool {
	if f == nil || !f.IsEnabled() { return false }
	if f.StrategyOrDefault() == "all" { return true }
	length := f.LengthOrDefault()
	return strings.HasPrefix(length, "1-") || length == "1"
}
