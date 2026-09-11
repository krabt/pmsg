package feishu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	timestamp, secret := "1700000000", "secret"
	hash := hmac.New(sha256.New, []byte(timestamp+"\n"+secret))
	want := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	if got := Sign(timestamp, secret); got != want {
		t.Fatalf("Sign() = %q, want %q", got, want)
	}
}

func TestValidate(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	secret := "secret"
	signature := Sign(timestamp, secret)
	if valid, err := Validate(signature, timestamp, secret); err != nil || !valid {
		t.Fatalf("Validate() = %t, %v", valid, err)
	}
	if valid, err := Validate("bad", timestamp, secret); err != nil || valid {
		t.Fatalf("Validate(bad signature) = %t, %v", valid, err)
	}
	if _, err := Validate(signature, "invalid", secret); err == nil {
		t.Fatal("Validate accepted an invalid timestamp")
	}
	oldTimestamp := strconv.FormatInt(time.Now().Add(-2*time.Hour).Unix(), 10)
	if _, err := Validate(signature, oldTimestamp, secret); err == nil {
		t.Fatal("Validate accepted an expired timestamp")
	}
}
