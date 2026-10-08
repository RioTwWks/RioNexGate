package api

import (
	"strconv"
	"testing"
	"time"
)

func TestSubscriptionSigRoundTrip(t *testing.T) {
	token := "abcd"
	secret := "signing-secret-key"
	exp := time.Now().Add(time.Hour).Unix()
	sig := signSubscription(token, exp, secret)
	if err := verifySubscriptionSig(token, strconv.FormatInt(exp, 10), sig, secret); err != nil {
		t.Fatal(err)
	}
	if err := verifySubscriptionSig(token, strconv.FormatInt(exp, 10), "deadbeef", secret); err == nil {
		t.Fatal("expected bad sig to fail")
	}
	past := time.Now().Add(-time.Hour).Unix()
	pastSig := signSubscription(token, past, secret)
	if err := verifySubscriptionSig(token, strconv.FormatInt(past, 10), pastSig, secret); err == nil {
		t.Fatal("expected expired to fail")
	}
}
