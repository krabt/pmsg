package dingtalk

import "testing"

func TestResponseMeta(t *testing.T) {
	if !(&ResponseMeta{}).Succeed() {
		t.Fatal("zero ResponseMeta should succeed")
	}
	got := (ResponseMeta{ErrorCode: 310000, ErrorMessage: "bad token"}).String()
	if got != "errcode: 310000, errmsg: bad token" {
		t.Fatalf("String() = %q", got)
	}
}
