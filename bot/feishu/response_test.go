package feishu

import "testing"

func TestResponseMeta(t *testing.T) {
	if !(&ResponseMeta{}).Succeed() {
		t.Fatal("zero ResponseMeta should succeed")
	}
	got := (ResponseMeta{Code: 19001, Message: "bad token"}).String()
	if got != "code: 19001, msg: bad token" {
		t.Fatalf("String() = %q", got)
	}
}
