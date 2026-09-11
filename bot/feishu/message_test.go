package feishu

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateMsgType(t *testing.T) {
	for _, value := range []string{MsgTypeText, MsgTypePost, MsgTypeImage, MsgTypeShareChat, MsgTypeInteractive} {
		if err := ValidateMsgType(value); err != nil {
			t.Errorf("ValidateMsgType(%q) returned error: %v", value, err)
		}
	}
	if err := ValidateMsgType("unknown"); err == nil {
		t.Fatal("ValidateMsgType accepted an unknown type")
	}
}

func TestMessageJSONOmitsUnsetContent(t *testing.T) {
	data, err := json.Marshal(Message{MsgType: MsgTypeText, Content: &ContentMeta{Text: "hello"}})
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, `"msg_type":"text"`) || !strings.Contains(encoded, `"content":{"text":"hello"}`) {
		t.Fatalf("message JSON = %s", encoded)
	}
	if strings.Contains(encoded, "card") {
		t.Fatalf("message JSON contains an unset field: %s", encoded)
	}
}
