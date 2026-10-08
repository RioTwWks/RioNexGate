package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"sync"
)

var apiKeyFileMu sync.Mutex

// GenerateAPIKey returns a 32-byte hex API key (64 chars).
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// PersistAPIKey updates server.api_key in the YAML file without rewriting the whole config.
func PersistAPIKey(path, newKey string) error {
	if path == "" {
		return fmt.Errorf("config path is empty")
	}
	apiKeyFileMu.Lock()
	defer apiKeyFileMu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(?m)^([ \t]*api_key:[ \t]*)(["']?)([^"'#\n]*?)(["']?)([ \t]*(?:#.*)?)?$`)
	loc := re.FindSubmatchIndex(data)
	if loc == nil {
		return fmt.Errorf("api_key field not found in %s", path)
	}
	// Rebuild the matched line: prefix + quoted new key + trailing comment.
	prefix := data[loc[2]:loc[3]]
	suffix := []byte{}
	if loc[10] >= 0 && loc[11] >= 0 {
		suffix = data[loc[10]:loc[11]]
	}
	replacement := append(append([]byte{}, prefix...), []byte(`"`+newKey+`"`)...)
	replacement = append(replacement, suffix...)
	out := append([]byte{}, data[:loc[0]]...)
	out = append(out, replacement...)
	out = append(out, data[loc[1]:]...)

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	_ = os.Chmod(path, 0o600)
	return nil
}
