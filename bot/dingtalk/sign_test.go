// Copyright 2022-2024 The pmsg Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	timestamp, secret := "1700000000000", "secret"
	hash := hmac.New(sha256.New, []byte(secret))
	_, _ = hash.Write([]byte(timestamp + "\n" + secret))
	want := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	got, err := Sign(timestamp, secret)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Sign() = %q, want %q", got, want)
	}
}

func TestValidate(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	secret := "secret"
	signature, err := Sign(timestamp, secret)
	if err != nil {
		t.Fatal(err)
	}
	if valid, err := Validate(signature, timestamp, secret); err != nil || !valid {
		t.Fatalf("Validate() = %t, %v", valid, err)
	}
	if valid, err := Validate("bad", timestamp, secret); err != nil || valid {
		t.Fatalf("Validate(bad signature) = %t, %v", valid, err)
	}
	if _, err := Validate(signature, "invalid", secret); err == nil {
		t.Fatal("Validate accepted an invalid timestamp")
	}
	oldTimestamp := strconv.FormatInt(time.Now().Add(-2*time.Hour).UnixMilli(), 10)
	if _, err := Validate(signature, oldTimestamp, secret); err == nil {
		t.Fatal("Validate accepted an expired timestamp")
	}
}
