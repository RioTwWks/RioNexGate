package config_test

import (
	"testing"
	"rionexgate/internal/config"
)

func TestFragmentationDefaults(t *testing.T) {
	f := config.StealthFragmentationConfig{}
	if f.PacketsValue() != "tlshello" { t.Fatal("expected tlshello") }
}
func TestFragmentationApplicable(t *testing.T) {
	s := &config.StealthConfig{TLS: config.StealthTLSConfig{Enabled: true}, Fragmentation: config.StealthFragmentationConfig{Enabled: true}, XHTTP: config.StealthXHTTPConfig{Enabled: true}}
	if !s.FragmentationApplicable() { t.Fatal("expected applicable") }
}
