package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

func signSubscription(token string, exp int64, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(token))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(strconv.FormatInt(exp, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifySubscriptionSig(token, expStr, sig, secret string) error {
	if token == "" || expStr == "" || sig == "" || secret == "" {
		return fmt.Errorf("missing signature parameters")
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid exp")
	}
	if exp < time.Now().Unix() {
		return fmt.Errorf("subscription link expired")
	}
	expected := signSubscription(token, exp, secret)
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func (h *Handler) signedSubscriptionURL(baseURL, token string, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = time.Duration(h.cfg.Server.SubscriptionDefaultTTLHours) * time.Hour
	}
	expTime := time.Now().Add(ttl)
	exp := expTime.Unix()
	secret := h.cfg.SubscriptionHMACSecret()
	if secret == "" {
		return "", time.Time{}, fmt.Errorf("no signing secret configured")
	}
	sig := signSubscription(token, exp, secret)
	url := fmt.Sprintf("%s/api/subscription/%s?exp=%d&sig=%s", baseURL, token, exp, sig)
	return url, expTime, nil
}
